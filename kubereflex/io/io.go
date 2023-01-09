package io

import (
	"os"

	istio_operator "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

func fileRead(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("Nope, file not found")
	}

	return data
}

func ReadYAMLResourceFile(path string) client.Object {
	var data = fileRead(path)

	var icp istio_operator.IstioControlPlane
	err := yaml.Unmarshal(data, &icp)

	if err != nil {
		panic(err)
	}
	return icp.DeepCopy()
}