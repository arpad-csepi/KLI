/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/util/homedir"

	kubereflex "github.com/arpad-csepi/kubereflex"
)

type chartData struct {
	chartUrl string
	repositoryName string
	chartName string
	releaseName string
	namespace string
	arguments map[string]string
	deploymentName string
}

// TODO: Fill up long description
// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install istio-operator and cluster-registry-controller",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		var kubeconfig = getKubeConfig()

		charts := []chartData{}

		// TODO: Read helm chart data from file
		charts = append(charts, chartData {
			chartUrl : "https://kubernetes-charts.banzaicloud.com",
			repositoryName : "banzaicloud-stable",
			chartName : "istio-operator",
			releaseName : "banzaicloud-stable",
			namespace : "istio-system",
			arguments : nil,
		})
		charts = append(charts, chartData {
			chartUrl : "https://cisco-open.github.io/cluster-registry-controller",
			repositoryName : "cluster-registry",
			chartName : "cluster-registry",
			releaseName : "cluster-registry",
			namespace : "cluster-registry",
			arguments : nil,
		})

		// Install istio-operator and cluster-registry-controller automaticly
		for _, chart := range charts {
			// TODO: Get the deployment name from InstallHelmChart and store in chart.deploymentName

			kubereflex.InstallHelmChart(chart.chartUrl,
				chart.repositoryName,
				chart.chartName,
				chart.releaseName,
				chart.namespace,
				chart.arguments,
				kubeconfig)
			}

			charts[0].deploymentName = "banzaicloud-stable-istio-operator"
			charts[1].deploymentName = "cluster-registry-controller"

		if verify {
			for _, chart := range charts {
				kubereflex.Verify(chart.deploymentName, chart.namespace, kubeconfig, time.Duration(timeout) * time.Second)
			}
		}

		custom_resource_path, err := cmd.Flags().GetString("custom-resource")
		if err != nil {
			fmt.Printf("Custom resource error: %s", err.Error())
		}
		if custom_resource_path != "" {
			kubereflex.Apply(custom_resource_path, kubeconfig)
		}
	},
}

var custom_resource_path string
var verify bool
var timeout int

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

	installCmd.Flags().StringVarP(&custom_resource_path, "custom-resource", "c", "default_resource.yaml", "Specify custom resource file location")
	installCmd.Flags().BoolVarP(&verify, "verify", "v", false, "Verify the deployment is ready or not")
	installCmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "Set verify timeout in seconds")
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
