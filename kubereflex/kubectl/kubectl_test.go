package kubectl

import (
	"path/filepath"
	"testing"
	"time"

	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
	"github.com/arpad-csepi/KLI/kubereflex/io"
)

var testNamespaceName string = "namespace-for-testing"
var testDeploymentName = "deployment-for-testing"
var testDeploymentReleaseName = "release-name-for-testing"
var testDeploymentAnnotations = map[string]string{"deploymentTestReleaseName": testDeploymentReleaseName}

var objectKey1 = client.ObjectKey{Namespace: testNamespaceName, Name: "demo-active"}

// var objectKey2 = client.ObjectKey{Namespace: testNamespaceName, Name: "demo-passive"}

var testContainer = &corev1.Container{
	Name:  "test-container",
	Image: "k8s.gcr.io/test-webserver",
}

var testDeployment = appsv1.Deployment{
	TypeMeta:   metav1.TypeMeta{},
	ObjectMeta: metav1.ObjectMeta{Name: testDeploymentName, Namespace: testNamespaceName, Annotations: testDeploymentAnnotations},
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": testDeploymentName}},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Name: testNamespaceName, Namespace: testNamespaceName, Labels: map[string]string{"app": testDeploymentName}},
			Spec:       corev1.PodSpec{Containers: []corev1.Container{*testContainer}}},
	},
	Status: appsv1.DeploymentStatus{Replicas: 3, ReadyReplicas: 3},
}

func createTestClient() {
	var err error
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	err = CreateClient(&kubeconfig)

	if err != nil {
		panic(err)
	}

	backupClusterState()
}

func backupClusterState() {}

func restoreClusterState() {}

func TestCreateNamespace(t *testing.T) {
	createTestClient()

	err := CreateNamespace(testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNamespace(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(testNamespaceName)

	namespace, err := GetNamespace(testNamespaceName)
	if err != nil || namespace == nil || namespace.Name != testNamespaceName {
		t.Error(err.Error())
	}

	restoreClusterState()
}

func TestDeleteNamespace(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(testNamespaceName)

	err := DeleteNamespace(testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}

	restoreClusterState()
}

func TestIsNamespaceExists(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(testNamespaceName)

	exists, err := IsNamespaceExists(testNamespaceName)
	if err != nil || exists != true {
		t.Error(err.Error())
	}

	restoreClusterState()
}

func TestVerify(t *testing.T) {
	createTestClient()

	testTimeout := 2 * time.Second

	_ = Apply(&testDeployment)

	err := Verify(testDeploymentName, testNamespaceName, testTimeout)
	if err != nil {
		t.Error(err.Error())
	}

	restoreClusterState()
}

func TestApply(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(testNamespaceName)

	err := Apply(&testDeployment)
	if err != nil {
		t.Error(err)
	}

	restoreClusterState()
}

func TestRemove(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(&testDeployment)

	err := Remove(&testDeployment)

	if err != nil {
		t.Error("Try to delete non-exist custom resource")
	}

	restoreClusterState()
}

func TestAPIServerEndpoint(t *testing.T) {
	createTestClient()

	endpoint, err := GetAPIServerEndpoint()
	if err != nil || endpoint == "" {
		t.Error(err.Error())
	}

	restoreClusterState()
}

func TestGetDeploymentName(t *testing.T) {
	createTestClient()

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(&testDeployment)

	deploymentName, err := GetDeploymentName(testDeploymentReleaseName, testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}

	if testDeploymentName != deploymentName {
		t.Error("Deployment name is wrong!")
	}

	restoreClusterState()
}

func TestAttach(t *testing.T) {
	// Install before attach

	restoreClusterState()
	t.Fail()
}

func TestDetach(t *testing.T) {
	// Install and attach before detach

	restoreClusterState()
	t.Fail()
}

func TestGetClusterInfo(t *testing.T) {
	createTestClient()
	Clients[0].client = client.NewNamespacedClient(Clients[0].client, testNamespaceName)

	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectKey1.Name,
			Namespace: objectKey1.Namespace,
		},
	}

	cluster1 := &cluster_registry.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      objectKey1.Name,
			Namespace: objectKey1.Namespace,
		},
	}

	url := "https://raw.githubusercontent.com/cisco-open/cluster-registry-controller/cb563ec383a6a98f8d8e5c79d3350997b7e70075/deploy/charts/cluster-registry/crds/clusterregistry.k8s.cisco.com_clusters.yaml"
	clusterCRD, err := io.GetClusterCRD(url)
	if err != nil {
		t.Error(err)
	}

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(secret1)
	_ = Apply(clusterCRD)
	_ = Apply(cluster1)

	clusterInfo, err := getClusterInfo(Clients[0].client, objectKey1)
	if err != nil {
		t.Error(err)
	}

	if clusterInfo.secret.Name != secret1.Name || clusterInfo.cluster.Name != cluster1.Name {
		t.Error("wrong resource name")
	}

	restoreClusterState()
}