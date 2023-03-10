package kubereflex

import (
	"time"

	"github.com/arpad-csepi/KLI/kubereflex/helm"
	"github.com/arpad-csepi/KLI/kubereflex/io"
	"github.com/arpad-csepi/KLI/kubereflex/kubectl"
)

func InstallHelmChart(chartUrl string, repositoryName string, chartName string, releaseName string, namespace string, args map[string]string, kubeconfig *string) {
	err := kubectl.CreateClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	isNamespaceExists, err := kubectl.IsNamespaceExists(namespace)
	if err != nil {
		panic(err)
	}
	if !isNamespaceExists {
		err := kubectl.CreateNamespace(namespace)
		if err != nil {
			panic(err)
		}
	}

	isRepositoryExists, err := helm.IsRepositoryExists(repositoryName)
	if err != nil {
		panic(err)
	}
	if !isRepositoryExists {
		err := helm.RepositoryAdd(repositoryName, chartUrl)
		if err != nil {
			panic(err)
		}
	}

	err = helm.Install(repositoryName, chartName, releaseName, namespace, args, kubeconfig)
	if err != nil {
		panic(err)
	}
}

func UninstallHelmChart(releaseName string, namespace string, kubeconfig *string) {
	err := helm.Uninstall(releaseName, namespace, kubeconfig)
	if err != nil {
		panic(err)
	}
}

func GetDeploymentName(releaseName string, namespace string, kubeconfig *string) string {
	err := kubectl.CreateClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	deploymentName, err := kubectl.GetDeploymentName(releaseName, namespace)
	if err != nil {
		panic(err.Error())
	}
	return deploymentName
}

func Verify(deploymentName string, namespace string, kubeconfig *string, timeout time.Duration) {
	err := kubectl.CreateClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	err = kubectl.Verify(deploymentName, namespace, timeout)
	if err != nil {
		panic(err)
	}
}

func GetAPIServerEndpoint(kubeconfig *string) string {
	// err := kubectl.CreateClient(kubeconfig)
	// if err != nil {
	// 	panic(err)
	// }

	// endpoint, err := kubectl.GetAPIServerEndpoint()
	// if err != nil {
	// 	panic(err)
	// }
	// return endpoint

	return "todo"
}

func Apply(CRDPath string, kubeconfig *string) {
	err := kubectl.CreateClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	CRObject, err := io.ReadYAMLResourceFile(CRDPath)
	if err != nil {
		panic(err)
	}

	err = kubectl.Apply(CRObject)
	if err != nil {
		panic(err)
	}
}

func Remove(CRDPath string, kubeconfig *string) {
	err := kubectl.CreateClient(kubeconfig)
	if err != nil {
		panic(err)
	}

	CRObject, err := io.ReadYAMLResourceFile(CRDPath)
	if err != nil {
		panic(err)
	}

	err = kubectl.Remove(CRObject)
	if err != nil {
		panic(err)
	}
}

func Attach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	err := kubectl.CreateClient(kubeconfig1, kubeconfig2)
	if err != nil {
		panic(err)
	}
	err = kubectl.Attach(namespace1, namespace2)
	if err != nil {
		panic(err)
	}
}

func Detach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	err := kubectl.CreateClient(kubeconfig1, kubeconfig2)
	if err != nil {
		panic(err)
	}
	err = kubectl.Detach(namespace1, namespace2)
	if err != nil {
		panic(err)
	}
}
