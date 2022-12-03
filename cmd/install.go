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

		// TODO: Read helm chart data from file
		clusterRegistry1 := chartData{
			chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
			repositoryName: "cluster-registry",
			chartName:      "cluster-registry",
			releaseName:    "cluster-registry",
			namespace:      "cluster-registry",
			arguments:      map[string]string{"set": "localCluster.name=demo-active,network.name=network1,controller.apiServerEndpointAddress=" + kubereflex.GetAPIServerEndpoint(&mainClusterConfigPath)},
		}

		clusterRegistry2 := chartData{
			chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
			repositoryName: "cluster-registry",
			chartName:      "cluster-registry",
			releaseName:    "cluster-registry",
			namespace:      "cluster-registry",
			arguments:      map[string]string{"set": "localCluster.name=demo-passive,network.name=network2,controller.apiServerEndpointAddress=" + kubereflex.GetAPIServerEndpoint(&secondaryClusterConfigPath)},
		}

		// Install cluster registry
		kubereflex.InstallHelmChart(clusterRegistry1.chartUrl,
			clusterRegistry1.repositoryName,
			clusterRegistry1.chartName,
			clusterRegistry1.releaseName,
			clusterRegistry1.namespace,
			clusterRegistry1.arguments,
			&mainClusterConfigPath)
		clusterRegistry1.deploymentName = kubereflex.GetDeploymentName(clusterRegistry1.releaseName, clusterRegistry1.namespace, &mainClusterConfigPath)

		if verify {
			kubereflex.Verify(clusterRegistry1.deploymentName, clusterRegistry1.namespace, &mainClusterConfigPath, time.Duration(timeout)*time.Second)
		}

		if secondaryClusterConfigPath != "" {
			// Install cluster registry
			kubereflex.InstallHelmChart(clusterRegistry2.chartUrl,
				clusterRegistry2.repositoryName,
				clusterRegistry2.chartName,
				clusterRegistry2.releaseName,
				clusterRegistry2.namespace,
				clusterRegistry2.arguments,
				&secondaryClusterConfigPath)
			clusterRegistry2.deploymentName = kubereflex.GetDeploymentName(clusterRegistry2.releaseName, clusterRegistry2.namespace, &secondaryClusterConfigPath)

			if verify {
				kubereflex.Verify(clusterRegistry2.deploymentName, clusterRegistry2.namespace, &secondaryClusterConfigPath, time.Duration(timeout)*time.Second)
			}
		}

		// TODO: Get namespace names from charts
		if attach {
			kubereflex.Attach(&mainClusterConfigPath, &secondaryClusterConfigPath, "cluster-registry", "cluster-registry")
		}

		istioOperator := chartData{
			chartUrl:       "https://kubernetes-charts.banzaicloud.com",
			repositoryName: "banzaicloud-stable",
			chartName:      "istio-operator",
			releaseName:    "banzaicloud-stable",
			namespace:      "istio-system",
			arguments:      map[string]string{"set": "clusterRegistry.clusterAPI.enabled=true,clusterRegistry.resourceSyncRules.enabled=true"},
		}

		// Install cluster registry
		kubereflex.InstallHelmChart(istioOperator.chartUrl,
			istioOperator.repositoryName,
			istioOperator.chartName,
			istioOperator.releaseName,
			istioOperator.namespace,
			istioOperator.arguments,
			&mainClusterConfigPath)
		istioOperator.deploymentName = kubereflex.GetDeploymentName(istioOperator.releaseName, istioOperator.namespace, &mainClusterConfigPath)

		if verify {
			kubereflex.Verify(istioOperator.deploymentName, istioOperator.namespace, &mainClusterConfigPath, time.Duration(timeout)*time.Second)
		}

		if activeCRDPath != "" {
			kubereflex.Apply(activeCRDPath, &mainClusterConfigPath)
		}

		if secondaryClusterConfigPath != "" {
			// Install cluster registry
			kubereflex.InstallHelmChart(istioOperator.chartUrl,
				istioOperator.repositoryName,
				istioOperator.chartName,
				istioOperator.releaseName,
				istioOperator.namespace,
				istioOperator.arguments,
				&secondaryClusterConfigPath)
			istioOperator.deploymentName = kubereflex.GetDeploymentName(istioOperator.releaseName, istioOperator.namespace, &secondaryClusterConfigPath)
			
			if verify {
				kubereflex.Verify(istioOperator.deploymentName, istioOperator.namespace, &secondaryClusterConfigPath, time.Duration(timeout)*time.Second)
			}
		}

		if passiveCRDPath != "" {
			kubereflex.Apply(passiveCRDPath, &secondaryClusterConfigPath)
		}
	},
}

var verify bool
var timeout int

var mainClusterConfigPath string
var secondaryClusterConfigPath string

var activeCRDPath string
var passiveCRDPath string

var attach bool

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

	installCmd.Flags().BoolVarP(&attach, "attach", "a", false, "Connect two cluster together")
	installCmd.Flags().StringVarP(&activeCRDPath, "active-custom-resource", "r", "", "Specify custom resource file location for the active cluster")
	installCmd.Flags().StringVarP(&passiveCRDPath, "passive-custom-resource", "R", "", "Specify custom resource file location for the passive cluster")
	installCmd.Flags().BoolVarP(&verify, "verify", "v", false, "Verify the deployment is ready or not")
	installCmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "Set verify timeout in seconds")
	installCmd.Flags().StringVarP(&mainClusterConfigPath, "main-cluster", "c", "", "Main cluster kubeconfig file location")
	installCmd.Flags().StringVarP(&secondaryClusterConfigPath, "secondary-cluster", "C", "", "Secondary cluster kubeconfig file location")
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
