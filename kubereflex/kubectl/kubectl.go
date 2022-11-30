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

// TODO: Deprecated due to helm can create namespace before install
// CreateNamespace create namespace to provided kubeconfig kubecontext
func CreateNamespace(namespace string, kubeconfig *string) {
	var clientset = setKubeClient(kubeconfig)

	//TODO: Check if setKubeClient failed

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

	// Add custom istio_operator types to the runtime scheme
	istio_operator.SchemeBuilder.AddToScheme(runtimeScheme)

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

func Apply(CRDpath string, kubeconfig *string) {
	// CRD object read from file
	CRDObject := io.ReadYAMLResourceFile(CRDpath)

	restClient := createCustomClient(CRDObject.Namespace, kubeconfig)

	// Create a custom resource from the CRD file
	err := restClient.Create(context.TODO(), &CRDObject)
	if err != nil {
		panic(err)
	}
}

func Delete(CRDpath string, kubeconfig *string) {
	// CRD object read from file
	CRDObject := io.ReadYAMLResourceFile(CRDpath)

	restClient := createCustomClient(CRDObject.Namespace, kubeconfig)

	// Create a custom resource from the CRD file
	err := restClient.DeleteAllOf(context.TODO(), &CRDObject)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Hooray, %s successfuly deleted\n", CRDObject.Name)
}

func GetAPIServerEndpoint(kubeconfig *string) string {
	clientset := setKubeClient(kubeconfig)
	return clientset.DiscoveryClient.RESTClient().Get().URL().String()
}

func Attach() {
	panic("Not implemented")
}

func Detach() {
	panic("Not implemented")
}
