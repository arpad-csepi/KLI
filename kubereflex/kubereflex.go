package kubereflex

import (
	"time"

	"github.com/arpad-csepi/KLI/kubereflex/helm"
	"github.com/arpad-csepi/KLI/kubereflex/io"
	"github.com/arpad-csepi/KLI/kubereflex/kubectl"

	"github.com/manifoldco/promptui"
)

var usedContexts = []string{}

func ChooseContextFromConfig(kubeconfig *string) string {
	contexts, err := io.GetContextsFromConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	notUsedContexts := []string{}

	for _, context := range contexts {
		usedFound := false

		for _, usedContext := range usedContexts {
			if context == usedContext {
				usedFound = true
			}
		}

		if !usedFound {
			notUsedContexts = append(notUsedContexts, context)
		}
	}

	selectedItem := ""

	if len(notUsedContexts) == 0 {
		panic("no more unused context remained")
	}

	if len(notUsedContexts) > 1 {
		prompt := promptui.Select{
			Label: "Select context for the cluster",
			Items: notUsedContexts,
		}
		_, selectedItem, err = prompt.Run()
		if err != nil {
			panic(err)
		}
	} else {
		selectedItem = notUsedContexts[0]
	}

	usedContexts = append(usedContexts, selectedItem)

	return selectedItem
}

func InstallHelmChart(chartUrl string, repositoryName string, chartName string, releaseName string, namespace string, args map[string]string, kubeconfig *string, context string) {
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

	err = helm.Install(repositoryName, chartName, releaseName, namespace, args, kubeconfig, context)
	if err != nil {
		panic(err)
	}
}

func UninstallHelmChart(releaseName string, namespace string, kubeconfig *string, context string) {
	err := helm.Uninstall(releaseName, namespace, kubeconfig, context)
	if err != nil {
		panic(err)
	}
}

func GetDeploymentName(releaseName string, namespace string, kubeconfig *string, context string) string {
	clientConfig := map[string]string{
		"kubeconfig": *kubeconfig,
		"context":    context,
	}

	err := kubectl.CreateClient(clientConfig)
	if err != nil {
		panic(err)
	}

	deploymentName, err := kubectl.GetDeploymentName(releaseName, namespace)
	if err != nil {
		panic(err.Error())
	}

	kubectl.RemoveAllClients()

	return deploymentName
}

func Verify(deploymentName string, namespace string, kubeconfig *string, context string, timeout time.Duration) {
	clientConfig := map[string]string{
		"kubeconfig": *kubeconfig,
		"context":    context,
	}

	err := kubectl.CreateClient(clientConfig)
	if err != nil {
		panic(err)
	}

	err = kubectl.Verify(deploymentName, namespace, timeout)
	if err != nil {
		panic(err)
	}

	kubectl.RemoveAllClients()
}

func GetAPIServerEndpoint(kubeconfig *string, context string) string {
	clientConfig := map[string]string{
		"kubeconfig": *kubeconfig,
		"context":    context,
	}

	err := kubectl.CreateClient(clientConfig)
	if err != nil {
		panic(err)
	}

	endpoint, err := kubectl.GetAPIServerEndpoint()
	if err != nil {
		panic(err)
	}

	kubectl.RemoveAllClients()

	return endpoint
}

func Apply(CRDPath string, kubeconfig *string, context string) {
	clientConfig := map[string]string{
		"kubeconfig": *kubeconfig,
		"context":    context,
	}

	err := kubectl.CreateClient(clientConfig)
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

	kubectl.RemoveAllClients()
}

func Remove(CRDPath string, kubeconfig *string, context string) {
	clientConfig := map[string]string{
		"kubeconfig": *kubeconfig,
		"context":    context,
	}

	err := kubectl.CreateClient(clientConfig)
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

	kubectl.RemoveAllClients()
}

func Attach(kubeconfig1 *string, context1 string, kubeconfig2 *string, context2 string, namespace1 string, namespace2 string) {
	clientConfig1 := map[string]string{
		"kubeconfig": *kubeconfig1,
		"context":    context1,
	}
	clientConfig2 := map[string]string{
		"kubeconfig": *kubeconfig2,
		"context":    context2,
	}

	err := kubectl.CreateClient(clientConfig1, clientConfig2)
	if err != nil {
		panic(err)
	}
	err = kubectl.Attach(namespace1, namespace2)
	if err != nil {
		panic(err)
	}

	kubectl.RemoveAllClients()
}

func Detach(kubeconfig1 *string, context1 string, kubeconfig2 *string, context2 string, namespace1 string, namespace2 string) {
	clientConfig1 := map[string]string{
		"kubeconfig": *kubeconfig1,
		"context":    context1,
	}
	clientConfig2 := map[string]string{
		"kubeconfig": *kubeconfig2,
		"context":    context2,
	}

	err := kubectl.CreateClient(clientConfig1, clientConfig2)
	if err != nil {
		panic(err)
	}
	err = kubectl.Detach(namespace1, namespace2)
	if err != nil {
		panic(err)
	}

	kubectl.RemoveAllClients()
}
