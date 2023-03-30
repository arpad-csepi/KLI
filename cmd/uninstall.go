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
	Short: "Uninstall istio-operator and cluster-registry-controller",
	Long: "Uninstall command is uninstall charts deployment with helm package manager and clean-up depends on other parameters",
	Run: func(cmd *cobra.Command, args []string) {
		if mainClusterConfigPath == "" {
			mainClusterConfigPath = *getKubeConfig()
		}

		

		if activeCRDPath != "" {
			kubereflex.Remove(activeCRDPath, &mainClusterConfigPath, "TODO")
		}

		kubereflex.UninstallHelmChart("cluster-registry", "cluster-registry", &mainClusterConfigPath)
		kubereflex.UninstallHelmChart("banzaicloud-stable", "istio-system", &mainClusterConfigPath)

		if secondaryClusterConfigPath != "" {
			if passiveCRDPath != "" {
				kubereflex.Remove(passiveCRDPath, &secondaryClusterConfigPath, "TODO")
			}

			kubereflex.UninstallHelmChart("cluster-registry", "cluster-registry", &secondaryClusterConfigPath)
			kubereflex.UninstallHelmChart("banzaicloud-stable", "istio-system", &secondaryClusterConfigPath)

			if detach {
				kubereflex.Detach(&mainClusterConfigPath, "TODO", &secondaryClusterConfigPath, "TODO", "cluster-registry", "cluster-registry")
			}
		}

		
	},
}

var detach bool

func init() {
	rootCmd.AddCommand(uninstallCmd)

	uninstallCmd.Flags().StringVarP(&activeCRDPath, "active-custom-resource", "r", "", "Specify custom resource file location for the active cluster")
	uninstallCmd.Flags().StringVarP(&passiveCRDPath, "passive-custom-resource", "R", "", "Specify custom resource file location for the passive cluster")

	uninstallCmd.Flags().StringVarP(&mainClusterConfigPath, "main-cluster", "c", "", "Main cluster kubeconfig file location")
	uninstallCmd.Flags().StringVarP(&secondaryClusterConfigPath, "secondary-cluster", "C", "", "Secondary cluster kubeconfig file location")

	uninstallCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Remove cluster connections")
}
