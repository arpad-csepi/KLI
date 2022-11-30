/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com
*/
package cmd

import (
	"github.com/arpad-csepi/KLI/kubereflex"
	"github.com/spf13/cobra"
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
		// Uninstall istio-operator and cluster-registry-controller automaticly
		kubereflex.UninstallHelmChart("banzaicloud-stable", "istio-system")
		kubereflex.UninstallHelmChart("cluster-registry", "cluster-registry")

		if activeCRDPath != "" {
			// TODO: Call kubereflex Delete function for delete the custom resource
			panic("Not implemented")
		}
		if passiveCRDPath != "" {
			// TODO: Call kubereflex Delete function for delete the custom resource
			panic("Not implemented")
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

	uninstallCmd.Flags().StringVarP(&activeCRDPath, "active-custom-resource", "a", "", "Specify custom resource file location for the active cluster")
	uninstallCmd.Flags().StringVarP(&passiveCRDPath, "passive-custom-resource", "p", "", "Specify custom resource file location for the passive cluster")
}
