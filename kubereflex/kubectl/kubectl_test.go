package kubectl

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var namespaceTestName string = "namespace-for-testing"
var deploymentTestName = "deployment-for-testing"

func createTestClient() {
	Clientset = testclient.NewSimpleClientset()

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

	err := CreateNamespace(namespaceTestName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNamespace(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(namespaceTestName)

	namespace, err := GetNamespace(namespaceTestName)
	if err != nil || namespace == nil {
		t.Error(err.Error())
	}
}

func TestDeleteNamespace(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(namespaceTestName)

	err := DeleteNamespace(namespaceTestName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestIsNamespaceExists(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(namespaceTestName)

	exists, err := IsNamespaceExists(namespaceTestName)
	if err != nil || exists != true {
		t.Error(err.Error())
	}
}

func TestVerify(t *testing.T) {
	createTestClient()

	testTimeout := 2 * time.Second

	deploymentData := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentTestName,
			Namespace:   namespaceTestName,
		},
		Status:     appsv1.DeploymentStatus{
			Replicas:            3,
			ReadyReplicas:       3,
		},
	}

	_, err := Clientset.AppsV1().Deployments(namespaceTestName).Create(context.TODO(),
		&deploymentData, metav1.CreateOptions{})
	if err != nil {
		t.Error(err.Error())
	}

	err = Verify(deploymentTestName, namespaceTestName, testTimeout)
	if err != nil {
		t.Error(err.Error())
	}
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
	createTestClient()

	deploymentTestReleaseName := "annotation-for-testing"
	deploymentTestAnnotations := map[string]string{"deploymentTestReleaseName": deploymentTestReleaseName}

	deploymentData := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        deploymentTestName,
			Namespace:   namespaceTestName,
			Annotations: deploymentTestAnnotations,
		},
	}

	_, err := Clientset.AppsV1().Deployments(namespaceTestName).Create(context.TODO(), &deploymentData, metav1.CreateOptions{})
	if err != nil {
		t.Error(err.Error())
	}

	deploymentName, err := GetDeploymentName(deploymentTestReleaseName, namespaceTestName)
	if err != nil {
		t.Error(err.Error())
	}

	if deploymentTestName != deploymentName {
		t.Error("Deployment name is wrong!")
	}
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
