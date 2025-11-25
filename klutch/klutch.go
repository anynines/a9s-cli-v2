package klutch

import (
	_ "embed"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/k8s"
	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
	"gopkg.in/yaml.v2"
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
	PortFlag int = 8080
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
		demo.UnattendedMode,
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
func ApplyKlutchControlPlane(host string, ingressPort int, acmCertificateARN string, hostedZoneName string) {
	makeup.PrintWelcomeScreen(
		demo.UnattendedMode,
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

	var provisioner CertificateProvisioner
	if hostedZoneName != "" {
		verifyHostedZoneResolvable(hostedZoneName)
		provisioner = NewCertificateProvisioner("")
		verifyHostedZoneRequirements(provisioner, hostedZoneName, baseDomain, dexHost)
	}

	dexHost := ""
	if baseDomain != "" {
		if strings.HasPrefix(baseDomain, "dex.") {
			dexHost = baseDomain
		} else {
			dexHost = fmt.Sprintf("dex.%s", baseDomain)
		}
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
		makeup.WaitForUser(demo.UnattendedMode)

		makeup.PrintInfo(fmt.Sprintf("Requesting ACM certificate for %s (SANs: %v) in hosted zone %s.", primary, altNames, hostedZoneName))
		arn, err := provisioner.EnsureCertificate(primary, altNames, hostedZoneName)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to request ACM certificate.")
		}
		acmCertificateARN = arn
		makeup.PrintInfo(fmt.Sprintf("Using ACM certificate ARN: %s", acmCertificateARN))
	}

	manager := NewKlutchManagerWithContexts("", "")
	manager.applyControlPlaneToContext(baseDomain, dexHost, hostedZoneName, provisioner, strconv.Itoa(ingressPort), acmCertificateARN)
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

	k.DeployBindBackend(port, ingressClass, scheme)
	k.WaitForBindBackend()

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
}

func (k *KlutchManager) deployAppCluster() {

	DeployAppCluster(appClusterName)
	WaitForKindCluster(k.appK8s)

	switchContext(k.appContext) // in case the app cluster already existed, switch to it.
}

func printSummary() {
	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Deployed a Klutch Control Plane Cluster with Kind.")
	makeup.PrintCheckmark("Deployed Dex Idp and the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to App Clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark("Deployed an App Cluster.")
	makeup.PrintSuccessSummary("You are now ready to bind APIs from the App Cluster using the `a9s klutch bind` command.")
}

func (k *KlutchManager) applyControlPlaneToContext(baseDomain string, dexHost string, hostedZoneName string, provisioner CertificateProvisioner, ingressPort string, acmCertificateARN string) {
	ingressClass := detectIngressClass(k.cpK8s)
	tlsEnabled := acmCertificateARN != "" && ingressClass == "alb"

	// ALB ingress is created without TLS in our manifests; default to port 80 when no cert is provided
	// to avoid calling a closed HTTPS listener. When a cert is provided, force port 443.
	if tlsEnabled && ingressPort != "443" {
		makeup.PrintInfo("ACM certificate provided; enabling TLS on ALB and using port 443.")
		ingressPort = "443"
	} else if ingressClass == "alb" && ingressPort == "443" {
		makeup.PrintInfo("Detected ALB ingress and no TLS configuration. Switching ingress port to 80.")
		ingressPort = "80"
	}

	scheme := determineIngressScheme(ingressClass, tlsEnabled)

	publicHost := dexHost
	if publicHost == "" {
		publicHost = baseDomain
	}
	if publicHost == "" {
		publicHost = getClusterExternalHost("")
		makeup.PrintInfo(fmt.Sprintf("No host provided. Using provisional host `%s` to bootstrap deployment.", publicHost))
	} else {
		makeup.PrintInfo(fmt.Sprintf("Using host `%s` for Dex and OIDC configuration.", publicHost))
	}

	if ingressClass == "nginx" {
		k.DeployIngressNginx()
		k.WaitForIngressNginx()
	} else {
		makeup.PrintInfo(fmt.Sprintf("Ingress class `%s` detected. Skipping ingress-nginx installation.", ingressClass))
	}

	k.DeployDex(publicHost, ingressPort, ingressClass, scheme, acmCertificateARN)
	k.WaitForDex()

	k.DeployBindBackend(ingressPort, ingressClass, scheme)

	// If we're using ALB, wait for the ALB hostname. When a hosted zone is available, create a CNAME
	// to the ALB and keep using the public host; otherwise fall back to using the ALB hostname directly.
	if ingressClass == "alb" {
		resolvedHost := waitForIngressHost(k.cpK8s, "dex-ingress", "default")
		if resolvedHost == "" {
			makeup.ExitDueToFatalError(nil, "Could not determine ingress hostname/IP. Aborting instead of using the Kubernetes API server host.")
		}

		if provisioner != nil && hostedZoneName != "" && publicHost != "" {
			records := map[string]string{
				publicHost: resolvedHost,
			}
			// Also wire the base domain if it differs, so hub.a9s.io resolves too.
			if baseDomain != "" && baseDomain != publicHost {
				records[baseDomain] = resolvedHost
			}

			if len(records) > 0 {
				makeup.PrintInfo(fmt.Sprintf("Planned action: create/update CNAMEs %v -> %s in hosted zone %s for ingress.", keys(records), resolvedHost, hostedZoneName))
				makeup.WaitForUser(demo.UnattendedMode)

				if err := provisioner.EnsureCNAMERecords(hostedZoneName, records); err != nil {
					makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not create CNAME records in hosted zone %s.", hostedZoneName))
				}
				makeup.PrintInfo(fmt.Sprintf("Ensured DNS CNAMEs %v -> %s in hosted zone %s.", keys(records), resolvedHost, hostedZoneName))
			}
		}
	}

	k.WaitForBindBackend()

	writeControlPlaneClusterInfoToFile(demo.DemoConfig.WorkingDir, publicHost, ingressPort)

	k.DeployCrossplaneComponents()

	makeup.H1("Klutch components are deployed. Deploying the a8s stack...")
	makeup.WaitForUser(demo.UnattendedMode)
	a8s := demo.NewA8sDemoManager(k.cpContext)
	a8s.DeployA8sStack()
}

func detectIngressClass(k8sClient *k8s.KubeClient) string {
	cmd := k8sClient.KubectlWithContextCommand("get", "ingressclass", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.Output()
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

// keys returns the keys of a string map for logging.
func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// verifyHostedZoneRequirements ensures that the provided hosts are within the hosted zone
// and that the zone is properly delegated. It prints actionable instructions and exits if requirements are not met.
func verifyHostedZoneRequirements(provisioner CertificateProvisioner, hostedZoneName, baseDomain, dexHost string) {
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

	expectedNS, err := provisioner.GetHostedZoneNS(zone)
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not fetch NS records for hosted zone %s from Route53.", zone))
	}

	liveNS, _ := net.LookupNS(zone)
	if len(liveNS) == 0 {
		parent := parentDomain(zone)
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Hosted zone %s is not publicly delegated. Create an NS record in parent zone %s with these values: %s", zone, parent, strings.Join(expectedNS, ", ")))
	}

	// If there is delegation but it doesn't match Route53 NS, warn with instructions.
	delegated := make(map[string]struct{}, len(liveNS))
	for _, ns := range liveNS {
		delegated[strings.TrimSuffix(strings.ToLower(ns.Host), ".")] = struct{}{}
	}
	mismatch := false
	for _, ns := range expectedNS {
		n := strings.TrimSuffix(strings.ToLower(ns), ".")
		if _, ok := delegated[n]; !ok {
			mismatch = true
			break
		}
	}
	if mismatch {
		parent := parentDomain(zone)
		makeup.ExitDueToFatalError(nil, fmt.Sprintf("Hosted zone %s is not delegated to the expected Route53 nameservers. Update the NS record in parent zone %s to: %s", zone, parent, strings.Join(expectedNS, ", ")))
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
			descCmd := k8sClient.KubectlWithContextCommand("describe", "ingress", name, "-n", namespace)
			if out, err := descCmd.CombinedOutput(); err == nil {
				makeup.PrintWarning(fmt.Sprintf("Ingress %s/%s description:\n%s", namespace, name, string(out)))
			}
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Timed out waiting for ingress %s/%s to have a hostname/IP", namespace, name))
		case <-tick:
			cmd := k8sClient.KubectlWithContextCommand("get", "ingress", name, "-n", namespace, "-o", "jsonpath={.status.loadBalancer.ingress[0].hostname}")
			output, err := cmd.Output()
			if err != nil {
				makeup.PrintWarning(fmt.Sprintf("Error while checking ingress %s/%s hostname: %v", namespace, name, err))
				continue
			}
			host := strings.TrimSpace(string(output))
			if host != "" {
				return host
			}

			// try IP as fallback
			cmdIP := k8sClient.KubectlWithContextCommand("get", "ingress", name, "-n", namespace, "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
			ipOutput, err := cmdIP.Output()
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

func printControlPlaneSummary(workDir string) {
	filePath := filepath.Join(workDir, controlPlaneClusterInfoFilePath, controlPlaneClusterInfoFileName)

	makeup.PrintH1("Summary")
	makeup.Print("You've successfully accomplished the followings steps:")
	makeup.PrintCheckmark("Installed the Klutch control plane components into the current Kubernetes cluster.")
	makeup.PrintCheckmark("Deployed Dex Idp and the anynines klutch-bind backend.")
	makeup.PrintCheckmark("Deployed Crossplane and the Kubernetes provider.")
	makeup.PrintCheckmark("Deployed the Klutch Crossplane configuration package.")
	makeup.PrintCheckmark("Deployed Klutch API Service Export Templates to make the Klutch Crossplane APIs available to App Clusters.")
	makeup.PrintCheckmark("Deployed the a8s Stack.")
	makeup.PrintCheckmark(fmt.Sprintf("Wrote Control Plane Cluster information to %s", filePath))
	makeup.PrintSuccessSummary("You are now ready to bind APIs from an App Cluster using the `a9s klutch bind` command.")
}

// Writes information about the Control Plane Cluster to a file, to give other commands such as `bind` the information they need.
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
