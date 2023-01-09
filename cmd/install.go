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

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install istio-operator and cluster-registry-controller",
	Long:  `Install command is create charts, install with helm package manager and configure depends on other parameters`,
	Run: func(cmd *cobra.Command, args []string) {
		if mainClusterConfigPath == "" {
			mainClusterConfigPath = *getKubeConfig()
		}

		clusterRegistry1 := chartData{
			chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
			repositoryName: "cluster-registry",
			chartName:      "cluster-registry",
			releaseName:    "cluster-registry",
			namespace:      "cluster-registry",
			arguments:      map[string]string{"set": "localCluster.name=demo-active,network.name=network1,controller.apiServerEndpointAddress=" + kubereflex.GetAPIServerEndpoint(&mainClusterConfigPath)},
		}
		
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
			clusterRegistry2 := chartData{
				chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
				repositoryName: "cluster-registry",
				chartName:      "cluster-registry",
				releaseName:    "cluster-registry",
				namespace:      "cluster-registry",
				arguments:      map[string]string{"set": "localCluster.name=demo-passive,network.name=network2,controller.apiServerEndpointAddress=" + kubereflex.GetAPIServerEndpoint(&secondaryClusterConfigPath)},
			}

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

	installCmd.Flags().BoolVarP(&attach, "attach", "a", false, "Connect two cluster together")
	installCmd.Flags().StringVarP(&activeCRDPath, "active-custom-resource", "r", "", "Specify custom resource file location for the active cluster")
	installCmd.Flags().StringVarP(&passiveCRDPath, "passive-custom-resource", "R", "", "Specify custom resource file location for the passive cluster")
	installCmd.Flags().BoolVarP(&verify, "verify", "v", false, "Verify the deployment is ready or not")
	installCmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "Set verify timeout in seconds")
	installCmd.Flags().StringVarP(&mainClusterConfigPath, "main-cluster", "c", "", "Main cluster kubeconfig file location")
	installCmd.Flags().StringVarP(&secondaryClusterConfigPath, "secondary-cluster", "C", "", "Secondary cluster kubeconfig file location")
}

// getKubeConfig is try to find default kube config in some default paths
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
