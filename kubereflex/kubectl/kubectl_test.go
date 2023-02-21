package kubectl

import (
	"flag"
	"path/filepath"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
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

func TestCreateNamespace(t *testing.T) {
	namespaceName := "namespace-for-testing"
	var namespace *v1.Namespace
	timeout := 3 * time.Second
	kubeconfig := getKubeConfig("")
	var err error

	err = CreateNamespace(namespaceName, kubeconfig)
	if err != nil {
		t.Errorf("Error when CreateNamespace called: %s", err)
	}

	for start := time.Now(); ; {
		namespace, err = GetNamespace(namespaceName, kubeconfig)

		if namespace != nil {
			break
		}

		if time.Since(start) > timeout {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	if namespace == nil {
		t.Errorf("Namespace does not exists")
	}

	if err != nil {
		t.Errorf("Error when clientset get the namespace: %s", err)
	}

	DeleteNamespace(namespaceName, kubeconfig)
}

func TestGetNamespace(t *testing.T) {
	namespaceName := "namespace-for-testing"
	var namespace *v1.Namespace
	timeout := 3 * time.Second
	kubeconfig := getKubeConfig("")
	var err error

	err = CreateNamespace(namespaceName, kubeconfig)
	if err != nil {
		t.Errorf("Error when CreateNamespace called: %s", err)
	}

	for start := time.Now(); ; {
		namespace, err = GetNamespace(namespaceName, kubeconfig)

		if namespace != nil {
			break
		}

		if time.Since(start) > timeout {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	if namespace == nil {
		t.Errorf("Namespace does not exists")
	}

	if err != nil {
		t.Errorf("Error when clientset get the namespace: %s", err)
	}

	DeleteNamespace(namespaceName, kubeconfig)
}

func TestDeleteNamespace(t *testing.T) {
	namespaceName := "namespace-for-testing"
	var namespace *v1.Namespace
	timeout := 3 * time.Second
	kubeconfig := getKubeConfig("")
	var err error

	err = CreateNamespace(namespaceName, kubeconfig)
	if err != nil {
		t.Errorf("Error when CreateNamespace called: %s", err)
	}

	for start := time.Now(); ; {
		namespace, err = GetNamespace(namespaceName, kubeconfig)

		if namespace != nil {
			break
		}

		if time.Since(start) > timeout {
			break
		}

		time.Sleep(250 * time.Millisecond)
	}

	if namespace == nil {
		t.Errorf("Namespace does not exists")
	}

	if err != nil {
		t.Errorf("Error when clientset get the namespace: %s", err)
	}

	DeleteNamespace(namespaceName, kubeconfig)
}

func TestIsNamespaceExists(t *testing.T) {
	namespace := "namespace-for-testing"
	kubeconfig := getKubeConfig("")

	CreateNamespace(namespace, kubeconfig)

	exists, err := IsNamespaceExists(namespace, kubeconfig)
	if err != nil {
		t.Errorf("%s", err)
	}
	if exists != true {
		t.Errorf("Namespace should exists")
	}

}

func TestAPIServerEndpoint(t *testing.T) {
	kubeconfig := getKubeConfig("")

	endpoint, err := GetAPIServerEndpoint(kubeconfig)
	if err != nil {
		t.Errorf("%s", err)
	}
	if endpoint == "" {
		t.Errorf("Server endpoint should not empty")
	}
}

func TestGetDeploymentName(t *testing.T) {
	// Deploy before check
}

func TestAttach(t *testing.T) {
	// Install before attach
}

func TestDetach(t *testing.T) {
	// Install and attach before detach
}

func TestGetClusterInfo(t *testing.T) {
	objectKey := client.ObjectKey{Namespace: "cluster-registry", Name: "demo-active"}

	restClient, err := createCustomClient(objectKey.Namespace, getKubeConfig(""))
	if err != nil {
		t.Error(err)
	}

	clusterInfo, err := getClusterInfo(restClient, objectKey)
	if err != nil {
		t.Error(err)
	}

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
