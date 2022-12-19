package kubereflex

import (
	"time"

	"github.com/arpad-csepi/KLI/kubereflex/helm"
	"github.com/arpad-csepi/KLI/kubereflex/kubectl"
)

// TODO: Optional parameters like args, namespace (maybe in KLI pass nil parameter to here)
// TODO: Default value to namespace (maybe define defaults in KLI call this function)
func InstallHelmChart(chartUrl string, repositoryName string, chartName string, releaseName string, namespace string, args map[string]string, kubeconfig *string) {
	if !kubectl.IsNamespaceExists(namespace, kubeconfig) {
		kubectl.CreateNamespace(namespace, kubeconfig)
	}

	if !helm.IsRepositoryExists(repositoryName) {
		helm.RepositoryAdd(repositoryName, chartUrl)
	}

	helm.Install(repositoryName, chartName, releaseName, namespace, args, kubeconfig)
}

func UninstallHelmChart(releaseName string, namespace string, kubeconfig *string) {
	helm.Uninstall(releaseName, namespace, kubeconfig)
}

func GetDeploymentName(releaseName string, namespace string, kubeconfig *string) string {
	return kubectl.GetDeploymentName(releaseName, namespace, kubeconfig)
}

func Verify(deploymentName string, namespace string, kubeconfig *string, timeout time.Duration) {
	kubectl.Verify(deploymentName, namespace, kubeconfig, timeout)
}

func GetAPIServerEndpoint(kubeconfig *string) string {
	return kubectl.GetAPIServerEndpoint(kubeconfig)
}

func Apply(CRDpath string, kubeconfig *string) {
	kubectl.Apply(CRDpath, kubeconfig)
}

func Remove(CRDpath string, kubeconfig *string) {
	kubectl.Remove(CRDpath, kubeconfig)
}

func Attach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	kubectl.Attach(kubeconfig1, kubeconfig2, namespace1, namespace2)
}

func Detach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	kubectl.Detach(kubeconfig1, kubeconfig2, namespace1, namespace2)
}
