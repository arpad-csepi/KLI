package io

import (
	"os"
	"testing"
)

func TestReadYAMLResourceFile(t *testing.T) {
	yaml_content := "kind: IstioControlPlane\nmetadata:\n  name: icp-v115x\n  namespace: istio-system"


	err := os.WriteFile("test_resource.yaml", []byte(yaml_content), 0755)
    if err != nil {
        t.Errorf("Unable to write file: %v", err)
    }

	object, err := ReadYAMLResourceFile("test_resource.yaml")
	if err != nil {
		t.Errorf(err.Error())
	}

	if object == nil {
		t.Errorf(err.Error())
	}

	if object.GetName() == "" || object.GetNamespace() == "" || object.GetObjectKind() == nil {
		t.Errorf("Object not properly converted")
	}

	os.Remove("test_resource.yaml")
}

func TestGetClusterCRD(t *testing.T) {
	url := "https://raw.githubusercontent.com/cisco-open/cluster-registry-controller/cb563ec383a6a98f8d8e5c79d3350997b7e70075/deploy/charts/cluster-registry/crds/clusterregistry.k8s.cisco.com_clusters.yaml"
	clusterCRD, err := GetClusterCRD(url)
	if err != nil {
		t.Error(err)
	}

	kind := clusterCRD.GetObjectKind().GroupVersionKind().Kind

	if kind != "CustomResourceDefinition" {
		t.Error("crd kind(a) bad")
	}
}