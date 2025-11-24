package klutch

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	acmTypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const (
	acmIssueTimeout   = 20 * time.Minute
	acmPollInterval   = 15 * time.Second
	route53PollDelay  = 5 * time.Second
	route53MaxRetries = 60
)

type CertificateProvisioner interface {
	EnsureCertificate(domainName, hostedZoneName string) (string, error)
}

type AWSProvisioner struct {
	acmClient *acm.Client
	r53Client *route53.Client
}

// NewCertificateProvisioner initializes ACM and Route53 clients using the default AWS configuration.
func NewCertificateProvisioner() CertificateProvisioner {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if region == "" {
		makeup.ExitDueToFatalError(fmt.Errorf("AWS_REGION or AWS_DEFAULT_REGION must be set"), "Could not determine AWS region for ACM certificate request.")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		makeup.ExitDueToFatalError(err, "Failed to load AWS configuration.")
	}

	return &AWSProvisioner{
		acmClient: acm.NewFromConfig(cfg),
		r53Client: route53.NewFromConfig(cfg),
	}
}

// EnsureCertificate requests a public ACM certificate for domainName, creates the DNS validation record in the given hosted zone,
// waits for DNS to propagate, and waits for the certificate to be issued. Returns the certificate ARN.
func (p *AWSProvisioner) EnsureCertificate(domainName, hostedZoneName string) (string, error) {
	ctx := context.Background()

	reqOut, err := p.acmClient.RequestCertificate(ctx, &acm.RequestCertificateInput{
		DomainName:       aws.String(domainName),
		ValidationMethod: acmTypes.ValidationMethodDns,
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

	out, err := p.r53Client.ListHostedZonesByName(ctx, &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(normalized),
	})
	if err != nil {
		return "", fmt.Errorf("listing hosted zones: %w", err)
	}

	for _, zone := range out.HostedZones {
		if aws.ToString(zone.Name) == normalized {
			return strings.TrimPrefix(aws.ToString(zone.Id), "/hostedzone/"), nil
		}
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
