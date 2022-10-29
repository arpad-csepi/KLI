/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"flag"
	"path/filepath"

	"k8s.io/client-go/util/homedir"

	kubereflex "github.com/arpad-csepi/kubereflex"
)

// TODO: Fill up long description
// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a resource",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {

		switch resource_name, _ := cmd.Flags().GetString("resource"); resource_name {
		case "istio-operator":
			var (
				chartUrl       = "https://kubernetes-charts.banzaicloud.com"
				repositoryName = "banzaicloud-stable"
				chartName      = "istio-operator"
				releaseName    = "banzaicloud-stable"
				namespace      = "istio-system"
			)

			kubereflex.InstallHelmChart(chartUrl, repositoryName, chartName, releaseName, namespace, nil, getKubeConfig())
		case "cluster-registry":
			var (
				chartUrl       = "https://cisco-open.github.io/cluster-registry-controller"
				repositoryName = "cluster-registry"
				chartName      = "cluster-registry"
				releaseName    = "cluster-registry"
				namespace      = "cluster-registry"
			)

			kubereflex.InstallHelmChart(chartUrl, repositoryName, chartName, releaseName, namespace, nil, getKubeConfig())
		default:
			fmt.Printf("Unknown resource")
		}
	},
}

var resource string

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	//TODO: Maybe seperate kubernetes pod install and helm chart install here later

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	installCmd.Flags().StringVarP(&resource, "resource", "r", "", "Resource is required")
	installCmd.MarkFlagRequired("resource")
}

func getKubeConfig() *string {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return kubeconfig
}
