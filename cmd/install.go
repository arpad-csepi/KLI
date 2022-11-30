/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com
*/
package cmd

import (
	"github.com/arpad-csepi/KLI/kubereflex"

	"github.com/spf13/cobra"

	"flag"
	"path/filepath"
	"time"

	"k8s.io/client-go/util/homedir"
)

type chartData struct {
	chartUrl       string
	repositoryName string
	chartName      string
	releaseName    string
	namespace      string
	arguments      map[string]string
	deploymentName string
}

// TODO: Fill up long description
// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install istio-operator and cluster-registry-controller",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		if mainClusterConfigPath == "" {
			mainClusterConfigPath = *getKubeConfig()
		}

		var activeCharts = []chartData{}

		// TODO: Read helm chart data from file
		activeCharts = append(activeCharts, chartData{
			chartUrl:       "https://kubernetes-charts.banzaicloud.com",
			repositoryName: "banzaicloud-stable",
			chartName:      "istio-operator",
			releaseName:    "banzaicloud-stable",
			namespace:      "istio-system",
			arguments:      nil,
		})
		activeCharts = append(activeCharts, chartData{
			chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
			repositoryName: "cluster-registry",
			chartName:      "cluster-registry",
			releaseName:    "cluster-registry",
			namespace:      "cluster-registry",
			arguments:      map[string]string{"set": "localCluster.name=demo-active,network.name=network1,controller.apiServerEndpointAddress="+kubereflex.GetAPIServerEndpoint(&mainClusterConfigPath)},

		})

		// Install istio-operator and cluster-registry-controller automaticly
		for index, chart := range activeCharts {
			kubereflex.InstallHelmChart(chart.chartUrl,
				chart.repositoryName,
				chart.chartName,
				chart.releaseName,
				chart.namespace,
				chart.arguments,
				&mainClusterConfigPath)
			activeCharts[index].deploymentName = kubereflex.GetDeploymentName(chart.releaseName, chart.namespace, &mainClusterConfigPath)
		}

		if verify {
			for _, chart := range activeCharts {
				kubereflex.Verify(chart.deploymentName, chart.namespace, &mainClusterConfigPath, time.Duration(timeout)*time.Second)
			}
		}

		if activeCRDPath != "" {
			kubereflex.Apply(activeCRDPath, &mainClusterConfigPath)
		}

		if secondaryClusterConfigPath != "" {
			var passiveCharts = []chartData{}

			passiveCharts = append(activeCharts, chartData{
				chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
				repositoryName: "cluster-registry",
				chartName:      "cluster-registry",
				releaseName:    "cluster-registry",
				namespace:      "cluster-registry",
				arguments:      map[string]string{"set": "localCluster.name=demo-passive,network.name=network2,controller.apiServerEndpointAddress="+kubereflex.GetAPIServerEndpoint(&mainClusterConfigPath)},
			})

			for index, chart := range passiveCharts {
				kubereflex.InstallHelmChart(chart.chartUrl,
					chart.repositoryName,
					chart.chartName,
					chart.releaseName,
					chart.namespace,
					chart.arguments,
					&secondaryClusterConfigPath)
					passiveCharts[index].deploymentName = kubereflex.GetDeploymentName(chart.releaseName, chart.namespace, &secondaryClusterConfigPath)
			}

			if verify {
				for _, chart := range passiveCharts {
					kubereflex.Verify(chart.deploymentName, chart.namespace, &secondaryClusterConfigPath, time.Duration(timeout)*time.Second)
				}
			}

			if passiveCRDPath != "" {
				kubereflex.Apply(passiveCRDPath, &secondaryClusterConfigPath)
			}
		}
	},
}

var verify bool
var timeout int

var mainClusterConfigPath string
var secondaryClusterConfigPath string

var activeCRDPath string
var passiveCRDPath string

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

	installCmd.Flags().StringVarP(&activeCRDPath, "active-custom-resource", "a", "", "Specify custom resource file location for the active cluster")
	installCmd.Flags().StringVarP(&passiveCRDPath, "passive-custom-resource", "p", "", "Specify custom resource file location for the passive cluster")
	installCmd.Flags().BoolVarP(&verify, "verify", "v", false, "Verify the deployment is ready or not")
	installCmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "Set verify timeout in seconds")
	installCmd.Flags().StringVarP(&mainClusterConfigPath, "main-cluster", "m", "", "Main cluster kubeconfig file location")
	installCmd.Flags().StringVarP(&secondaryClusterConfigPath, "secondary-cluster", "s", "", "Secondary cluster kubeconfig file location")
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
