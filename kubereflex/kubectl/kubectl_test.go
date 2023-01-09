package kubectl

import (
	"flag"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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

func TestGetClusterInfo(t *testing.T) {
	objectKey := client.ObjectKey{Namespace: "cluster-registry", Name: "demo-active"}

	restClient := createCustomClient(objectKey.Namespace, getKubeConfig(""))

	clusterInfo := getClusterInfo(restClient, objectKey)
	if clusterInfo.secret.CreationTimestamp.IsZero() {
        t.Error("Failed to get secret")
    }
	if clusterInfo.cluster.CreationTimestamp.IsZero() {
        t.Error("Failed to get cluster")
    }
}

func TestRemove(t *testing.T) {
	// TODO: Ugly & disgusting
	CRDpath := "/home/kormi/Cisco/KLI/default_active_resource.yaml"
	
	cluster := getKubeConfig("")
	// cluster := getKubeConfig("cluster2.yaml")

	Apply(CRDpath, cluster)

	deploymentName, err := GetDeploymentName("icp-v115x", "istio-system", cluster)

	if deploymentName == "" && err != nil {
		t.Error("Custom resource not found after apply")
	}
	
	err = Remove(CRDpath, cluster)

	if err != nil {
		t.Error("Try to delete non-exist custom resource")
	}

	deploymentName, err = GetDeploymentName("icp-v115x", "istio-system", cluster)

	if deploymentName != "" && err == nil {
		t.Error("Custom resource not removed after delete")
	}
}
