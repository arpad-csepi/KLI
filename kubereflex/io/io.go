package io

import (
	"encoding/json"
	"io/fs"
	"os"

	banzaicloud "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	"sigs.k8s.io/yaml"
)

func fileRead(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		panic("Nope, file not found")
	}

	return data
}

func ReadYAMLResourceFile(path string) banzaicloud.IstioControlPlane {
	var data = fileRead(path)
	var controlPlane banzaicloud.IstioControlPlane
	err := yaml.Unmarshal(data, &controlPlane)
	if err != nil {
		panic("Aww, this resource file cannot convert to IstioControlPlane resource")
	}

	return controlPlane
}

func ReadJSONResourceFile(path string) banzaicloud.IstioControlPlane {
	var data = fileRead(path)
	var controlPlane banzaicloud.IstioControlPlane

	err := json.Unmarshal(data, &controlPlane)
	if err != nil {
		panic("Aww, this resource file cannot convert to IstioControlPlane resource")
	}

	return controlPlane
}

func ReadYAMLChartsFile(path string) {
	// TODO: Read chars from file in install command
}

func WriteObjectContent(controlplane banzaicloud.IstioControlPlane) {
	data, err := json.Marshal(controlplane)
	if err != nil {
		panic(err)
	}

	os.WriteFile("controlplane.json", data, fs.FileMode.Perm(0777))
}