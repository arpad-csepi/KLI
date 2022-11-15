/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com

*/
package cmd

import (
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
		// Uninstall istio-operator and cluster-registry-controller automaticly
		kubereflex.UninstallHelmChart("banzaicloud-stable", "istio-system")
		kubereflex.UninstallHelmChart("cluster-registry", "cluster-registry")
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
	uninstallCmd.Flags().StringVarP(&custom_resource_path, "custom_resource_path", "c", "", "Specify custom resource file location")
}
