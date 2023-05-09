package kubectl

import (
	"context"
	"fmt"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arpad-csepi/KLI/kubereflex/io"

	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
)

var testNamespaceName string = "namespace-for-testing"
var testDeploymentName = "deployment-for-testing"
var testDeploymentReleaseName = "release-name-for-testing"
var testDeploymentAnnotations = map[string]string{"deploymentTestReleaseName": testDeploymentReleaseName}

var objectKey1 = client.ObjectKey{Namespace: testNamespaceName, Name: "demo-active"}
var objectKey2 = client.ObjectKey{Namespace: testNamespaceName, Name: "demo-passive"}

var testSecret1 = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      objectKey1.Name,
		Namespace: objectKey1.Namespace,
	},
}
var testSecret2 = &corev1.Secret{
	ObjectMeta: metav1.ObjectMeta{
		Name:      objectKey2.Name,
		Namespace: objectKey2.Namespace,
	},
}

var testCluster1 = &cluster_registry.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: objectKey1.Name,
	},
}
var testCluster2 = &cluster_registry.Cluster{
	ObjectMeta: metav1.ObjectMeta{
		Name: objectKey2.Name,
	},
}

var testContainer = &corev1.Container{
	Name:  "test-container",
	Image: "k8s.gcr.io/test-webserver",
}

var testDeployment = appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:        testDeploymentName,
		Namespace:   testNamespaceName,
		Annotations: testDeploymentAnnotations,
	},
	Spec: appsv1.DeploymentSpec{
		Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": testDeploymentName}},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testNamespaceName,
				Namespace: testNamespaceName,
				Labels:    map[string]string{"app": testDeploymentName},
			},
			Spec: corev1.PodSpec{Containers: []corev1.Container{*testContainer}}},
	},
}

func createTestClient() {
	if len(clients) == 0 {
		var err error
		var kubeconfig string
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		clientConfig1 := map[string]string{
			"kubeconfig": kubeconfig,
			"context":    "kind-kind-test",
		}
		clientConfig2 := map[string]string{
			"kubeconfig": kubeconfig,
			"context":    "kind-kind2-test",
		}

		err = CreateClient(clientConfig1, clientConfig2)
		if err != nil {
			panic(err)
		}
	}
}

// BUG: clusterCRD will be overriden after first Apply. Cannot use for second Apply!
// TODO: Should apply CRD for all clientset at once
func appendCRD() {
	url := "https://raw.githubusercontent.com/cisco-open/cluster-registry-controller/cb563ec383a6a98f8d8e5c79d3350997b7e70075/deploy/charts/cluster-registry/crds/clusterregistry.k8s.cisco.com_clusters.yaml"
	clusterCRD, err := io.GetClusterCRD(url)
	if err != nil {
		panic(err.Error())
	}

	_ = Apply(clusterCRD) // Need clientset mapper refresh

	time.Sleep(3 * time.Second) // Wait for cluster CRD init
	clients = []Clientset{}     // Delete all clientset with old mapping
	createTestClient()          // Create new clientsets with new mapping (memcache will be invalidated)
}

func GetNamespaceStatus(namespace string) string {
	key := client.ObjectKey{Name: namespace}
	ns := &corev1.Namespace{
		Status: corev1.NamespaceStatus{},
	}
	err := ActiveClientset.client.Get(context.TODO(), key, ns, &client.GetOptions{})
	if err != nil && err.Error() == "namespaces \"namespace-for-testing\" not found" {
		return "Not found"
	}

	if ns.Status.Phase == corev1.NamespacePhase("Active") {
		return "Active"
	}

	return "Terminating"
}

func WaitForReadyDeployment(deployment appsv1.Deployment) {
	key := client.ObjectKey{
		Name:      deployment.Name,
		Namespace: deployment.Namespace}
	deploy := &appsv1.Deployment{}

	timeout := 30 * time.Second
	for timeout > 0 {
		err := ActiveClientset.client.Get(context.TODO(), key, deploy, &client.GetOptions{})
		if err == nil && deploy.Status.ReadyReplicas > 0 && deploy.Status.Replicas == deploy.Status.ReadyReplicas {
			return
		}

		timeout -= 3 * time.Second
		time.Sleep(3 * time.Second)
	}
}

func setupCluster() {
	for i := 0; i < len(clients); i++ {
		SetActiveClientset(clients[i])

		appendCRD()

		timeout := 120 * time.Second
		for true {
			nsPhase := GetNamespaceStatus(testNamespaceName)
			if nsPhase == "Active" {
				break
			} else if nsPhase == "Not found" {
				_ = CreateNamespace(testNamespaceName)
			}

			fmt.Println("Phase: " + nsPhase)
			if timeout == 0 {
				panic("cannot create namespace")
			}

			time.Sleep(3 * time.Second)
			timeout -= 3 * time.Second
		}
	}
}

func resetCluster() {
	for i := 0; i < len(clients); i++ {
		SetActiveClientset(clients[i])

		_ = Remove(&testDeployment)

		_ = Detach(testNamespaceName, testNamespaceName)

		_ = Remove(testCluster1)
		_ = Remove(testCluster2)
		_ = Remove(testSecret1)
		_ = Remove(testSecret2)

		testDeployment.ResourceVersion = ""
		testCluster1.ResourceVersion = ""
		testCluster2.ResourceVersion = ""
		testSecret1.ResourceVersion = ""
		testSecret2.ResourceVersion = ""

		_ = DeleteNamespace(testNamespaceName)
	}
}

func TestCreateNamespace(t *testing.T) {
	createTestClient()
	resetCluster()

	err := CreateNamespace(testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetNamespace(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	namespace, err := GetNamespace(testNamespaceName)
	if err != nil || namespace == nil || namespace.Name != testNamespaceName {
		t.Error(err.Error())
	}

}

func TestDeleteNamespace(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	err := DeleteNamespace(testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestIsNamespaceExists(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	exists, err := IsNamespaceExists(testNamespaceName)
	if err != nil && exists != true {
		t.Error(err.Error())
	}
}

func TestApply(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	err := Apply(&testDeployment)
	if err != nil {
		t.Error(err)
	}

	WaitForReadyDeployment(testDeployment)
}

func TestRemove(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	_ = Apply(&testDeployment)
	WaitForReadyDeployment(testDeployment)

	err := Remove(&testDeployment)

	if err != nil {
		t.Error("Try to delete non-exist custom resource")
	}
}

func TestAPIServerEndpoint(t *testing.T) {
	createTestClient()

	endpoint, err := GetAPIServerEndpoint()
	if err != nil || endpoint == "" {
		t.Error(err.Error())
	}
}

func TestVerify(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	_ = Apply(&testDeployment)
	WaitForReadyDeployment(testDeployment)

	testTimeout := 15 * time.Second
	err := Verify(testDeploymentName, testNamespaceName, testTimeout)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestGetDeploymentName(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	_ = Apply(&testDeployment)
	WaitForReadyDeployment(testDeployment)

	deploymentName, err := GetDeploymentName(testDeploymentReleaseName, testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}

	if testDeploymentName != deploymentName {
		t.Error("Deployment name is wrong!")
	}
}

func TestGetClusterInfo(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	_ = Apply(testSecret1)
	_ = Apply(testCluster1)

	NamespacedClient := client.NewNamespacedClient(clients[0].client, testNamespaceName)
	clusterInfo, err := getClusterInfo(NamespacedClient, objectKey1)
	if err != nil {
		t.Error(err)
	}

	if clusterInfo.secret.Name != testSecret1.Name || clusterInfo.cluster.Name != testCluster1.Name {
		t.Error("wrong resource name")
	}
}

func TestAttach(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(testSecret1)
	_ = Apply(testCluster1)

	SetActiveClientset(clients[1])

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(testSecret2)
	_ = Apply(testCluster2)

	err := Attach(testNamespaceName, testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestDetach(t *testing.T) {
	createTestClient()
	resetCluster()
	setupCluster()

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(testSecret1)
	_ = Apply(testCluster1)

	SetActiveClientset(clients[1])

	_ = CreateNamespace(testNamespaceName)
	_ = Apply(testSecret2)
	_ = Apply(testCluster2)

	_ = Attach(testNamespaceName, testNamespaceName)

	err := Detach(testNamespaceName, testNamespaceName)
	if err != nil {
		t.Error(err.Error())
	}
}
