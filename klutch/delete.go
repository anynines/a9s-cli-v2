package klutch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/anynines/a9s-cli-v2/demo"
	"github.com/anynines/a9s-cli-v2/makeup"
	prereq "github.com/anynines/a9s-cli-v2/prerequisites"
)

func DeleteClusters() {
	demo.EstablishConfig()

	checkDeletePrerequisites()

	makeup.PrintH1("Are you sure you want to delete the Klutch clusters?")
	makeup.WaitForUser(demo.UnattendedMode)

	makeup.PrintInfo("Deleting Klutch clusters...")
	deleteManagementInfoFile(demo.DemoConfig.WorkingDir)
	deleteCluster(controlPlaneClusterName)
	deleteCluster(appClusterName)
}

// deleteCluster deletes a kind cluster with the given name.
func deleteCluster(name string) {
	cmd := exec.Command("kind", "delete", "cluster", "-n", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Could not delete cluster: %s", string(output)))
	}
	makeup.PrintCheckmark(fmt.Sprintf("Deleted cluster %s", name))
}

// deleteManagementInfoFile deletes the management info file in the configured working dir.
func deleteManagementInfoFile(workDir string) {
	path := filepath.Join(workDir, controlPlaneClusterInfoFilePath, controlPlaneClusterInfoFileName)

	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		makeup.ExitDueToFatalError(err, fmt.Sprintf("Unexpected error while deleting Control Plane Cluster info to file %s", path))
	}
}

// Checks if prerequisites of the delete command are met.
func checkDeletePrerequisites() {
	makeup.PrintH1("Checking Prerequisites...")

	commonTools := prereq.GetCommonRequiredTools()

	requiredTools := []prereq.RequiredTool{
		commonTools[prereq.ToolDocker],
		commonTools[prereq.ToolKind],
	}

	prereq.CheckRequiredTools(requiredTools)

	prereq.CheckDockerRunning()
}
