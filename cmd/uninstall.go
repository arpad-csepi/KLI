/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	kubereflex "github.com/arpad-csepi/kubereflex"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		switch resource_name, _ := cmd.Flags().GetString("resource"); resource_name {
		case "istio-operator":
			var (
				// chartName      = "istio-operator"
				releaseName = "banzaicloud-stable"
				namespace   = "istio-system"
			)

			kubereflex.UninstallHelmChart(releaseName, namespace)
		case "cluster-registry":
			var (
				// chartName      = "cluster-registry"
				releaseName = "cluster-registry"
				namespace   = "cluster-registry"
			)

			kubereflex.UninstallHelmChart(releaseName, namespace)
		default:
			fmt.Printf("Unknown resource")
		}
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// uninstallCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// uninstallCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	uninstallCmd.Flags().StringVarP(&resource, "resource", "r", "", "Resource is required")
	uninstallCmd.MarkFlagRequired("resource")
}
