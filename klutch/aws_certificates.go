package klutch

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmTypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	elbv2 "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const (
	acmIssueTimeout   = 20 * time.Minute
	acmPollInterval   = 15 * time.Second
	route53PollDelay  = 5 * time.Second
	route53MaxRetries = 60
)

const (
	klutchTagKey   = "Klutch"
	klutchTagValue = "ControlPlane"
)

type CertificateProvisioner interface {
	EnsureCertificate(domainName string, altNames []string, hostedZoneName string) (string, error)
	EnsureCNAMERecords(hostedZoneName string, records map[string]string) error
	GetHostedZoneNS(hostedZoneName string) ([]string, error)
	EnsureALBAliasRecord(hostedZoneName, recordName, albDNSName string) error
}

type AWSProvisioner struct {
	acmClient *acm.Client
	r53Client *route53.Client
	elbv2     *elbv2.Client
	awsRegion string
}

// NewCertificateProvisioner initializes ACM and Route53 clients using the default AWS configuration.
func NewCertificateProvisioner(kubeContext string) CertificateProvisioner {
	region := detectAWSRegion(kubeContext)

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to load AWS configuration.")
	}

	return &AWSProvisioner{
		acmClient: acm.NewFromConfig(cfg),
		r53Client: route53.NewFromConfig(cfg),
		elbv2:     elbv2.NewFromConfig(cfg),
		awsRegion: region,
	}
}

// detectAWSRegion extracts the AWS region from the EKS API endpoint in kubeconfig.
func detectAWSRegion(kubeContext string) string {
	clusterURL := getClusterURLFromKubeconfig(kubeContext)
	host := clusterURL.Hostname()
	parts := strings.Split(host, ".")

	// EKS endpoints look like: XYZ.eu-central-1.eks.amazonaws.com
	if len(parts) >= 4 && parts[len(parts)-3] == "eks" && parts[len(parts)-2] == "amazonaws" {
		region := parts[len(parts)-4]
		if region != "" {
			return region
		}
	}

	makeup.ExitDueToFatalError(nil, fmt.Sprintf("Could not determine AWS region from cluster endpoint %s (expected EKS-style hostname).", host))
	return ""
}

// verifyHostedZoneResolvable ensures the hosted zone name is publicly resolvable (delegated).
// If it is not resolvable, ACM DNS validation will never succeed.
func verifyHostedZoneResolvable(hostedZoneName string) {
	lookupName := hostedZoneName
	if !strings.HasSuffix(lookupName, ".") {
		lookupName = lookupName + "."
	}

	nsRecords, err := net.LookupNS(lookupName)
	if err != nil || len(nsRecords) == 0 {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Hosted zone %s is not publicly resolvable. Ensure it is delegated before requesting an ACM certificate.", hostedZoneName))
	}

	nsList := make([]string, 0, len(nsRecords))
	for _, ns := range nsRecords {
		nsList = append(nsList, ns.Host)
	}
	makeup.PrintInfo(fmt.Sprintf("Hosted zone %s is publicly delegated to: %s", hostedZoneName, strings.Join(nsList, ", ")))
}

func normalizeDomain(name string) string {
	return strings.TrimSuffix(strings.ToLower(name), ".")
}

func ensureTrailingDot(name string) string {
	if name == "" {
		return name
	}
	if strings.HasSuffix(name, ".") {
		return name
	}
	return name + "."
}

func ensureDualstackALBDNS(albDNS string) string {
	if albDNS == "" {
		return albDNS
	}
	if strings.HasPrefix(albDNS, "dualstack.") {
		return albDNS
	}
	return "dualstack." + albDNS
}

func certificateTags(domainName string) []acmTypes.Tag {
	sanitized := strings.ReplaceAll(strings.TrimPrefix(normalizeDomain(domainName), "*."), "*", "")
	if sanitized == "" {
		sanitized = "klutch-certificate"
	}
	return []acmTypes.Tag{
		{Key: aws.String(klutchTagKey), Value: aws.String(klutchTagValue)},
		{Key: aws.String("Name"), Value: aws.String(fmt.Sprintf("klutch-certificate-%s", sanitized))},
	}
}

// findExistingIssuedCertificate returns an issued certificate ARN covering all required names, if any.
func (p *AWSProvisioner) findExistingIssuedCertificate(ctx context.Context, requiredNames []string) (string, error) {
	if len(requiredNames) == 0 {
		return "", nil
	}

	required := make(map[string]struct{}, len(requiredNames))
	for _, n := range requiredNames {
		if n == "" {
			continue
		}
		required[normalizeDomain(n)] = struct{}{}
	}

	paginator := acm.NewListCertificatesPaginator(p.acmClient, &acm.ListCertificatesInput{
		CertificateStatuses: []acmTypes.CertificateStatus{acmTypes.CertificateStatusIssued},
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("listing ACM certificates: %w", err)
		}

		for _, summary := range page.CertificateSummaryList {
			arn := aws.ToString(summary.CertificateArn)
			if arn == "" {
				continue
			}

			desc, err := p.acmClient.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
				CertificateArn: aws.String(arn),
			})
			if err != nil {
				return "", fmt.Errorf("describing certificate %s: %w", arn, err)
			}

			present := make(map[string]struct{})
			present[normalizeDomain(aws.ToString(desc.Certificate.DomainName))] = struct{}{}
			for _, san := range desc.Certificate.SubjectAlternativeNames {
				present[normalizeDomain(san)] = struct{}{}
			}

			allCovered := true
			for rn := range required {
				if _, ok := present[rn]; !ok {
					allCovered = false
					break
				}
			}

			if allCovered {
				makeup.PrintInfo(fmt.Sprintf("Found existing issued ACM certificate covering %v: %s", requiredNames, arn))
				return arn, nil
			}
		}
	}

	return "", nil
}

// ensureCertificateTags makes sure our standard tags are present on the given certificate.
func (p *AWSProvisioner) ensureCertificateTags(ctx context.Context, certArn, domainName string) error {
	existingTagsOut, err := p.acmClient.ListTagsForCertificate(ctx, &acm.ListTagsForCertificateInput{
		CertificateArn: aws.String(certArn),
	})
	if err != nil {
		return fmt.Errorf("listing tags for certificate %s: %w", certArn, err)
	}

	existing := make(map[string]string, len(existingTagsOut.Tags))
	for _, t := range existingTagsOut.Tags {
		existing[aws.ToString(t.Key)] = aws.ToString(t.Value)
	}

	desired := certificateTags(domainName)
	var toAdd []acmTypes.Tag
	for _, t := range desired {
		if val, ok := existing[aws.ToString(t.Key)]; !ok || val != aws.ToString(t.Value) {
			toAdd = append(toAdd, t)
		}
	}

	if len(toAdd) == 0 {
		return nil
	}

	_, err = p.acmClient.AddTagsToCertificate(ctx, &acm.AddTagsToCertificateInput{
		CertificateArn: aws.String(certArn),
		Tags:           toAdd,
	})
	if err != nil {
		return fmt.Errorf("adding tags to certificate %s: %w", certArn, err)
	}

	makeup.PrintInfo(fmt.Sprintf("Ensured standard tags on certificate %s", certArn))
	return nil
}

// EnsureCertificate requests a public ACM certificate for domainName (+ altNames), creates the DNS validation record in the given hosted zone,
// waits for DNS to propagate, and waits for the certificate to be issued. Returns the certificate ARN.
func (p *AWSProvisioner) EnsureCertificate(domainName string, altNames []string, hostedZoneName string) (string, error) {
	ctx := context.Background()

	requiredNames := []string{domainName}
	requiredNames = append(requiredNames, altNames...)

	if existingArn, err := p.findExistingIssuedCertificate(ctx, requiredNames); err == nil && existingArn != "" {
		if err := p.ensureCertificateTags(ctx, existingArn, domainName); err != nil {
			return "", err
		}
		return existingArn, nil
	} else if err != nil {
		return "", err
	}

	reqOut, err := p.acmClient.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName:              aws.String(domainName),
		SubjectAlternativeNames: altNames,
		ValidationMethod:        acmTypes.ValidationMethodDns,
		Tags:                    certificateTags(domainName),
	})
	if err != nil {
		return "", fmt.Errorf("requesting ACM certificate: %w", err)
	}

	certArn := aws.ToString(reqOut.CertificateArn)
	if certArn == "" {
		return "", fmt.Errorf("ACM did not return a certificate ARN")
	}

	validationRecord, err := p.waitForDNSValidationRecord(ctx, certArn)
	if err != nil {
		return "", err
	}

	zoneID, err := p.getHostedZoneID(ctx, hostedZoneName)
	if err != nil {
		return "", err
	}

	if err := p.upsertValidationRecord(ctx, zoneID, validationRecord); err != nil {
		return "", err
	}

	if err := p.waitForCertificateIssued(ctx, certArn); err != nil {
		return "", err
	}

	return certArn, nil
}

func (p *AWSProvisioner) waitForDNSValidationRecord(ctx context.Context, certArn string) (*acmTypes.ResourceRecord, error) {
	deadline := time.After(acmIssueTimeout)
	ticker := time.NewTicker(acmPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return nil, fmt.Errorf("timed out waiting for DNS validation options from ACM for %s", certArn)
		case <-ticker.C:
			desc, err := p.acmClient.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
				CertificateArn: aws.String(certArn),
			})
			if err != nil {
				return nil, fmt.Errorf("describing certificate: %w", err)
			}

			for _, detail := range desc.Certificate.DomainValidationOptions {
				if detail.ResourceRecord != nil && detail.ValidationStatus == acmTypes.DomainStatusPendingValidation {
					return detail.ResourceRecord, nil
				}
			}
		}
	}
}

func (p *AWSProvisioner) getHostedZoneID(ctx context.Context, hostedZoneName string) (string, error) {
	normalized := hostedZoneName
	if !strings.HasSuffix(normalized, ".") {
		normalized = normalized + "."
	}

	var (
		publicZoneID  string
		privateZoneID string
	)

	out, err := p.r53Client.ListHostedZonesByName(ctx, &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(normalized),
	})
	if err != nil {
		return "", fmt.Errorf("listing hosted zones: %w", err)
	}

	for _, zone := range out.HostedZones {
		if aws.ToString(zone.Name) == normalized {
			if zone.Config != nil && zone.Config.PrivateZone {
				privateZoneID = strings.TrimPrefix(aws.ToString(zone.Id), "/hostedzone/")
				continue
			}

			publicZoneID = strings.TrimPrefix(aws.ToString(zone.Id), "/hostedzone/")
			break
		}
	}

	if publicZoneID != "" {
		makeup.PrintInfo(fmt.Sprintf("Using public hosted zone %s (%s) for certificate validation.", hostedZoneName, publicZoneID))
		return publicZoneID, nil
	}

	if privateZoneID != "" {
		makeup.PrintWarning(fmt.Sprintf("Public hosted zone %s not found; falling back to private hosted zone (%s). ACM validation may fail if the zone isn't publicly resolvable.", hostedZoneName, privateZoneID))
		return privateZoneID, nil
	}

	return "", fmt.Errorf("hosted zone %s not found", hostedZoneName)
}

func (p *AWSProvisioner) upsertValidationRecord(ctx context.Context, zoneID string, record *acmTypes.ResourceRecord) error {
	change := &types.Change{
		Action: types.ChangeActionUpsert,
		ResourceRecordSet: &types.ResourceRecordSet{
			Name: aws.String(aws.ToString(record.Name)),
			Type: types.RRTypeCname,
			TTL:  aws.Int64(300),
			ResourceRecords: []types.ResourceRecord{
				{Value: aws.String(aws.ToString(record.Value))},
			},
		},
	}

	changeResp, err := p.r53Client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{*change},
		},
	})
	if err != nil {
		return fmt.Errorf("creating DNS validation record: %w", err)
	}

	changeID := aws.ToString(changeResp.ChangeInfo.Id)
	for i := 0; i < route53MaxRetries; i++ {
		time.Sleep(route53PollDelay)
		statusOut, err := p.r53Client.GetChange(ctx, &route53.GetChangeInput{Id: aws.String(changeID)})
		if err != nil {
			return fmt.Errorf("checking route53 change status: %w", err)
		}
		if statusOut.ChangeInfo.Status == types.ChangeStatusInsync {
			return nil
		}
	}

	return fmt.Errorf("route53 change %s did not become INSYNC in time", changeID)
}

func (p *AWSProvisioner) waitForCertificateIssued(ctx context.Context, certArn string) error {
	deadline := time.After(acmIssueTimeout)
	ticker := time.NewTicker(acmPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-deadline:
			return fmt.Errorf("timed out waiting for ACM certificate %s to be issued", certArn)
		case <-ticker.C:
			desc, err := p.acmClient.DescribeCertificate(ctx, &acm.DescribeCertificateInput{
				CertificateArn: aws.String(certArn),
			})
			if err != nil {
				return fmt.Errorf("describing certificate: %w", err)
			}

			status := desc.Certificate.Status
			if status == acmTypes.CertificateStatusIssued {
				return nil
			}
			if status == acmTypes.CertificateStatusFailed {
				return fmt.Errorf("ACM certificate %s failed issuance", certArn)
			}
		}
	}
}

// EnsureCNAMERecords creates or updates CNAME records (name -> target) in the given hosted zone.
func (p *AWSProvisioner) EnsureCNAMERecords(hostedZoneName string, records map[string]string) error {
	if len(records) == 0 {
		return nil
	}

	ctx := context.Background()
	zoneID, err := p.getHostedZoneID(ctx, hostedZoneName)
	if err != nil {
		return err
	}

	var changes []types.Change
	for name, target := range records {
		if name == "" || target == "" {
			continue
		}
		changes = append(changes, types.Change{
			Action: types.ChangeActionUpsert,
			ResourceRecordSet: &types.ResourceRecordSet{
				Name: aws.String(ensureTrailingDot(name)),
				Type: types.RRTypeCname,
				TTL:  aws.Int64(60),
				ResourceRecords: []types.ResourceRecord{
					{Value: aws.String(ensureTrailingDot(target))},
				},
			},
		})
	}

	if len(changes) == 0 {
		return nil
	}

	changeResp, err := p.r53Client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: changes,
		},
	})
	if err != nil {
		return fmt.Errorf("creating CNAME records: %w", err)
	}

	changeID := aws.ToString(changeResp.ChangeInfo.Id)
	for i := 0; i < route53MaxRetries; i++ {
		time.Sleep(route53PollDelay)
		statusOut, err := p.r53Client.GetChange(ctx, &route53.GetChangeInput{Id: aws.String(changeID)})
		if err != nil {
			return fmt.Errorf("checking route53 change status: %w", err)
		}
		if statusOut.ChangeInfo.Status == types.ChangeStatusInsync {
			return nil
		}
	}

	return fmt.Errorf("route53 change %s did not become INSYNC in time", changeID)
}

// EnsureALBAliasRecord creates or updates ALIAS A/AAAA records pointing to the given ALB DNS name.
func (p *AWSProvisioner) EnsureALBAliasRecord(hostedZoneName, recordName, albDNSName string) error {
	if hostedZoneName == "" || recordName == "" || albDNSName == "" {
		return fmt.Errorf("hostedZoneName, recordName, and albDNSName are required for alias creation")
	}

	ctx := context.Background()
	zoneID, err := p.getHostedZoneID(ctx, hostedZoneName)
	if err != nil {
		return err
	}

	parts := strings.Split(albDNSName, ".")
	if len(parts) == 0 {
		return fmt.Errorf("could not parse ALB DNS name %s", albDNSName)
	}
	lbName := parts[0]

	desc, err := p.elbv2.DescribeLoadBalancers(ctx, &elbv2.DescribeLoadBalancersInput{
		Names: []string{lbName},
	})
	if err != nil {
		return fmt.Errorf("describing ALB %s: %w", lbName, err)
	}
	if len(desc.LoadBalancers) == 0 {
		return fmt.Errorf("ALB %s not found", lbName)
	}

	canonicalHZ := aws.ToString(desc.LoadBalancers[0].CanonicalHostedZoneId)
	targetDNS := ensureDualstackALBDNS(albDNSName)

	aliasTarget := &types.AliasTarget{
		DNSName:              aws.String(ensureTrailingDot(targetDNS)),
		HostedZoneId:         aws.String(canonicalHZ),
		EvaluateTargetHealth: false,
	}

	changes := []types.Change{
		{
			Action: types.ChangeActionUpsert,
			ResourceRecordSet: &types.ResourceRecordSet{
				Name:        aws.String(ensureTrailingDot(recordName)),
				Type:        types.RRTypeA,
				AliasTarget: aliasTarget,
			},
		},
		{
			Action: types.ChangeActionUpsert,
			ResourceRecordSet: &types.ResourceRecordSet{
				Name:        aws.String(ensureTrailingDot(recordName)),
				Type:        types.RRTypeAaaa,
				AliasTarget: aliasTarget,
			},
		},
	}

	changeResp, err := p.r53Client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneID),
		ChangeBatch: &types.ChangeBatch{
			Changes: changes,
		},
	})
	if err != nil {
		return fmt.Errorf("creating ALIAS records for %s -> %s: %w", recordName, albDNSName, err)
	}

	changeID := aws.ToString(changeResp.ChangeInfo.Id)
	for i := 0; i < route53MaxRetries; i++ {
		time.Sleep(route53PollDelay)
		statusOut, err := p.r53Client.GetChange(ctx, &route53.GetChangeInput{Id: aws.String(changeID)})
		if err != nil {
			return fmt.Errorf("checking route53 change status: %w", err)
		}
		if statusOut.ChangeInfo.Status == types.ChangeStatusInsync {
			return nil
		}
	}

	return fmt.Errorf("route53 change %s did not become INSYNC in time", changeID)
}

// GetHostedZoneNS returns the nameservers configured for the hosted zone.
func (p *AWSProvisioner) GetHostedZoneNS(hostedZoneName string) ([]string, error) {
	ctx := context.Background()
	zoneID, err := p.getHostedZoneID(ctx, hostedZoneName)
	if err != nil {
		return nil, err
	}

	out, err := p.r53Client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: aws.String(zoneID),
	})
	if err != nil {
		return nil, fmt.Errorf("getting hosted zone %s: %w", hostedZoneName, err)
	}

	return out.DelegationSet.NameServers, nil
}
