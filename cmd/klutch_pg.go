package cmd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

const (
	klutchPGAPIVersion             = "anynines.com/v1"
	klutchPGInstanceKind           = "PostgresqlInstance"
	klutchPGServiceBindingKind     = "ServiceBinding"
	klutchPGBackupKind             = "Backup"
	klutchPGRestoreKind            = "Restore"
	klutchPGInstanceResource       = "postgresqlinstances.anynines.com"
	klutchPGServiceBindingResource = "servicebindings.anynines.com"
	klutchPGBackupResource         = "backups.anynines.com"
	klutchPGRestoreResource        = "restores.anynines.com"
)

var createKlutchPGInstanceName string
var createKlutchPGInstanceNamespace string
var createKlutchPGInstanceService string
var createKlutchPGInstancePlan string
var createKlutchPGInstanceExpose string
var createKlutchPGInstanceComposition string
var createKlutchPGInstanceNoApply bool
var createKlutchPGInstanceWait bool
var createKlutchPGInstanceWaitTimeout string

var createKlutchPGServiceBindingName string
var createKlutchPGServiceBindingNamespace string
var createKlutchPGServiceBindingInstanceRef string
var createKlutchPGServiceBindingInstanceType string
var createKlutchPGServiceBindingComposition string
var createKlutchPGServiceBindingNoApply bool
var createKlutchPGServiceBindingWait bool
var createKlutchPGServiceBindingWaitTimeout string

var createKlutchPGBackupName string
var createKlutchPGBackupNamespace string
var createKlutchPGBackupInstanceRef string
var createKlutchPGBackupInstanceType string
var createKlutchPGBackupComposition string
var createKlutchPGBackupNoApply bool
var createKlutchPGBackupWait bool
var createKlutchPGBackupWaitTimeout string

var createKlutchPGRestoreName string
var createKlutchPGRestoreNamespace string
var createKlutchPGRestoreBackupRef string
var createKlutchPGRestoreInstanceRef string
var createKlutchPGRestoreInstanceType string
var createKlutchPGRestoreComposition string
var createKlutchPGRestoreNoApply bool
var createKlutchPGRestoreWait bool
var createKlutchPGRestoreWaitTimeout string

var deleteKlutchPGInstanceName string
var deleteKlutchPGInstanceNamespace string
var deleteKlutchPGInstanceWait bool
var deleteKlutchPGInstanceWaitTimeout string

var deleteKlutchPGServiceBindingName string
var deleteKlutchPGServiceBindingNamespace string
var deleteKlutchPGServiceBindingWait bool
var deleteKlutchPGServiceBindingWaitTimeout string

var deleteKlutchPGBackupName string
var deleteKlutchPGBackupNamespace string
var deleteKlutchPGBackupWait bool
var deleteKlutchPGBackupWaitTimeout string

var deleteKlutchPGRestoreName string
var deleteKlutchPGRestoreNamespace string
var deleteKlutchPGRestoreWait bool
var deleteKlutchPGRestoreWaitTimeout string

var cmdCreateKlutchPG = &cobra.Command{
	Use:   "pg",
	Short: "Create Klutch-managed PostgreSQL claim resources.",
	Long:  `Create Klutch-managed PostgreSQL claim resources on a workload cluster bound via klutch-bind.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdCreateKlutchPGInstance = &cobra.Command{
	Use:   "instance",
	Short: "Create a Klutch-managed PostgreSQL instance claim.",
	Long:  `Creates an anynines.com/v1 PostgresqlInstance claim for a Klutch-managed PostgreSQL service instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(createKlutchPGInstanceName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}

		manifest, err := buildKlutchPGInstanceManifest(
			createKlutchPGInstanceName,
			createKlutchPGInstanceNamespace,
			createKlutchPGInstanceService,
			createKlutchPGInstancePlan,
			createKlutchPGInstanceExpose,
			createKlutchPGInstanceComposition,
		)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to render Klutch PostgreSQL instance manifest.")
		}

		if createKlutchPGInstanceNoApply {
			makeup.PrintInfo("Skipping apply because --no-apply was provided.")
			makeup.PrintYAML(manifest, false)
			return
		}

		if output, err := runKubectlWithInput(manifest, "apply", "-f", "-"); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to create Klutch PostgreSQL instance.\n%s", strings.TrimSpace(output)))
		}

		if createKlutchPGInstanceWait {
			resourceName := fmt.Sprintf("%s/%s", klutchPGInstanceResource, createKlutchPGInstanceName)
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", createKlutchPGInstanceNamespace,
				"--for=condition=Ready",
				"--timeout", createKlutchPGInstanceWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch PostgreSQL instance did not become ready.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintSuccessSummary(fmt.Sprintf("Klutch PostgreSQL instance %s created in namespace %s.", createKlutchPGInstanceName, createKlutchPGInstanceNamespace))
	},
}

var cmdCreateKlutchPGServiceBinding = &cobra.Command{
	Use:   "servicebinding",
	Short: "Create a Klutch-managed PostgreSQL service binding claim.",
	Long:  `Creates an anynines.com/v1 ServiceBinding claim for a Klutch-managed PostgreSQL instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(createKlutchPGServiceBindingName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}
		if strings.TrimSpace(createKlutchPGServiceBindingInstanceRef) == "" {
			makeup.ExitDueToFatalError(nil, "The --service-instance flag is required.")
		}

		exists, err := klutchResourceExists(klutchPGInstanceResource, createKlutchPGServiceBindingInstanceRef, createKlutchPGServiceBindingNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL instance before creating service binding.")
		}
		if !exists {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create Klutch service binding for non-existing service instance %s in namespace %s", createKlutchPGServiceBindingInstanceRef, createKlutchPGServiceBindingNamespace))
		}

		manifest, err := buildKlutchPGServiceBindingManifest(
			createKlutchPGServiceBindingName,
			createKlutchPGServiceBindingNamespace,
			createKlutchPGServiceBindingInstanceRef,
			createKlutchPGServiceBindingInstanceType,
			createKlutchPGServiceBindingComposition,
		)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to render Klutch PostgreSQL service binding manifest.")
		}

		if createKlutchPGServiceBindingNoApply {
			makeup.PrintInfo("Skipping apply because --no-apply was provided.")
			makeup.PrintYAML(manifest, false)
			return
		}

		if output, err := runKubectlWithInput(manifest, "apply", "-f", "-"); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to create Klutch PostgreSQL service binding.\n%s", strings.TrimSpace(output)))
		}

		if createKlutchPGServiceBindingWait {
			resourceName := fmt.Sprintf("%s/%s", klutchPGServiceBindingResource, createKlutchPGServiceBindingName)
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", createKlutchPGServiceBindingNamespace,
				"--for=condition=Ready",
				"--timeout", createKlutchPGServiceBindingWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch PostgreSQL service binding did not become ready.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintSuccessSummary(fmt.Sprintf("Klutch PostgreSQL service binding %s created in namespace %s.", createKlutchPGServiceBindingName, createKlutchPGServiceBindingNamespace))
	},
}

var cmdCreateKlutchPGBackup = &cobra.Command{
	Use:   "backup",
	Short: "Create a Klutch-managed PostgreSQL backup claim.",
	Long:  `Creates an anynines.com/v1 Backup claim for a Klutch-managed PostgreSQL instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(createKlutchPGBackupName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}
		if strings.TrimSpace(createKlutchPGBackupInstanceRef) == "" {
			makeup.ExitDueToFatalError(nil, "The --service-instance flag is required.")
		}

		exists, err := klutchResourceExists(klutchPGInstanceResource, createKlutchPGBackupInstanceRef, createKlutchPGBackupNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL instance before creating backup.")
		}
		if !exists {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create Klutch backup for non-existing service instance %s in namespace %s", createKlutchPGBackupInstanceRef, createKlutchPGBackupNamespace))
		}

		manifest, err := buildKlutchPGBackupManifest(
			createKlutchPGBackupName,
			createKlutchPGBackupNamespace,
			createKlutchPGBackupInstanceRef,
			createKlutchPGBackupInstanceType,
			createKlutchPGBackupComposition,
		)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to render Klutch PostgreSQL backup manifest.")
		}

		if createKlutchPGBackupNoApply {
			makeup.PrintInfo("Skipping apply because --no-apply was provided.")
			makeup.PrintYAML(manifest, false)
			return
		}

		if output, err := runKubectlWithInput(manifest, "apply", "-f", "-"); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to create Klutch PostgreSQL backup.\n%s", strings.TrimSpace(output)))
		}

		if createKlutchPGBackupWait {
			resourceName := fmt.Sprintf("%s/%s", klutchPGBackupResource, createKlutchPGBackupName)
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", createKlutchPGBackupNamespace,
				"--for=condition=Ready",
				"--timeout", createKlutchPGBackupWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch PostgreSQL backup did not become ready.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintSuccessSummary(fmt.Sprintf("Klutch PostgreSQL backup %s created in namespace %s.", createKlutchPGBackupName, createKlutchPGBackupNamespace))
	},
}

var cmdCreateKlutchPGRestore = &cobra.Command{
	Use:   "restore",
	Short: "Create a Klutch-managed PostgreSQL restore claim.",
	Long:  `Creates an anynines.com/v1 Restore claim for a Klutch-managed PostgreSQL instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(createKlutchPGRestoreName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}
		if strings.TrimSpace(createKlutchPGRestoreBackupRef) == "" {
			makeup.ExitDueToFatalError(nil, "The --backup flag is required.")
		}
		if strings.TrimSpace(createKlutchPGRestoreInstanceRef) == "" {
			makeup.ExitDueToFatalError(nil, "The --service-instance flag is required.")
		}

		backupExists, err := klutchResourceExists(klutchPGBackupResource, createKlutchPGRestoreBackupRef, createKlutchPGRestoreNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL backup before creating restore.")
		}
		if !backupExists {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create Klutch restore for non-existing backup %s in namespace %s", createKlutchPGRestoreBackupRef, createKlutchPGRestoreNamespace))
		}

		instanceExists, err := klutchResourceExists(klutchPGInstanceResource, createKlutchPGRestoreInstanceRef, createKlutchPGRestoreNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL instance before creating restore.")
		}
		if !instanceExists {
			makeup.ExitDueToFatalError(nil, fmt.Sprintf("Can't create Klutch restore for non-existing service instance %s in namespace %s", createKlutchPGRestoreInstanceRef, createKlutchPGRestoreNamespace))
		}

		manifest, err := buildKlutchPGRestoreManifest(
			createKlutchPGRestoreName,
			createKlutchPGRestoreNamespace,
			createKlutchPGRestoreBackupRef,
			createKlutchPGRestoreInstanceRef,
			createKlutchPGRestoreInstanceType,
			createKlutchPGRestoreComposition,
		)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to render Klutch PostgreSQL restore manifest.")
		}

		if createKlutchPGRestoreNoApply {
			makeup.PrintInfo("Skipping apply because --no-apply was provided.")
			makeup.PrintYAML(manifest, false)
			return
		}

		if output, err := runKubectlWithInput(manifest, "apply", "-f", "-"); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Failed to create Klutch PostgreSQL restore.\n%s", strings.TrimSpace(output)))
		}

		if createKlutchPGRestoreWait {
			resourceName := fmt.Sprintf("%s/%s", klutchPGRestoreResource, createKlutchPGRestoreName)
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", createKlutchPGRestoreNamespace,
				"--for=condition=Ready",
				"--timeout", createKlutchPGRestoreWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch PostgreSQL restore did not become ready.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintSuccessSummary(fmt.Sprintf("Klutch PostgreSQL restore %s created in namespace %s.", createKlutchPGRestoreName, createKlutchPGRestoreNamespace))
	},
}

var cmdDeleteKlutchPG = &cobra.Command{
	Use:   "pg",
	Short: "Delete Klutch-managed PostgreSQL claim resources.",
	Long:  `Delete Klutch-managed PostgreSQL claim resources from a workload cluster bound via klutch-bind.`,
	Run: func(cmd *cobra.Command, args []string) {
		makeup.PrintWarning(" " + "Please use a sub-command.")
		cmd.Help()
	},
}

var cmdDeleteKlutchPGInstance = &cobra.Command{
	Use:   "instance",
	Short: "Delete a Klutch-managed PostgreSQL instance claim.",
	Long:  `Deletes an anynines.com/v1 PostgresqlInstance claim.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(deleteKlutchPGInstanceName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}

		exists, err := klutchResourceExists(klutchPGInstanceResource, deleteKlutchPGInstanceName, deleteKlutchPGInstanceNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL instance before deletion.")
		}
		if !exists {
			makeup.PrintWarning(fmt.Sprintf("Can't delete Klutch service instance. Service instance %s doesn't exist in namespace %s!", deleteKlutchPGInstanceName, deleteKlutchPGInstanceNamespace))
			return
		}

		resourceName := fmt.Sprintf("%s/%s", klutchPGInstanceResource, deleteKlutchPGInstanceName)
		if output, err := runKubectlWithInput(nil, "delete", resourceName, "-n", deleteKlutchPGInstanceNamespace); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Couldn't delete Klutch service instance.\n%s", strings.TrimSpace(output)))
		}

		if deleteKlutchPGInstanceWait {
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", deleteKlutchPGInstanceNamespace,
				"--for=delete",
				"--timeout", deleteKlutchPGInstanceWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch service instance deletion did not complete.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintCheckmark(fmt.Sprintf("Klutch service instance %s successfully deleted from namespace %s.", deleteKlutchPGInstanceName, deleteKlutchPGInstanceNamespace))
	},
}

var cmdDeleteKlutchPGServiceBinding = &cobra.Command{
	Use:   "servicebinding",
	Short: "Delete a Klutch-managed PostgreSQL service binding claim.",
	Long:  `Deletes an anynines.com/v1 ServiceBinding claim.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(deleteKlutchPGServiceBindingName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}

		exists, err := klutchResourceExists(klutchPGServiceBindingResource, deleteKlutchPGServiceBindingName, deleteKlutchPGServiceBindingNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL service binding before deletion.")
		}
		if !exists {
			makeup.PrintWarning(fmt.Sprintf("Can't delete Klutch service binding. Service binding %s doesn't exist in namespace %s!", deleteKlutchPGServiceBindingName, deleteKlutchPGServiceBindingNamespace))
			return
		}

		resourceName := fmt.Sprintf("%s/%s", klutchPGServiceBindingResource, deleteKlutchPGServiceBindingName)
		if output, err := runKubectlWithInput(nil, "delete", resourceName, "-n", deleteKlutchPGServiceBindingNamespace); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Couldn't delete Klutch service binding.\n%s", strings.TrimSpace(output)))
		}

		if deleteKlutchPGServiceBindingWait {
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", deleteKlutchPGServiceBindingNamespace,
				"--for=delete",
				"--timeout", deleteKlutchPGServiceBindingWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch service binding deletion did not complete.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintCheckmark(fmt.Sprintf("Klutch service binding %s successfully deleted from namespace %s.", deleteKlutchPGServiceBindingName, deleteKlutchPGServiceBindingNamespace))
	},
}

var cmdDeleteKlutchPGBackup = &cobra.Command{
	Use:   "backup",
	Short: "Delete a Klutch-managed PostgreSQL backup claim.",
	Long:  `Deletes an anynines.com/v1 Backup claim.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(deleteKlutchPGBackupName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}

		exists, err := klutchResourceExists(klutchPGBackupResource, deleteKlutchPGBackupName, deleteKlutchPGBackupNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL backup before deletion.")
		}
		if !exists {
			makeup.PrintWarning(fmt.Sprintf("Can't delete Klutch backup. Backup %s doesn't exist in namespace %s!", deleteKlutchPGBackupName, deleteKlutchPGBackupNamespace))
			return
		}

		resourceName := fmt.Sprintf("%s/%s", klutchPGBackupResource, deleteKlutchPGBackupName)
		if output, err := runKubectlWithInput(nil, "delete", resourceName, "-n", deleteKlutchPGBackupNamespace); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Couldn't delete Klutch backup.\n%s", strings.TrimSpace(output)))
		}

		if deleteKlutchPGBackupWait {
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", deleteKlutchPGBackupNamespace,
				"--for=delete",
				"--timeout", deleteKlutchPGBackupWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch backup deletion did not complete.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintCheckmark(fmt.Sprintf("Klutch backup %s successfully deleted from namespace %s.", deleteKlutchPGBackupName, deleteKlutchPGBackupNamespace))
	},
}

var cmdDeleteKlutchPGRestore = &cobra.Command{
	Use:   "restore",
	Short: "Delete a Klutch-managed PostgreSQL restore claim.",
	Long:  `Deletes an anynines.com/v1 Restore claim.`,
	Run: func(cmd *cobra.Command, args []string) {
		if strings.TrimSpace(deleteKlutchPGRestoreName) == "" {
			makeup.ExitDueToFatalError(nil, "The --name flag is required.")
		}

		exists, err := klutchResourceExists(klutchPGRestoreResource, deleteKlutchPGRestoreName, deleteKlutchPGRestoreNamespace)
		if err != nil {
			makeup.ExitDueToFatalError(err, "Failed to verify Klutch PostgreSQL restore before deletion.")
		}
		if !exists {
			makeup.PrintWarning(fmt.Sprintf("Can't delete Klutch restore. Restore %s doesn't exist in namespace %s!", deleteKlutchPGRestoreName, deleteKlutchPGRestoreNamespace))
			return
		}

		resourceName := fmt.Sprintf("%s/%s", klutchPGRestoreResource, deleteKlutchPGRestoreName)
		if output, err := runKubectlWithInput(nil, "delete", resourceName, "-n", deleteKlutchPGRestoreNamespace); err != nil {
			makeup.ExitDueToFatalError(err, fmt.Sprintf("Couldn't delete Klutch restore.\n%s", strings.TrimSpace(output)))
		}

		if deleteKlutchPGRestoreWait {
			if output, err := runKubectlWithInput(nil,
				"wait",
				resourceName,
				"-n", deleteKlutchPGRestoreNamespace,
				"--for=delete",
				"--timeout", deleteKlutchPGRestoreWaitTimeout,
			); err != nil {
				makeup.ExitDueToFatalError(err, fmt.Sprintf("Klutch restore deletion did not complete.\n%s", strings.TrimSpace(output)))
			}
		}

		makeup.PrintCheckmark(fmt.Sprintf("Klutch restore %s successfully deleted from namespace %s.", deleteKlutchPGRestoreName, deleteKlutchPGRestoreNamespace))
	},
}

func buildKlutchPGInstanceManifest(name, namespace, service, plan, expose, composition string) ([]byte, error) {
	manifest := map[string]interface{}{
		"apiVersion": klutchPGAPIVersion,
		"kind":       klutchPGInstanceKind,
		"metadata": map[string]interface{}{
			"name":      strings.TrimSpace(name),
			"namespace": strings.TrimSpace(namespace),
		},
		"spec": map[string]interface{}{
			"service": strings.TrimSpace(service),
			"plan":    strings.TrimSpace(plan),
			"expose":  strings.TrimSpace(expose),
			"compositionRef": map[string]interface{}{
				"name": strings.TrimSpace(composition),
			},
		},
	}

	return yaml.Marshal(manifest)
}

func buildKlutchPGServiceBindingManifest(name, namespace, instanceRef, instanceType, composition string) ([]byte, error) {
	manifest := map[string]interface{}{
		"apiVersion": klutchPGAPIVersion,
		"kind":       klutchPGServiceBindingKind,
		"metadata": map[string]interface{}{
			"name":      strings.TrimSpace(name),
			"namespace": strings.TrimSpace(namespace),
		},
		"spec": map[string]interface{}{
			"instanceRef":         strings.TrimSpace(instanceRef),
			"serviceInstanceType": strings.TrimSpace(instanceType),
			"compositionRef": map[string]interface{}{
				"name": strings.TrimSpace(composition),
			},
		},
	}

	return yaml.Marshal(manifest)
}

func buildKlutchPGBackupManifest(name, namespace, instanceRef, instanceType, composition string) ([]byte, error) {
	manifest := map[string]interface{}{
		"apiVersion": klutchPGAPIVersion,
		"kind":       klutchPGBackupKind,
		"metadata": map[string]interface{}{
			"name":      strings.TrimSpace(name),
			"namespace": strings.TrimSpace(namespace),
		},
		"spec": map[string]interface{}{
			"instanceRef":         strings.TrimSpace(instanceRef),
			"serviceInstanceType": strings.TrimSpace(instanceType),
			"compositionRef": map[string]interface{}{
				"name": strings.TrimSpace(composition),
			},
		},
	}

	return yaml.Marshal(manifest)
}

func buildKlutchPGRestoreManifest(name, namespace, backupRef, instanceRef, instanceType, composition string) ([]byte, error) {
	manifest := map[string]interface{}{
		"apiVersion": klutchPGAPIVersion,
		"kind":       klutchPGRestoreKind,
		"metadata": map[string]interface{}{
			"name":      strings.TrimSpace(name),
			"namespace": strings.TrimSpace(namespace),
		},
		"spec": map[string]interface{}{
			"backupRef":           strings.TrimSpace(backupRef),
			"instanceRef":         strings.TrimSpace(instanceRef),
			"serviceInstanceType": strings.TrimSpace(instanceType),
			"compositionRef": map[string]interface{}{
				"name": strings.TrimSpace(composition),
			},
		},
	}

	return yaml.Marshal(manifest)
}

func runKubectlWithInput(input []byte, args ...string) (string, error) {
	cmd := exec.Command("kubectl", args...)
	if len(input) > 0 {
		cmd.Stdin = bytes.NewBuffer(input)
	}

	makeup.PrintCommandBox(cmd.String())
	makeup.WaitForUser(demo.UnattendedMode)

	output, err := cmd.CombinedOutput()
	if makeup.Verbose {
		trimmed := strings.TrimSpace(string(output))
		if trimmed != "" {
			makeup.Print(trimmed)
		}
	}

	return string(output), err
}

func klutchResourceExists(resource, name, namespace string) (bool, error) {
	resourceName := fmt.Sprintf("%s/%s", strings.TrimSpace(resource), strings.TrimSpace(name))
	output, err := runKubectlWithInput(nil, "get", resourceName, "-n", strings.TrimSpace(namespace), "--ignore-not-found", "-o", "name")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) != "", nil
}

func init() {
	cmdCreateKlutchPGInstance.Flags().StringVar(&createKlutchPGInstanceName, "name", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service instance claim.")
	cmdCreateKlutchPGInstance.Flags().StringVarP(&createKlutchPGInstanceNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL service instance claim.")
	cmdCreateKlutchPGInstance.Flags().StringVar(&createKlutchPGInstanceService, "service", "a9s-postgresql13", "Service name for the Klutch PostgreSQL claim.")
	cmdCreateKlutchPGInstance.Flags().StringVar(&createKlutchPGInstancePlan, "plan", "postgresql-single-nano", "Plan name for the Klutch PostgreSQL claim.")
	cmdCreateKlutchPGInstance.Flags().StringVar(&createKlutchPGInstanceExpose, "expose", "Internal", "Exposure mode for the Klutch PostgreSQL claim.")
	cmdCreateKlutchPGInstance.Flags().StringVar(&createKlutchPGInstanceComposition, "composition", "a8s-postgresql", "Composition name for the Klutch PostgreSQL claim.")
	cmdCreateKlutchPGInstance.Flags().BoolVar(&createKlutchPGInstanceNoApply, "no-apply", false, "Render the manifest but do not apply it.")
	cmdCreateKlutchPGInstance.Flags().BoolVar(&createKlutchPGInstanceWait, "wait", true, "Wait for the Klutch PostgreSQL instance claim to become ready.")
	cmdCreateKlutchPGInstance.Flags().StringVar(&createKlutchPGInstanceWaitTimeout, "wait-timeout", "30m", "Timeout used with --wait.")

	cmdCreateKlutchPGServiceBinding.Flags().StringVar(&createKlutchPGServiceBindingName, "name", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service binding claim.")
	cmdCreateKlutchPGServiceBinding.Flags().StringVarP(&createKlutchPGServiceBindingNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL service binding claim.")
	cmdCreateKlutchPGServiceBinding.Flags().StringVarP(&createKlutchPGServiceBindingInstanceRef, "service-instance", "i", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service instance claim to bind to.")
	cmdCreateKlutchPGServiceBinding.Flags().StringVar(&createKlutchPGServiceBindingInstanceType, "service-instance-type", "postgresql", "Service instance type for the Klutch service binding claim.")
	cmdCreateKlutchPGServiceBinding.Flags().StringVar(&createKlutchPGServiceBindingComposition, "composition", "a8s-servicebinding", "Composition name for the Klutch service binding claim.")
	cmdCreateKlutchPGServiceBinding.Flags().BoolVar(&createKlutchPGServiceBindingNoApply, "no-apply", false, "Render the manifest but do not apply it.")
	cmdCreateKlutchPGServiceBinding.Flags().BoolVar(&createKlutchPGServiceBindingWait, "wait", true, "Wait for the Klutch service binding claim to become implemented.")
	cmdCreateKlutchPGServiceBinding.Flags().StringVar(&createKlutchPGServiceBindingWaitTimeout, "wait-timeout", "15m", "Timeout used with --wait.")

	cmdCreateKlutchPGBackup.Flags().StringVar(&createKlutchPGBackupName, "name", "example-a8s-postgresql", "Name of the Klutch PostgreSQL backup claim.")
	cmdCreateKlutchPGBackup.Flags().StringVarP(&createKlutchPGBackupNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL backup claim.")
	cmdCreateKlutchPGBackup.Flags().StringVarP(&createKlutchPGBackupInstanceRef, "service-instance", "i", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service instance claim to back up.")
	cmdCreateKlutchPGBackup.Flags().StringVar(&createKlutchPGBackupInstanceType, "service-instance-type", "postgresql", "Service instance type for the Klutch backup claim.")
	cmdCreateKlutchPGBackup.Flags().StringVar(&createKlutchPGBackupComposition, "composition", "a8s-backup", "Composition name for the Klutch backup claim.")
	cmdCreateKlutchPGBackup.Flags().BoolVar(&createKlutchPGBackupNoApply, "no-apply", false, "Render the manifest but do not apply it.")
	cmdCreateKlutchPGBackup.Flags().BoolVar(&createKlutchPGBackupWait, "wait", true, "Wait for the Klutch backup claim to become ready.")
	cmdCreateKlutchPGBackup.Flags().StringVar(&createKlutchPGBackupWaitTimeout, "wait-timeout", "30m", "Timeout used with --wait.")

	cmdCreateKlutchPGRestore.Flags().StringVar(&createKlutchPGRestoreName, "name", "example-a8s-postgresql", "Name of the Klutch PostgreSQL restore claim.")
	cmdCreateKlutchPGRestore.Flags().StringVarP(&createKlutchPGRestoreNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL restore claim.")
	cmdCreateKlutchPGRestore.Flags().StringVarP(&createKlutchPGRestoreBackupRef, "backup", "b", "example-a8s-postgresql-bu", "Name of the Klutch backup claim to restore.")
	cmdCreateKlutchPGRestore.Flags().StringVarP(&createKlutchPGRestoreInstanceRef, "service-instance", "i", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service instance claim to restore into.")
	cmdCreateKlutchPGRestore.Flags().StringVar(&createKlutchPGRestoreInstanceType, "service-instance-type", "postgresql", "Service instance type for the Klutch restore claim.")
	cmdCreateKlutchPGRestore.Flags().StringVar(&createKlutchPGRestoreComposition, "composition", "a8s-restore", "Composition name for the Klutch restore claim.")
	cmdCreateKlutchPGRestore.Flags().BoolVar(&createKlutchPGRestoreNoApply, "no-apply", false, "Render the manifest but do not apply it.")
	cmdCreateKlutchPGRestore.Flags().BoolVar(&createKlutchPGRestoreWait, "wait", true, "Wait for the Klutch restore claim to become ready.")
	cmdCreateKlutchPGRestore.Flags().StringVar(&createKlutchPGRestoreWaitTimeout, "wait-timeout", "30m", "Timeout used with --wait.")

	cmdCreateKlutchPG.AddCommand(cmdCreateKlutchPGInstance)
	cmdCreateKlutchPG.AddCommand(cmdCreateKlutchPGServiceBinding)
	cmdCreateKlutchPG.AddCommand(cmdCreateKlutchPGBackup)
	cmdCreateKlutchPG.AddCommand(cmdCreateKlutchPGRestore)
	cmdCreateKlutch.AddCommand(cmdCreateKlutchPG)

	cmdDeleteKlutchPGInstance.Flags().StringVar(&deleteKlutchPGInstanceName, "name", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service instance claim to delete.")
	cmdDeleteKlutchPGInstance.Flags().StringVarP(&deleteKlutchPGInstanceNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL service instance claim to delete.")
	cmdDeleteKlutchPGInstance.Flags().BoolVar(&deleteKlutchPGInstanceWait, "wait", false, "Wait for the Klutch PostgreSQL service instance claim to be deleted.")
	cmdDeleteKlutchPGInstance.Flags().StringVar(&deleteKlutchPGInstanceWaitTimeout, "wait-timeout", "15m", "Timeout used with --wait.")

	cmdDeleteKlutchPGServiceBinding.Flags().StringVar(&deleteKlutchPGServiceBindingName, "name", "example-a8s-postgresql", "Name of the Klutch PostgreSQL service binding claim to delete.")
	cmdDeleteKlutchPGServiceBinding.Flags().StringVarP(&deleteKlutchPGServiceBindingNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL service binding claim to delete.")
	cmdDeleteKlutchPGServiceBinding.Flags().BoolVar(&deleteKlutchPGServiceBindingWait, "wait", false, "Wait for the Klutch PostgreSQL service binding claim to be deleted.")
	cmdDeleteKlutchPGServiceBinding.Flags().StringVar(&deleteKlutchPGServiceBindingWaitTimeout, "wait-timeout", "15m", "Timeout used with --wait.")

	cmdDeleteKlutchPGBackup.Flags().StringVar(&deleteKlutchPGBackupName, "name", "example-a8s-postgresql-bu", "Name of the Klutch PostgreSQL backup claim to delete.")
	cmdDeleteKlutchPGBackup.Flags().StringVarP(&deleteKlutchPGBackupNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL backup claim to delete.")
	cmdDeleteKlutchPGBackup.Flags().BoolVar(&deleteKlutchPGBackupWait, "wait", false, "Wait for the Klutch PostgreSQL backup claim to be deleted.")
	cmdDeleteKlutchPGBackup.Flags().StringVar(&deleteKlutchPGBackupWaitTimeout, "wait-timeout", "15m", "Timeout used with --wait.")

	cmdDeleteKlutchPGRestore.Flags().StringVar(&deleteKlutchPGRestoreName, "name", "example-a8s-postgresql-rs", "Name of the Klutch PostgreSQL restore claim to delete.")
	cmdDeleteKlutchPGRestore.Flags().StringVarP(&deleteKlutchPGRestoreNamespace, "namespace", "n", "default", "Namespace of the Klutch PostgreSQL restore claim to delete.")
	cmdDeleteKlutchPGRestore.Flags().BoolVar(&deleteKlutchPGRestoreWait, "wait", false, "Wait for the Klutch PostgreSQL restore claim to be deleted.")
	cmdDeleteKlutchPGRestore.Flags().StringVar(&deleteKlutchPGRestoreWaitTimeout, "wait-timeout", "15m", "Timeout used with --wait.")

	cmdDeleteKlutchPG.AddCommand(cmdDeleteKlutchPGInstance)
	cmdDeleteKlutchPG.AddCommand(cmdDeleteKlutchPGServiceBinding)
	cmdDeleteKlutchPG.AddCommand(cmdDeleteKlutchPGBackup)
	cmdDeleteKlutchPG.AddCommand(cmdDeleteKlutchPGRestore)
	cmdDeleteKlutch.AddCommand(cmdDeleteKlutchPG)
}
