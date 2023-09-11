/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/fischerjulian/a8s-demo/demo"
	"github.com/spf13/cobra"
)

var a8sPGProductName = "a8s Postgres"

// runDemoCmd represents the runDemo command
var a8sPGDemo = &cobra.Command{
	Use:   "a8s-pg-demo",
	Short: "Create a local demo environment for " + a8sPGProductName + " using a Kind Kubernetes cluster and installs",
	Long: `The demo assistent guides through the creation of Kind based Kubernetes cluster, 
	helps to install all necessary prerequisites within the cluster and finally configures and installes
	the operator.:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		demo.PrintWelcomeScreen()

		demo.EstablishConfigFilePath()

		if !demo.LoadConfig() {
			demo.EstablishWorkingDir()
		}

		demo.CheckPrerequisites()

		demo.WaitForUser()

		demo.CheckoutDeploymentGitRepository()

		if demo.CountPodsInDemoNamespace() == 0 {
			demo.PrintCheckmark("Kubernetes cluster has no pods in " + demo.GetConfig().DemoSpace + " namespace.")
		}

		demo.EstablishBackupStoreCredentials()

		demo.ApplyCertManagerManifests()

		demo.ApplyA8sManifests()

		demo.PrintDemoSummary()

	},
}

func init() {
	rootCmd.AddCommand(a8sPGDemo)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runDemoCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runDemoCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
