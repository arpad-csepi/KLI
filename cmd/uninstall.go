/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com
*/
package cmd

import (
	"fmt"

	"github.com/arpad-csepi/KLI/kubereflex"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall istio-operator and cluster-registry-controller",
	Long: "Uninstall command is uninstall charts deployment with helm package manager and clean-up depends on other parameters",
	Run: func(_ *cobra.Command, _ []string) {
		if mainClusterConfigPath == "" {
			mainClusterConfigPath = *getKubeConfig()
		}

		if mainContext == "" {
			fmt.Println("Main cluster context switcher:")
			mainContext = kubereflex.ChooseContextFromConfig(&mainClusterConfigPath)
		}

		if activeCRDPath != "" {
			kubereflex.Remove(activeCRDPath, &mainClusterConfigPath, mainContext)
		}

		kubereflex.UninstallHelmChart("cluster-registry", "cluster-registry", &mainClusterConfigPath, mainContext)
		kubereflex.UninstallHelmChart("banzaicloud-stable", "istio-system", &mainClusterConfigPath, mainContext)

		if secondaryClusterConfigPath == "" {
			secondaryClusterConfigPath = mainClusterConfigPath
		}
		if secondaryClusterConfigPath != "" {
			if secondaryContext == "" {
				fmt.Println("Main cluster context switcher:")
				secondaryContext = kubereflex.ChooseContextFromConfig(&secondaryClusterConfigPath)
			}
			
			if passiveCRDPath != "" {
				kubereflex.Remove(passiveCRDPath, &secondaryClusterConfigPath, secondaryContext)
			}

			kubereflex.UninstallHelmChart("cluster-registry", "cluster-registry", &secondaryClusterConfigPath, secondaryContext)
			kubereflex.UninstallHelmChart("banzaicloud-stable", "istio-system", &secondaryClusterConfigPath, secondaryContext)

			if detach {
				kubereflex.Detach(&mainClusterConfigPath, mainContext, &secondaryClusterConfigPath, secondaryContext, "cluster-registry", "cluster-registry")
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

	installCmd.Flags().StringVarP(&mainContext, "main-context", "k", "", "Main cluster context name")
	installCmd.Flags().StringVarP(&secondaryContext, "secondary-context", "K", "", "Secondary cluster context name")

	uninstallCmd.Flags().BoolVarP(&detach, "detach", "d", false, "Remove cluster connections")
}
