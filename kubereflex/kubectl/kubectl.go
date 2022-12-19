package kubectl

import (
	"context"
	"fmt"
	"time"

	"github.com/arpad-csepi/KLI/kubereflex/io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istio_operator "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
)

// setKubeClient set up kubernetes clientset from the given kubeconfig and return with that
func setKubeClient(kubeconfig *string) *kubernetes.Clientset {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	return clientset
}

func createCustomClient(namespace string, kubeconfig *string) client.Client {
	// REST configuration for creating custom client
	var restConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}

	// discoverClient discover server-supported API groups, versions and resources.
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		panic(err)
	}

	// runtimeScheme contains already registered types in the API server
	var runtimeScheme = scheme.Scheme

	// Add custom types to the runtime scheme
	istio_operator.SchemeBuilder.AddToScheme(runtimeScheme)
	cluster_registry.SchemeBuilder.AddToScheme(runtimeScheme)

	// mapper initializes a mapping between Kind and APIVersion to a resource name and back based on the objects in a runtime.Scheme and the Kubernetes API conventions.
	var mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	// restCLient is the custom client which known the custom resource types
	restClient, err := client.New(restConfig, client.Options{Scheme: runtimeScheme, Mapper: mapper, Opts: client.WarningHandlerOptions{}})
	if err != nil {
		panic(err)
	}

	// Set a namespace for the client
	var namespacedRestClient client.Client = client.NewNamespacedClient(restClient, namespace)

	return namespacedRestClient
}

// TODO: Deprecated due to helm can create namespace before install
// CreateNamespace create namespace to provided kubeconfig kubecontext
func CreateNamespace(namespace string, kubeconfig *string) {
	var clientset = setKubeClient(kubeconfig)

	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	clientset.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
}

// IsNamespaceExists check the given namespace is exists already or not
func IsNamespaceExists(namespace string, kubeconfig *string) bool {
	var clientset = setKubeClient(kubeconfig)

	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}

	for _, namespaceItem := range namespaces.Items {
		if namespaceItem.Name == namespace {
			return true
		}
	}

	return false
}

// Verify check release status until the given time
// TODO: Make this asynchronous so other resources can be installed while verify is running (if not dependent one resource on another)
func Verify(deploymentName string, namespace string, kubeconfig *string, timeout time.Duration) {
	var clientset = setKubeClient(kubeconfig)
	// TODO: Make timeout check event based for more efficiency
	var animation = [7]string{"_", "-", "`", "'", "´", "-", "_"}
	var frame = 0

	for start := time.Now(); ; {
		fmt.Printf("Verifing the %s deployment: [%s]", deploymentName, animation[frame])
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
			fmt.Println("\nOk! Verify process was successful!")
			break
		}
		if time.Since(start) > timeout {
			// TODO: List the resources which cause the timeout
			fmt.Println("\nAww. One or more resource is not ready! Please check your cluster to more info.")
			break
		}
		time.Sleep(150 * time.Millisecond)
		fmt.Print("\033[G") // Move cursor to line begining

		// TODO: Fix this if-else with 1 line formula later
		if frame == 6 {
			frame = 0
		} else {
			frame += 1
		}
	}
}

func Apply(CRDPath string, kubeconfig *string) {
	CRD := io.ReadYAMLResourceFile(CRDPath)

	restClient := createCustomClient(CRD.GetNamespace(), kubeconfig)

	// Create a custom resource from the CRD file
	err := restClient.Create(context.TODO(), CRD)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Yep, %s is created\n", CRD.GetName())
}

func Remove(CRDpath string, kubeconfig *string) {
	// CRD object read from file
	CRD := io.ReadYAMLResourceFile(CRDpath)

	restClient := createCustomClient(CRD.GetNamespace(), kubeconfig)

	// Create a custom resource from the CRD file
	err := restClient.DeleteAllOf(context.TODO(), CRD)
	if err != nil {
		fmt.Println(err)
	}
}

func GetAPIServerEndpoint(kubeconfig *string) string {
	clientset := setKubeClient(kubeconfig)
	return clientset.DiscoveryClient.RESTClient().Get().URL().String()
}

func GetDeploymentName(releaseName string, namespace string, kubeconfig *string) string {
	clientset := setKubeClient(kubeconfig)
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// TODO: More efficient way to found out the deployment name ???
	for _, deployment := range deployments.Items {
		for _, value := range deployment.Annotations {
			if value == releaseName {
				return deployment.Name
			}
		}
	}
	panic("Deployment name not found")
}

func create(restClient client.Client, obj client.Object) {
	err := restClient.Create(context.TODO(), obj)
	if err != nil {
		panic(err)
	}
}

func delete(restClient client.Client, obj client.Object) {
	err := restClient.DeleteAllOf(context.TODO(), obj)
	if err != nil {
		fmt.Println(err)
	}
}

func Attach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	restClient1 := createCustomClient(namespace1, kubeconfig1)
	secret1, cluster1 := getClusterInfo(restClient1, objectKey1)

	restClient2 := createCustomClient(namespace2, kubeconfig2)
	secret2, cluster2 := getClusterInfo(restClient2, objectKey2)

	create(restClient1, &secret2)
	create(restClient1, &cluster2)

	create(restClient2, &secret1)
	create(restClient2, &cluster1)
}

func getClusterInfo(restClient client.Client, objectKey client.ObjectKey) (corev1.Secret, cluster_registry.Cluster) {
	var secret corev1.Secret
	var cluster cluster_registry.Cluster
	var errSecret error
	var errCluster error

	// 50 * 100ms = 5sec
	var timeout = 50

	for {
		if secret.CreationTimestamp.IsZero() {
			errSecret = restClient.Get(context.TODO(), objectKey, &secret)
		}

		if cluster.CreationTimestamp.IsZero() {
			errCluster = restClient.Get(context.TODO(), objectKey, &cluster)
		}

		if errSecret == nil && errCluster == nil {
			break
		} else if timeout == 0 {
			err := fmt.Sprintf("Secret resource error: %s\nCluster resource error: %s", errSecret.Error(), errCluster.Error())
			panic(err)
		}
		time.Sleep(100 * time.Millisecond)
		timeout--
	}

	secret.ResourceVersion = ""
	cluster.ResourceVersion = ""

	return secret, cluster
}

func Detach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	restClient1 := createCustomClient(namespace1, kubeconfig1)
	secretActive1, clusterActive1 := getClusterInfo(restClient1, objectKey1)
	secretActive2, clusterActive2 := getClusterInfo(restClient1, objectKey2)

	restClient2 := createCustomClient(namespace2, kubeconfig2)
	secretPassive1, clusterPassive1 := getClusterInfo(restClient2, objectKey1)
	secretPassive2, clusterPassive2 := getClusterInfo(restClient2, objectKey2)

	delete(restClient1, &secretActive1)
	delete(restClient1, &secretActive2)
	delete(restClient1, &clusterActive1)
	delete(restClient1, &clusterActive2)

	delete(restClient2, &secretPassive1)
	delete(restClient2, &secretPassive2)
	delete(restClient2, &clusterPassive1)
	delete(restClient2, &clusterPassive2)
}
