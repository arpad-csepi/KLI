package kubectl

import (
	"context"
	"errors"
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

type clusterInfo struct {
	restClient client.Client
	secret     corev1.Secret
	cluster    cluster_registry.Cluster
	errSecret  error
	errCluster error
}

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

// createCustomClient set up kubernetes REST client which scheme contains custom kubernetes types from banzaicloud and cisco-open
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

	// restClient is the custom client which known the custom resource types
	restClient, err := client.New(restConfig, client.Options{Scheme: runtimeScheme, Mapper: mapper, Opts: client.WarningHandlerOptions{}})
	if err != nil {
		panic(err)
	}

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
func Verify(deploymentName string, namespace string, kubeconfig *string, timeout time.Duration) {
	var clientset = setKubeClient(kubeconfig)
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
			fmt.Println("\nAww. One or more resource is not ready! Please check your cluster to more info.")
			break
		}
		time.Sleep(150 * time.Millisecond)
		fmt.Print("\033[G")
		if frame == 6 {
			frame = 0
		} else {
			frame += 1
		}
	}
}

// Apply is read the custom resource definition and apply it with custom REST client
func Apply(CRDPath string, kubeconfig *string) {
	fmt.Printf("Apply %s resource file\n", CRDPath)
	CRD := io.ReadYAMLResourceFile(CRDPath)
	restClient := createCustomClient(CRD.GetNamespace(), kubeconfig)

	err := restClient.Create(context.TODO(), CRD)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("Yep, %s resource applied\n", CRD.GetName())
}

// Remove is read the custom resource definition and remove it with custom REST client
func Remove(CRDpath string, kubeconfig *string) error {
	fmt.Printf("Remove resource based on %s\n", CRDpath)
	CRD := io.ReadYAMLResourceFile(CRDpath)
	restClient := createCustomClient(CRD.GetNamespace(), kubeconfig)

	err := restClient.DeleteAllOf(context.TODO(), CRD)
	if err != nil {
		return (err)
	}

	fmt.Println("Resource deleted!")
	return nil
}

// GetAPIServerEndpoint is return with the API endpoint URL address
func GetAPIServerEndpoint(kubeconfig *string) string {
	clientset := setKubeClient(kubeconfig)
	return clientset.DiscoveryClient.RESTClient().Get().URL().String()
}

// GetDeploymentName is search the deployment name based on the chart release name
func GetDeploymentName(releaseName string, namespace string, kubeconfig *string) (string, error) {
	clientset := setKubeClient(kubeconfig)
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		return "", err
	}

	for _, deployment := range deployments.Items {
		for _, value := range deployment.Annotations {
			if value == releaseName {
				return deployment.Name, nil
			}
		}
	}

	err = errors.New("deployment not found")
	return "", err
}

// create is create an object in the cluster with the given REST client
func create(restClient client.Client, obj client.Object) {
	err := restClient.Create(context.TODO(), obj)
	if err != nil {
		panic(err)
	}
}

// delete is delete an object in the cluster with the given REST client
func delete(restClient client.Client, obj client.Object) {
	err := restClient.DeleteAllOf(context.TODO(), obj)
	if err != nil {
		fmt.Println(err)
	}
}

// Attach is get the secret and cluster objects and create on the another cluster so can sync after that
func Attach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	fmt.Println("Attach process started")

	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	restClient1 := createCustomClient(namespace1, kubeconfig1)
	restClient2 := createCustomClient(namespace2, kubeconfig2)

	fmt.Println("Get some info from clusters")
	cluster1Info := getClusterInfo(restClient1, objectKey1)
	cluster2Info := getClusterInfo(restClient2, objectKey2)

	if cluster2Info.errCluster != nil || cluster2Info.errSecret != nil {
		fmt.Printf("Cluster obj error: %v\n", cluster2Info.errCluster)
		fmt.Printf("Secret obj error: %v\n", cluster2Info.errSecret)
	} else if cluster1Info.errCluster != nil || cluster1Info.errSecret != nil {
		fmt.Printf("Cluster obj error: %v\n", cluster1Info.errCluster)
		fmt.Printf("Secret obj error: %v\n", cluster1Info.errSecret)
	} else {
		fmt.Println("Sync resources between clusters")
		create(cluster1Info.restClient, &cluster2Info.secret)
		create(cluster1Info.restClient, &cluster2Info.cluster)
		create(cluster2Info.restClient, &cluster1Info.secret)
		create(cluster2Info.restClient, &cluster1Info.cluster)
	}

	fmt.Println("Attach completed!")
}

// Detach is get the secret and cluster objects and delete on the another cluster so break the sync after that
func Detach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) {
	fmt.Println("Detach process started!")

	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	restClient1 := createCustomClient(namespace1, kubeconfig1)
	restClient2 := createCustomClient(namespace2, kubeconfig2)

	fmt.Println("Get clusters and secrets info, please wait...")

	cluster1Info := getClusterInfo(restClient1, objectKey1)
	cluster1Info2 := getClusterInfo(restClient1, objectKey2)

	cluster2Info := getClusterInfo(restClient2, objectKey1)
	cluster2Info2 := getClusterInfo(restClient2, objectKey2)

	for _, ci := range []clusterInfo{cluster1Info, cluster1Info2, cluster2Info, cluster2Info2} {
		if ci.errCluster != nil || ci.errSecret != nil {
			fmt.Println("Cluster or secret object not found! Skipped.")
		} else {
			delete(ci.restClient, &ci.cluster)
			delete(ci.restClient, &ci.secret)
			fmt.Println("Cluster or secret object found! Removed.")
		}
	}

	fmt.Println("Detach completed!")
}

// getClusterInfo is return the secret and cluster object from the given REST client cluster
func getClusterInfo(restClient client.Client, objectKey client.ObjectKey) clusterInfo {
	var clusterInfoObj clusterInfo
	var timeout = 3

	for {
		if clusterInfoObj.secret.CreationTimestamp.IsZero() {
			clusterInfoObj.errSecret = restClient.Get(context.TODO(), objectKey, &clusterInfoObj.secret)
		}

		if clusterInfoObj.cluster.CreationTimestamp.IsZero() {
			clusterInfoObj.errCluster = restClient.Get(context.TODO(), objectKey, &clusterInfoObj.cluster)
		}

		if clusterInfoObj.errSecret == nil && clusterInfoObj.errCluster == nil {
			break
		} else if timeout == 0 {
			return clusterInfoObj
		}
		time.Sleep(1 * time.Second)
		timeout--
	}

	clusterInfoObj.secret.ResourceVersion = ""
	clusterInfoObj.cluster.ResourceVersion = ""
	clusterInfoObj.restClient = restClient

	return clusterInfoObj
}
