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