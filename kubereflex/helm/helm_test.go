package helm

import (
	"flag"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"

	"github.com/arpad-csepi/KLI/kubereflex/kubectl"
)

func getKubeConfig(config string) *string {
	// TODO: Ugly & disgusting path management
	var kubeconfig *string
	if home := homedir.HomeDir(); config == "" && home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, "Cisco", "KLI", config), "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return kubeconfig
}

func TestUninstall(t *testing.T) {
	// cluster := getKubeConfig("")
	cluster := getKubeConfig("cluster2.yaml")

	// repositoryName := "cluster-registry"
	// chartName := "cluster-registry"
	// args := map[string]string{}
	releaseName := "cluster-registry"
	namespace := "cluster-registry"

	// repositoryName := "banzaicloud-stable"
	// chartName := "istio-operator"
	// releaseName := "banzaicloud-stable"
	// namespace := "istio-system"
	// args := map[string]string{"set": "clusterRegistry.clusterAPI.enabled=true,clusterRegistry.resourceSyncRules.enabled=true"}
	// args := map[string]string{}

	// Install(repositoryName, chartName, releaseName, namespace, args, cluster)
	Uninstall(releaseName, namespace, cluster)
		
	deploymentName, err := kubectl.GetDeploymentName(releaseName, namespace, cluster)
	
	if deploymentName != "" && err == nil {
		t.Errorf("Deployment found after uninstall")
	}
}

func TestInstall(t *testing.T) {
	cluster := getKubeConfig("")
	// cluster := getKubeConfig("cluster2.yaml")

	repositoryName := "cluster-registry"
	chartName := "cluster-registry"
	releaseName := "cluster-registry"
	namespace := "cluster-registry"
	args := map[string]string{}
	// args := map[string]string{"set": "localCluster.name=demo-active,network.name=network1,controller.apiServerEndpointAddress=" + kubectl.GetAPIServerEndpoint(cluster)}
	// args := map[string]string{"set": "localCluster.name=demo-active,network.name=network1,controller.apiServerEndpointAddress=" + kubectl.GetAPIServerEndpoint(cluster)}

	// repositoryName := "banzaicloud-stable"
	// chartName := "istio-operator"
	// releaseName := "banzaicloud-stable"
	// namespace := "istio-system"
	// // args := map[string]string{"set": "clusterRegistry.clusterAPI.enabled=true,clusterRegistry.resourceSyncRules.enabled=true"}
	// args := map[string]string{}

	Install(repositoryName, chartName, releaseName, namespace, args, cluster)

	deploymentName, err := kubectl.GetDeploymentName(releaseName, namespace, cluster)
	
	if deploymentName == "" && err != nil {
		t.Errorf("Deployment not found after install")
	}
}