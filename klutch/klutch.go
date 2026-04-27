package klutch

import (
	"context"
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	klutchaws "github.com/anynines/a9s-cli-v2/klutch/aws"
	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
	"github.com/anynines/klutchio/bind/deploy/crd"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
	sigyaml "sigs.k8s.io/yaml"
)

const (
	demoTitle                       = "Klutch Demo"
	applyControlPlaneTitle          = "Applying Klutch Control Plane to the current Kubernetes cluster"
	controlPlaneClusterInfoFilePath = "klutch"
	controlPlaneClusterInfoFileName = "control-plane-cluster-info.yaml"
	controlPlaneClusterName         = "klutch-control-plane"
	appClusterName                  = "klutch-app"
	contextControlPlane             = "kind-" + controlPlaneClusterName
	contextApp                      = "kind-" + appClusterName
)

var (
	PortFlag        int = 8080
	publicResolvers     = []string{"8.8.8.8:53", "1.1.1.1:53"}
)

// ControlPlaneClusterInfo contains information about the created Control Plane Cluster.
type ControlPlaneClusterInfo struct {
	Host        string `yaml:"host"`
	IngressPort string `yaml:"ingressPort"`
}

type KlutchManager struct {
	// cpContext is the kube context for the Klutch Control Plane Cluster.
	cpContext string
	// cpK8s is the Kubernetes client for the Klutch Control Plane Cluster.
	cpK8s *k8s.KubeClient

	// appContext is the kube context for the Klutch App Cluster.
	appContext string
	// appK8s is the Kubernetes client for the Klutch App Cluster.
	appK8s *k8s.KubeClient
}

func NewKlutchManager() *KlutchManager {
	return &KlutchManager{
		cpContext:  contextControlPlane,
		cpK8s:      k8s.NewKubeClient(contextControlPlane),
		appContext: contextApp,
		appK8s:     k8s.NewKubeClient(contextApp),
	}
}

func NewKlutchManagerWithContexts(controlPlaneContext, appContext string) *KlutchManager {
	return &KlutchManager{
		cpContext:  controlPlaneContext,
		cpK8s:      k8s.NewKubeClient(controlPlaneContext),
		appContext: appContext,
		appK8s:     k8s.NewKubeClient(appContext),
	}
}

// DeployKlutchClusters deploys the Control Plane Cluster with all Klutch components, and a
// app cluster.
func DeployKlutchClusters() {
	makeup.PrintWelcomeScreen(
		makeup.UnattendedMode,
		demoTitle,
		"Let's deploy a Klutch setup with Kind...")

	// Makes sure the "WorkingDir" variable is set. This allows us to re-use existing code to deploy
	// the a8s stack, and write information about the created cluster to a file in the user's
	// configured working dir.
	demo.EstablishConfig()

	checkDeployPrerequisites()

	klutch := NewKlutchManager()
	klutch.deployControlPlaneCluster()
	klutch.deployAppCluster()
	printSummary()
}

// ApplyKlutchControlPlane installs the Klutch control plane components into the current kube context.
func ApplyKlutchControlPlane(host string, ingressPort int, acmCertificateARN string, hostedZoneName string, clusterName string) {
	makeup.PrintWelcomeScreen(
		makeup.UnattendedMode,
		applyControlPlaneTitle,
		"Let's install the Klutch control plane into your current Kubernetes cluster...")

	demo.EstablishConfig()

	checkControlPlaneInstallPrerequisites()

	if ingressPort < 1 || ingressPort > 65535 {
		makeup.ExitDueToFatalError(nil, "Invalid ingress port. Must be between 1 and 65535.")
	}

	baseDomain := strings.TrimSuffix(host, ".")
	if baseDomain == "" && hostedZoneName != "" {
		baseDomain = strings.TrimSuffix(hostedZoneName, ".")
	}

	dexHost := ""
	if baseDomain != "" {
		if strings.HasPrefix(baseDomain, "dex.") {
			dexHost = baseDomain
		} else {
			dexHost = fmt.Sprintf("dex.%s", baseDomain)
		}
	}
	backendHost := ""
	if baseDomain != "" {
		backendHost = fmt.Sprintf("klutch-bind.%s", baseDomain)
	}

	if controlPlaneOIDCOptions.Provider == "" {
		controlPlaneOIDCOptions.Provider = defaultOIDCProvider(demo.KubernetesTool)
	}
	controlPlaneOIDCOptions = controlPlaneOIDCOptions.normalize()
	useDex := controlPlaneOIDCOptions.Provider != OIDCProviderCognito
	if !useDex {
		dexHost = ""
	}
	makeup.PrintInfo(fmt.Sprintf("Using OIDC provider: %s", controlPlaneOIDCOptions.Provider))

	var provisioner CertificateProvisioner
	if hostedZoneName != "" {
		provisioner = NewCertificateProvisioner("")
		verifyHostedZoneResolvable(provisioner, hostedZoneName, clusterName)
		verifyHostedZoneRequirements(provisioner, hostedZoneName, baseDomain, dexHost, backendHost, clusterName)
	}

	// Auto-provision an ACM certificate if none was provided and a hosted zone is available.
	if acmCertificateARN == "" && hostedZoneName != "" {
		if baseDomain == "" {
			makeup.ExitDueToFatalError(nil, "A host or hosted zone is required to request an ACM certificate.")
		}

		primary := fmt.Sprintf("*.%s", baseDomain)
		altNames := []string{baseDomain}
		if dexHost != "" && dexHost != baseDomain {
			altNames = append(altNames, dexHost)
		}

		makeup.PrintInfo(fmt.Sprintf("Planned action: request or reuse ACM certificate for %s (SANs: %v) in hosted zone %s with DNS validation and tagging.", primary, altNames, hostedZoneName))
		makeup.WaitForUser()

		makeup.PrintInfo(fmt.Sprintf("Requesting ACM certificate for %s (SANs: %v) in hosted zone %s.", primary, altNames, hostedZoneName))
		arn, err := provisioner.EnsureCertificate(primary, altNames, hostedZoneName)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to request ACM certificate.")
		}
		acmCertificateARN = arn
		makeup.PrintInfo(fmt.Sprintf("Using ACM certificate ARN: %s", acmCertificateARN))
	}

	cpCtx, err := k8s.CurrentContext()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Can't retrieve the currently selected cluster:\n"+cpCtx)
	}

	manager := NewKlutchManagerWithContexts(cpCtx, "")
	manager.applyControlPlaneToContext(baseDomain, dexHost, backendHost, hostedZoneName, provisioner, strconv.Itoa(ingressPort), acmCertificateARN)
	printControlPlaneSummary(demo.DemoConfig.WorkingDir)
}

func (k *KlutchManager) deployControlPlaneCluster() {
	hostIP, err := determineHostLocalIP()
	if err != nil {
		makeup.ExitDueToFatalError(err, "Couldn't obtain the host's local IP address. Aborting...")
	}

	makeup.PrintInfo(fmt.Sprintf("Using IP address `%s` for the Control Plane Cluster.", hostIP))

	port := strconv.Itoa(PortFlag)
	DeployControlPlaneKindCluster(controlPlaneClusterName, hostIP, port)
	WaitForKindCluster(k.cpK8s)
	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, hostIP, port)

	ingressClass := "nginx"

	k.DeployIngressNginx()
	k.WaitForIngressNginx()

	scheme := determineIngressScheme(ingressClass, false)

	k.DeployDex(hostIP, port, ingressClass, scheme, "")
	k.WaitForDex()

	k.DeployBindBackend(hostIP, port, ingressClass, scheme, "", false, true)
	k.WaitForBindBackend(hostIP, port, scheme)

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser()
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
}

func (k *KlutchManager) deployAppCluster() {

	DeployAppCluster(appClusterName)
	WaitForKindCluster(k.appK8s)

	out, err := k8s.SwitchContext(k.appContext) // in case the app cluster already existed, switch to it.
	if err != nil {
		makeup.ExitDueToFatalError(err, "Could not switch to context "+k.appContext+":\n"+string(out))
	}
}

func printSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Deployed a Klutch Control Plane Cluster with Kind.")
	makeup.PrintCheckmark("Configured OIDC and deployed the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to App Clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark("Deployed an App Cluster.")
	makeup.PrintSuccessSummary("You are now ready to bind APIs from the App Cluster using the `a9s klutch bind` command.")
}

func (k *KlutchManager) applyControlPlaneToContext(baseDomain string, dexHost string, backendHost string, hostedZoneName string, provisioner CertificateProvisioner, ingressPort string, acmCertificateARN string) {
	ctx := context.Background()
	ingressClass := detectIngressClass(k.cpK8s)
	tlsEnabled := acmCertificateARN != "" && ingressClass == "alb"

	// Force port 443 when TLS is enabled; fall back to 80 for ALB without TLS.
	if tlsEnabled {
		if ingressPort != "443" {
			makeup.PrintInfo("ACM certificate provided; enabling TLS on ALB and using port 443.")
		}
		ingressPort = "443"
	} else if ingressClass == "alb" && ingressPort == "443" {
		makeup.PrintInfo("Detected ALB ingress and no TLS configuration. Switching ingress port to 80.")
		ingressPort = "80"
	}

	scheme := determineIngressScheme(ingressClass, tlsEnabled)

	useDex := controlPlaneOIDCOptions.Provider != OIDCProviderCognito
	zone := strings.TrimSuffix(hostedZoneName, ".")

	publicHost := ""
	if useDex {
		publicHost = dexHost
	}
	if publicHost == "" {
		publicHost = backendHost
	}
	if publicHost == "" {
		publicHost = baseDomain
	}
	if publicHost == "" {
		publicHost = getClusterExternalHost("")
		makeup.PrintInfo(fmt.Sprintf("No host provided. Using provisional host `%s` to bootstrap deployment.", publicHost))
	} else {
		makeup.PrintInfo(fmt.Sprintf("Using host `%s` for OIDC configuration.", publicHost))
	}

	if ingressClass == "nginx" {
		k.DeployIngressNginx()
		k.WaitForIngressNginx()
	} else {
		makeup.PrintInfo(fmt.Sprintf("Ingress class `%s` detected. Skipping ingress-nginx installation.", ingressClass))
	}

	waitHost := backendHost
	if waitHost == "" {
		waitHost = publicHost
	}

	resolvedOIDC := effectiveOIDCOptions(demo.KubernetesTool, scheme, waitHost)
	if resolvedOIDC.Provider == OIDCProviderCognito && (resolvedOIDC.IssuerURL == "" || resolvedOIDC.ClientID == "" || resolvedOIDC.ClientSecret == "") {
		region := defaultAWSRegion()
		prefix := sanitizeCognitoPrefix(baseDomain)
		if prefix == "" {
			prefix = sanitizeCognitoPrefix(demo.DemoClusterName)
		}
		if prefix == "" {
			prefix = "klutch"
		}
		makeup.PrintInfo(fmt.Sprintf("No Cognito settings provided. Provisioning Cognito (region: %s, prefix: %s)...", region, prefix))
		tenantUUID := uuid.New().String()
		oidcConn, err := klutchaws.EnsureCognitoOIDC(context.Background(), region, prefix, "", tenantUUID)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to provision Cognito for Klutch OIDC.")
		}
		if resolvedOIDC.IssuerURL == "" {
			resolvedOIDC.IssuerURL = oidcConn.IssuerURL
		}
		if resolvedOIDC.ClientID == "" {
			resolvedOIDC.ClientID = oidcConn.ClientID
		}
		if resolvedOIDC.ClientSecret == "" {
			resolvedOIDC.ClientSecret = oidcConn.ClientSecret
		}
		if resolvedOIDC.CallbackURL == "" {
			resolvedOIDC.CallbackURL = fmt.Sprintf("%s://%s/callback", scheme, waitHost)
		}
		makeup.PrintInfo(fmt.Sprintf("Using Cognito issuer %s", resolvedOIDC.IssuerURL))
	}
	if err := resolvedOIDC.validate(); err != nil {
		makeup.ExitDueToFatalError(err, "Invalid OIDC configuration for Klutch control plane.")
	}

	useDex = resolvedOIDC.Provider != OIDCProviderCognito

	if resolvedOIDC.IssuerURL != "" {
		if err := waitForOIDCDiscovery(resolvedOIDC.IssuerURL, 2*time.Minute); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("OIDC issuer %s is not reachable/valid. Provide explicit OIDC values or re-run with a working issuer.", resolvedOIDC.IssuerURL))
		}
	}

	if useDex {
		k.DeployDex(publicHost, ingressPort, ingressClass, scheme, acmCertificateARN)
		k.WaitForDex()
	} else {
		k.applyOIDCSecret(resolvedOIDC)
	}

	if err := applyControlPlaneCRDs(ctx, k.cpContext); err != nil {
		makeup.ExitDueToFatalError(err, "Failed to apply Klutch control-plane CRDs.")
	}

	k.DeployBindBackend(waitHost, ingressPort, ingressClass, scheme, acmCertificateARN, tlsEnabled, useDex)

	// If we're using ALB, wait for the ingress hostnames. When a hosted zone is available, create a CNAME
	// or ALIAS to the ALB and keep using the public hosts; otherwise fall back to using the ALB hostname directly.
	if ingressClass == "alb" {
		dexIngressHost := ""
		if useDex {
			dexIngressHost = waitForIngressHost(k.cpK8s, "dex-ingress", "default")
			if dexIngressHost == "" {
				makeup.ExitDueToFatalError(nil, "Could not determine ingress hostname/IP for dex. Aborting instead of using the Kubernetes API server host.")
			}
		}

		backendIngressHost := waitForIngressHost(k.cpK8s, "anynines-backend", "default")
		if backendIngressHost == "" {
			backendIngressHost = dexIngressHost
		}
		if backendIngressHost == "" {
			makeup.ExitDueToFatalError(nil, "Could not determine ingress hostname/IP for the Klutch backend.")
		}

		defaultIngressTarget := backendIngressHost
		if defaultIngressTarget == "" {
			defaultIngressTarget = dexIngressHost
		}

		if provisioner != nil && hostedZoneName != "" {
			hostSet := map[string]struct{}{}
			if publicHost != "" {
				hostSet[publicHost] = struct{}{}
			}
			if backendHost != "" {
				hostSet[backendHost] = struct{}{}
			}

			var aliasHosts []string
			records := map[string]string{}
			for h := range hostSet {
				target := defaultIngressTarget
				if h == backendHost && backendIngressHost != "" {
					target = backendIngressHost
				}
				if h == dexHost && dexIngressHost != "" {
					target = dexIngressHost
				}

				if h == zone {
					if target == "" {
						makeup.ExitDueToFatalError(nil, fmt.Sprintf("No ingress target available for alias host %s", h))
					}
					aliasHosts = append(aliasHosts, h)
					continue
				}

				records[h] = target
			}

			if len(aliasHosts) > 0 {
				makeup.PrintInfo(fmt.Sprintf("Planned action: create/update ALIAS %v -> %s in hosted zone %s for ingress.", aliasHosts, defaultIngressTarget, hostedZoneName))
				makeup.WaitForUser()
				for _, h := range aliasHosts {
					target := defaultIngressTarget
					if h == backendHost && backendIngressHost != "" {
						target = backendIngressHost
					}
					if h == dexHost && dexIngressHost != "" {
						target = dexIngressHost
					}
					if err := provisioner.EnsureALBAliasRecord(hostedZoneName, h, target); err != nil {
						makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not create ALIAS record in hosted zone %s.", hostedZoneName))
					}
				}
				makeup.PrintInfo(fmt.Sprintf("Ensured DNS ALIAS %v -> ingress hostnames in hosted zone %s.", aliasHosts, hostedZoneName))
			}

			if len(records) > 0 {
				makeup.PrintInfo(fmt.Sprintf("Planned action: create/update CNAMEs %v -> ingress hosts in hosted zone %s.", keys(records), hostedZoneName))
				makeup.WaitForUser()

				makeup.PrintInfo("Waiting for DNS CNAME propagation; this can take several minutes depending on your registrar (up to 30m).")
				if err := provisioner.EnsureCNAMERecords(hostedZoneName, records); err != nil {
					makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not create CNAME records in hosted zone %s.", hostedZoneName))
				}
				makeup.PrintInfo(fmt.Sprintf("Ensured DNS CNAMEs %v -> ingress hosts in hosted zone %s.", keys(records), hostedZoneName))

				// Verify DNS reflects the expected targets.
				for h, target := range records {
					waitForCNAMERecord(h, target, 30*time.Minute)
				}
			}
		}
	}

	k.WaitForBindBackend(waitHost, ingressPort, scheme)

	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, publicHost, ingressPort)

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser()
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
}

func detectIngressClass(k8sClient *k8s.KubeClient) string {
	output, err := k8sClient.Get("ingressclass", "", "", "jsonpath={.items[*].metadata.name}", false)
	if err != nil {
		makeup.PrintWarning(fmt.Sprintf("Could not detect ingress classes (defaulting to nginx): %v", err))
		return "nginx"
	}

	classes := strings.Fields(string(output))
	hasAlb := false
	hasNginx := false
	for _, c := range classes {
		if c == "alb" {
			hasAlb = true
		}
		if c == "nginx" {
			hasNginx = true
		}
	}

	if hasAlb {
		makeup.PrintInfo("Detected ingress class `alb`. Will use it and skip deploying ingress-nginx.")
		return "alb"
	}

	if hasNginx {
		makeup.PrintInfo("Detected ingress class `nginx`. Will use it.")
		return "nginx"
	}

	makeup.PrintWarning("No ingress class detected. Defaulting to `nginx` and deploying ingress-nginx.")
	return "nginx"
}

// determineIngressScheme returns the URL scheme to use for ingress endpoints.
func determineIngressScheme(_ string, tlsEnabled bool) string {
	if tlsEnabled {
		return "https"
	}
	return "http"
}

// formatHostWithPort returns host or host:port depending on scheme defaults.
func formatHostWithPort(scheme, host, port string) string {
	if port == "" {
		return host
	}
	if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
		return host
	}
	return host + ":" + port
}

// keys returns the keys of a string map for logging.
func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func defaultAWSRegion() string {
	envs := []string{"CONTROL_PLANE_CLUSTER_REGION", "WORKLOAD_CLUSTER_REGION", "AWS_REGION", "AWS_DEFAULT_REGION"}
	for _, e := range envs {
		if v := strings.TrimSpace(os.Getenv(e)); v != "" {
			return v
		}
	}
	return "eu-central-1"
}

func sanitizeCognitoPrefix(val string) string {
	val = strings.TrimSpace(val)
	val = strings.TrimSuffix(val, ".")
	for strings.Contains(val, "..") {
		val = strings.ReplaceAll(val, "..", ".")
	}
	if idx := strings.Index(val, "."); idx > 0 {
		val = val[:idx]
	}
	val = strings.ToLower(val)
	out := strings.Builder{}
	for _, r := range val {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			out.WriteRune(r)
		}
	}
	return strings.Trim(out.String(), "-")
}

// waitForOIDCDiscovery checks that the issuer exposes a valid discovery document, retrying for a while to allow hosted domains to become active.
func waitForOIDCDiscovery(issuer string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	url := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				resp.Body.Close()
				cancel()
				return nil
			}
			resp.Body.Close()
		}
		cancel()

		if time.Now().After(deadline) {
			if err != nil {
				return fmt.Errorf("failed to reach OIDC issuer %s: %w", issuer, err)
			}
			return fmt.Errorf("OIDC discovery at %s returned status %d", url, resp.StatusCode)
		}

		time.Sleep(5 * time.Second)
	}
}

// verifyHostedZoneRequirements ensures that the provided hosts are within the hosted zone
// and that the zone is properly delegated. It prints actionable instructions and exits if requirements are not met.
func verifyHostedZoneRequirements(provisioner CertificateProvisioner, hostedZoneName, baseDomain, dexHost, backendHost, clusterName string) {
	if hostedZoneName == "" || provisioner == nil {
		return
	}

	zone := strings.TrimSuffix(hostedZoneName, ".")
	requireHostInZone := func(host string, label string) {
		if host == "" {
			return
		}
		if host != zone && !strings.HasSuffix(host, "."+zone) {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("%s `%s` is not within hosted zone `%s`. Please use a hostname inside the zone.", label, host, zone))
		}
	}

	requireHostInZone(baseDomain, "Base domain")
	requireHostInZone(dexHost, "Dex host")
	requireHostInZone(backendHost, "Backend host")

	expectedNS, err := provisioner.GetHostedZoneNS(zone)
	if err != nil {
		makeup.PrintWarning(fmt.Sprintf("Hosted zone %s not found in Route53. It may have been deleted.", zone))
		makeup.PrintInfo(fmt.Sprintf("Creating public hosted zone %s and retrieving its NS records...", zone))
		makeup.WaitForUser()
		expectedNS, err = provisioner.EnsurePublicHostedZone(zone, clusterName)
		if err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not create or fetch NS records for hosted zone %s from Route53.", zone))
		}
	}

	liveNS, _ := lookupPublicNS(zone)
	if len(liveNS) == 0 {
		parent := parentDomain(zone)
		makeup.PrintWarning(fmt.Sprintf("Hosted zone %s is not publicly delegated.", zone))
		if parent != "" {
			makeup.PrintInfo(fmt.Sprintf("Create NS delegation records in parent zone %s (zone file format):", parent))
		} else {
			makeup.PrintInfo("Add the following NS delegation records (zone file format):")
		}
		for _, ns := range expectedNS {
			makeup.Print(fmt.Sprintf("%s\t300\tIN\tNS\t%s", ensureTrailingDot(zone), ensureTrailingDot(ns)))
		}
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("DNS delegation for %s is missing. Add the NS records above and rerun.", zone))
	}

	// If there is delegation but it doesn't match Route53 NS, warn with instructions.
	if !nsSetsMatch(expectedNS, liveNS) {
		parent := parentDomain(zone)
		makeup.PrintWarning(fmt.Sprintf("Hosted zone %s is not delegated to the expected Route53 nameservers.", zone))
		makeup.PrintInfo(fmt.Sprintf("Update the NS record in parent zone %s to the following (zone file format):", parent))
		for _, ns := range expectedNS {
			makeup.Print(fmt.Sprintf("%s.\t300\tIN\tNS\t%s.", zone, strings.TrimSuffix(ns, ".")))
		}
		makeup.PrintInfo("Waiting for delegation to reflect Route53 nameservers (up to 30m)...")

		deadline := time.Now().Add(30 * time.Minute)
		for time.Now().Before(deadline) {
			time.Sleep(15 * time.Second)
			liveNS, _ = lookupPublicNS(zone)
			if nsSetsMatch(expectedNS, liveNS) {
				makeup.PrintInfo("Delegation now matches Route53 nameservers.")
				return
			}
		}

		makeup.ExitDueToFatalError(nil, fmt.Sprintf("DNS delegation for %s does not match Route53 after waiting. Please update the NS records and rerun.", zone))
	}
}

// parentDomain returns the parent domain of a FQDN (e.g., hub.a9s.io -> a9s.io).
func parentDomain(domain string) string {
	parts := strings.Split(strings.TrimSuffix(domain, "."), ".")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[1:], ".")
}

// waitForIngressHost waits for an ingress to report a load balancer hostname/IP and returns it.
func waitForIngressHost(k8sClient *k8s.KubeClient, name, namespace string) string {
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(15 * time.Second)
	for {
		select {
		case <-timeout:

			if out, err := k8sClient.Describe("ingress", name, namespace); err == nil {
				makeup.PrintWarning(fmt.Sprintf("Ingress %s/%s description:\n%s", namespace, name, string(out)))
			}
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Timed out waiting for ingress %s/%s to have a hostname/IP", namespace, name))
		case <-tick:
			output, err := k8sClient.Get("ingress", name, namespace, "jsonpath={.status.loadBalancer.ingress[0].hostname}", false)
			if err != nil {
				makeup.PrintWarning(fmt.Sprintf("Error while checking ingress %s/%s hostname: %v", namespace, name, err))
				continue
			}
			host := strings.TrimSpace(string(output))
			if host != "" {
				return host
			}

			// try IP as fallback
			ipOutput, err := k8sClient.Get("ingress", name, namespace, "jsonpath={.status.loadBalancer.ingress[0].ip}", false)
			if err == nil {
				ip := strings.TrimSpace(string(ipOutput))
				if ip != "" {
					return ip
				}
			}
			makeup.PrintInfo(fmt.Sprintf("Ingress %s/%s has no hostname/IP yet. Waiting...", namespace, name))
		}
	}
}

// waitForServiceHost waits for a Service to report a load balancer hostname/IP and returns it.
func waitForServiceHost(k8sClient *k8s.KubeClient, name, namespace string) string {
	timeout := time.After(15 * time.Minute)
	tick := time.Tick(15 * time.Second)
	for {
		select {
		case <-timeout:
			if out, err := k8sClient.Describe("svc", name, namespace); err == nil {
				makeup.PrintWarning(fmt.Sprintf("Service %s/%s description:\n%s", namespace, name, string(out)))
			}
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Timed out waiting for service %s/%s to have a hostname/IP", namespace, name))
		case <-tick:
			output, err := k8sClient.Get("svc", name, namespace, "jsonpath={.status.loadBalancer.ingress[0].hostname}", false)
			if err != nil {
				makeup.PrintWarning(fmt.Sprintf("Error while checking service %s/%s hostname: %v", namespace, name, err))
				continue
			}
			host := strings.TrimSpace(string(output))
			if host != "" {
				return host
			}

			ipOutput, err := k8sClient.Get("svc", name, namespace, "jsonpath={.status.loadBalancer.ingress[0].ip}", false)
			if err == nil {
				ip := strings.TrimSpace(string(ipOutput))
				if ip != "" {
					return ip
				}
			}
			makeup.PrintInfo(fmt.Sprintf("Service %s/%s has no hostname/IP yet. Waiting...", namespace, name))
		}
	}
}

// waitForCNAMERecord waits until a CNAME record resolves to the expected target (substring match).
func waitForCNAMERecord(host, expectedTarget string, timeout time.Duration) {
	if host == "" || expectedTarget == "" {
		return
	}

	deadline := time.Now().Add(timeout)
	makeup.PrintInfo(fmt.Sprintf("Waiting for CNAME %s to point to %s (timeout %s)...", host, expectedTarget, timeout))
	for {
		cname, err := net.LookupCNAME(host)
		if err == nil && strings.Contains(cname, expectedTarget) {
			return
		}

		if time.Now().After(deadline) {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("CNAME %s did not point to %s within %s (got %s)", host, expectedTarget, timeout, cname))
		}

		makeup.PrintInfo(fmt.Sprintf("CNAME %s not pointing to %s yet (got %s). Waiting...", host, expectedTarget, cname))
		time.Sleep(5 * time.Second)
	}
}

func lookupPublicNS(name string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var lastErr error
	for _, server := range publicResolvers {
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: 3 * time.Second}
				return d.DialContext(ctx, "udp", server)
			},
		}
		records, err := r.LookupNS(ctx, name)
		if err == nil && len(records) > 0 {
			out := make([]string, 0, len(records))
			for _, ns := range records {
				out = append(out, strings.TrimSuffix(ns.Host, "."))
			}
			return out, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func nsSetsMatch(expected []string, live []string) bool {
	if len(expected) == 0 || len(live) == 0 {
		return false
	}

	liveSet := make(map[string]struct{}, len(live))
	for _, ns := range live {
		liveSet[strings.TrimSuffix(strings.ToLower(ns), ".")] = struct{}{}
	}

	for _, ns := range expected {
		if _, ok := liveSet[strings.TrimSuffix(strings.ToLower(ns), ".")]; !ok {
			return false
		}
	}

	return true
}

func applyControlPlaneCRDs(ctx context.Context, kubeContext string) error {
	crds, err := crd.CRDs()
	if err != nil {
		return fmt.Errorf("failed to load Klutch CRDs: %w", err)
	}
	var docs []string
	for i := range crds {
		y, err := sigyaml.Marshal(&crds[i])
		if err != nil {
			return fmt.Errorf("failed to render CRD %s: %w", crds[i].Name, err)
		}
		docs = append(docs, string(y))
	}
	manifest := strings.Join(docs, "\n---\n")
	makeup.PrintInfo(fmt.Sprintf("Applying Klutch control-plane CRDs to context %q...", kubeContext))
	k8sClient := k8s.NewKubeClient(kubeContext)
	if _, err := k8sClient.ApplyWithPrompt([]byte(manifest), "Klutch Control-Plane CRDs"); err != nil {
		return fmt.Errorf("failed to apply Klutch control-plane CRDs: %w", err)
	}
	makeup.PrintCheckmark("Applied Klutch control-plane CRDs.")
	return nil
}

func printControlPlaneSummary(workDir string) {
	filePath := filepath.Join(workDir, controlPlaneClusterInfoFilePath, controlPlaneClusterInfoFileName)

	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Installed the Klutch control plane components into the current Kubernetes cluster.")
	makeup.PrintCheckmark("Configured OIDC and deployed the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to App Clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark(fmt.Sprintf("Wrote Control Plane Cluster information to %s", filePath))
	makeup.PrintSuccessSummary("You are now ready to bind APIs from an App Cluster using the `a9s klutch bind` command.")
}

// Writes information about the Control Plane Cluster to a file and ConfigMap, to give other commands such as `bind` the information they need.
func writeControlPlaneClusterInfoToFile(workDir string, hostIP string, ingressPort string) {
	info := &ControlPlaneClusterInfo{
		Host:        hostIP,
		IngressPort: ingressPort,
	}

	data, err := yaml.Marshal(info)
	if err != nil {
		makeup.ExitDueToFatalError(err, "Unexpected error while writing Control Plane Cluster info to file.")
	}

	path := filepath.Join(workDir, controlPlaneClusterInfoFilePath)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while writing Control Plane Cluster info to file. Could not create path %s", path))
	}

	file := filepath.Join(path, controlPlaneClusterInfoFileName)
	err = os.WriteFile(file, data, 0644)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while writing Control Plane Cluster info to file. Could not write to file %s", file))
	}

	makeup.PrintInfo(fmt.Sprintf("Wrote Control Plane Cluster information to %s", file))

	// Also store in-cluster ConfigMap for discovery without local files.
	if err := SaveControlPlaneInfoToCluster("", *info); err != nil {
		makeup.PrintWarning(fmt.Sprintf("Could not store control plane info ConfigMap: %v", err))
	} else {
		makeup.PrintInfo("Stored Control Plane info ConfigMap in cluster.")
	}
}

// Checks if prerequisites of the deploy command are met.
func checkDeployPrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolGit],
		commonTools[prereq.ToolDocker],
		commonTools[prereq.ToolKind],
		commonTools[prereq.ToolKubectl],
		commonTools[prereq.ToolHelm],
	}

	prereq.CheckRequiredTools(requiredTools)

	prereq.CheckDockerRunning()
}

// Checks if prerequisites of the control plane install command are met.
func checkControlPlaneInstallPrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolGit],
		commonTools[prereq.ToolKubectl],
		commonTools[prereq.ToolHelm],
	}

	prereq.CheckRequiredTools(requiredTools)
}
