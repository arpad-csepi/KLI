package helm

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"

	"github.com/arpad-csepi/KLI/kubereflex/kubectl"
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

var testChart = chartData{
	chartUrl:       "https://cisco-open.github.io/cluster-registry-controller",
	repositoryName: "cluster-registry",
	chartName:      "cluster-registry",
	releaseName:    "cluster-registry",
	namespace:      "cluster-registry",
	arguments:      nil,
	// arguments:      map[string]string{"set": "localCluster.name=demo-active,network.name=network1,controller.apiServerEndpointAddress=" + kubereflex.GetAPIServerEndpoint(&mainClusterConfigPath)},
}

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

func TestSetSettings(t *testing.T) {
	kubeconfig := getKubeConfig("")

	setSettings(testChart.namespace, kubeconfig)

	if os.Getenv("HELM_NAMESPACE") != testChart.namespace {
		t.Errorf("Helm namespace is incorrect")
	}
	if settings.Namespace() != testChart.namespace {
		t.Errorf("Kubernetes namespace is incorrect")
	}
	if settings.KubeConfig != *kubeconfig {
		t.Errorf("Kubeconfig is incorrect")
	}
}

func TestInstall(t *testing.T) {
	kubeconfig := getKubeConfig("")

	Install(testChart.repositoryName, testChart.chartName, testChart.releaseName, testChart.namespace, testChart.arguments, kubeconfig)

	// TODO: Set a fake client up before use GetDeploymentName
	t.Errorf("No client definied")
	deploymentName, err := kubectl.GetDeploymentName(testChart.releaseName, testChart.namespace)
	if err != nil {
		t.Errorf("Error when GetDeploymentName called: %s", err)
	}

	if deploymentName == "" {
		t.Errorf("Deployment not found after install")
	}
}

func TestUninstall(t *testing.T) {
	kubeconfig := getKubeConfig("")

	Uninstall(testChart.releaseName, testChart.namespace, kubeconfig)

	var err error
	// TODO: Set a fake client up before use GetDeploymentName
	t.Errorf("No client definied")
	testChart.deploymentName, err = kubectl.GetDeploymentName(testChart.releaseName, testChart.namespace)
	if err != nil {
		t.Errorf("Error when GetDeploymentName called: %s", err)
	}

	if testChart.deploymentName != "" {
		t.Errorf("Deployment found after uninstall")
	}
}

func TestIsRepositoryExists(t *testing.T) {
	RepositoryAdd(testChart.repositoryName, testChart.chartUrl)

	exists, err := IsRepositoryExists("cluster-registry")
	if exists != true {
		t.Errorf("This repository should exists at this point")
	}

	if err != nil {
		t.Errorf("Error when repository check called: %s", err)
	}

	exists, err = IsRepositoryExists("this-repository-a-bit-sus")
	if exists != false {
		t.Errorf("This repository should not exists")
	}

	if err != nil {
		t.Errorf("Error when repository check called %s", err)
	}

}

func TestRepositoryAdd(t *testing.T) {
	err := RepositoryAdd(testChart.repositoryName, testChart.chartUrl)
	if err != nil {
		t.Errorf("Error when RepositoryAdd called: %s", err)
	}

	err = RepositoryAdd("this-repository-a-bit-sus", "no-where")
	if err == nil {
		t.Errorf("Error when RepositoryAdd called: %s", err)
	}
}

func TestRepositoryUpdate(t *testing.T) {
	err := RepositoryUpdate()

	if err != nil {
		t.Errorf("Repository update failed: %s", err)
	}
}
