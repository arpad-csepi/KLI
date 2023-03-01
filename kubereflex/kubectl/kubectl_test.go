package kubectl

import (
	"testing"

	testclient "k8s.io/client-go/kubernetes/fake"
)

var namespaceName string = "namespace-for-testing"

func createTestClient() {
	clientset = testclient.NewSimpleClientset()

	// // runtimeScheme contains already registered types in the API server
	// runtimeScheme := scheme.Scheme

	// // Add custom types to the runtime scheme
	// err := istio_operator.SchemeBuilder.AddToScheme(runtimeScheme)
	// if err != nil {
	// 	panic("Testclient add to scheme failed")
	// }

	// err = cluster_registry.SchemeBuilder.AddToScheme(runtimeScheme)
	// if err != nil {
	// 	panic("Testclient add to scheme failed")
	// }

	// mapper initializes a mapping between Kind and APIVersion to a resource name and back based on the objects in a runtime.Scheme and the Kubernetes API conventions.
	// mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(client.Discovery()))
}

func TestCreateNamespace(t *testing.T) {
	createTestClient()

	err := CreateNamespace(namespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNamespace(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(namespaceName)
	
	namespace, err := GetNamespace(namespaceName)
	if err != nil || namespace == nil {
		t.Error(err.Error())
	}
}

func TestDeleteNamespace(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(namespaceName)

	err := DeleteNamespace(namespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestIsNamespaceExists(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(namespaceName)

	exists, err := IsNamespaceExists(namespaceName)
	if err != nil || exists != true {
		t.Error(err.Error())
	}
}

func TestVerify(t *testing.T) {
	t.Fail()
}

func TestApply(t *testing.T) {
	t.Fail()
}

func TestRemove(t *testing.T) {
	// createTestClient()

	// // TODO: Ugly & disgusting
	// CRDpath := "/home/kormi/Cisco/KLI/default_active_resource.yaml"

	// Apply(CRDpath)

	// deploymentName, err := GetDeploymentName("icp-v115x", "istio-system")

	// if deploymentName == "" && err != nil {
	// 	t.Error("Custom resource not found after apply")
	// }

	// err = Remove(CRDpath)

	// if err != nil {
	// 	t.Error("Try to delete non-exist custom resource")
	// }

	// deploymentName, err = GetDeploymentName("icp-v115x", "istio-system")

	// if deploymentName != "" && err == nil {
	// 	t.Error("Custom resource not removed after delete")
	// }
	t.Fail()
}

func TestAPIServerEndpoint(t *testing.T) {
	createTestClient()

	endpoint, err := GetAPIServerEndpoint()
	if err != nil || endpoint == "" {
		t.Error(err.Error())
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
	// objectKey := client.ObjectKey{Namespace: "cluster-registry", Name: "demo-active"}

	// restClient, err := createCustomClient(objectKey.Namespace)
	// if err != nil {
	// 	t.Error(err)
	// }

	// clusterInfo, err := getClusterInfo(restClient, objectKey)
	// if err != nil {
	// 	t.Error(err)
	// }

	// if clusterInfo.secret.CreationTimestamp.IsZero() {
	// 	t.Error("Failed to get secret")
	// }
	// if clusterInfo.cluster.CreationTimestamp.IsZero() {
	// 	t.Error("Failed to get cluster")
	// }
	t.Fail()
}
