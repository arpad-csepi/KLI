package io

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	istio_operator "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func fileRead(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func ReadYAMLResourceFile(path string) (client.Object, error) {
	data, err := fileRead(path)
	if err != nil {
		return nil, err
	}

	var icp istio_operator.IstioControlPlane
	err = yaml.Unmarshal(data, &icp)

	if err != nil {
		return nil, err
	}
	return icp.DeepCopy(), nil
}

func GetClusterCRD(url string) (client.Object, error) {
	yamlData, err := fileDownload(url)
	if err != nil {
		return nil, err
	}

	var clusterCRD apiextensionsv1.CustomResourceDefinition
	err = yaml.Unmarshal(yamlData, &clusterCRD)

	if err != nil {
		return nil, err
	}
	return &clusterCRD, nil
}

func fileDownload(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("wrong http status code")
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return content, nil
}