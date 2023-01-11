package io

import (
	"os"

	istio_operator "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
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
