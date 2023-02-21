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
}

// setKubeClient set up kubernetes clientset from the given kubeconfig and return with that
func setKubeClient(kubeconfig *string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// createCustomClient set up kubernetes REST client which scheme contains custom kubernetes types from banzaicloud and cisco-open
func createCustomClient(namespace string, kubeconfig *string) (client.Client, error) {
	// REST configuration for creating custom client
	restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}

	// discoverClient discover server-supported API groups, versions and resources.
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	// runtimeScheme contains already registered types in the API server
	runtimeScheme := scheme.Scheme

	// Add custom types to the runtime scheme
	err = istio_operator.SchemeBuilder.AddToScheme(runtimeScheme)
	if err != nil {
		return nil, err
	}
	err = cluster_registry.SchemeBuilder.AddToScheme(runtimeScheme)
	if err != nil {
		return nil, err
	}
	// mapper initializes a mapping between Kind and APIVersion to a resource name and back based on the objects in a runtime.Scheme and the Kubernetes API conventions.
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

	// restClient is the custom client which known the custom resource types
	restClient, err := client.New(restConfig, client.Options{Scheme: runtimeScheme, Mapper: mapper, Opts: client.WarningHandlerOptions{}})
	if err != nil {
		return nil, err
	}

	namespacedRestClient := client.NewNamespacedClient(restClient, namespace)

	return namespacedRestClient, nil
}

// TODO: Deprecated due to helm can create namespace before install
// CreateNamespace create namespace to provided kubeconfig kubecontext
func CreateNamespace(namespace string, kubeconfig *string) error {
	clientset, err := setKubeClient(kubeconfig)

	if err != nil {
		return err
	}

	nsName := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), nsName, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func GetNamespace(namespace string, kubeconfig *string) (*corev1.Namespace, error) {
	var ns *corev1.Namespace

	clientset, err := setKubeClient(kubeconfig)

	if err != nil {
		return nil, err
	}

	ns, err = clientset.CoreV1().Namespaces().Get(context.Background(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func DeleteNamespace(namespace string, kubeconfig *string) error {
	clientset, err := setKubeClient(kubeconfig)

	if err != nil {
		return err
	}

	err = clientset.CoreV1().Namespaces().Delete(context.Background(), namespace, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// IsNamespaceExists check the given namespace is exists already or not
func IsNamespaceExists(namespace string, kubeconfig *string) (bool, error) {
	clientset, err := setKubeClient(kubeconfig)
	if err != nil {
		return false, err
	}
	namespaces, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, namespaceItem := range namespaces.Items {
		if namespaceItem.Name == namespace {
			return true, nil
		}
	}

	return false, nil
}

// Verify check release status until the given time
func Verify(deploymentName string, namespace string, kubeconfig *string, timeout time.Duration) error {
	clientset, err := setKubeClient(kubeconfig)
	if err != nil {
		return err
	}
	animation := [7]string{"_", "-", "`", "'", "´", "-", "_"}
	frame := 0

	for start := time.Now(); ; {
		fmt.Printf("Verifing the %s deployment: [%s]", deploymentName, animation[frame])
		deployment, err := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return err
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

	return nil
}

// Apply is read the custom resource definition and apply it with custom REST client
func Apply(CRDPath string, kubeconfig *string) error {
	fmt.Printf("Apply %s resource file\n", CRDPath)
	CRD, err := io.ReadYAMLResourceFile(CRDPath)
	if err != nil {
		return err
	}
	restClient, err := createCustomClient(CRD.GetNamespace(), kubeconfig)
	if err != nil {
		return err
	}

	err = restClient.Create(context.TODO(), CRD)
	if err != nil {
		return err
	}
	fmt.Printf("Yep, %s resource applied\n", CRD.GetName())

	return nil
}

// Remove is read the custom resource definition and remove it with custom REST client
func Remove(CRDpath string, kubeconfig *string) error {
	fmt.Printf("Remove resource based on %s\n", CRDpath)
	CRD, err := io.ReadYAMLResourceFile(CRDpath)
	if err != nil {
		return err
	}
	restClient, err := createCustomClient(CRD.GetNamespace(), kubeconfig)
	if err != nil {
		return err
	}
	err = restClient.DeleteAllOf(context.TODO(), CRD)
	if err != nil {
		return err
	}

	fmt.Println("Resource deleted!")
	return nil
}

// GetAPIServerEndpoint is return with the API endpoint URL address
func GetAPIServerEndpoint(kubeconfig *string) (string, error) {
	clientset, err := setKubeClient(kubeconfig)
	if err != nil {
		return "", err
	}
	return clientset.DiscoveryClient.RESTClient().Get().URL().String(), nil
}

// GetDeploymentName is search the deployment name based on the chart release name
func GetDeploymentName(releaseName string, namespace string, kubeconfig *string) (string, error) {
	clientset, err := setKubeClient(kubeconfig)
	if err != nil {
		return "", err
	}
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
func Attach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) error {
	fmt.Println("Attach process started")

	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	restClient1, err := createCustomClient(namespace1, kubeconfig1)
	if err != nil {
		return err
	}
	restClient2, err := createCustomClient(namespace2, kubeconfig2)
	if err != nil {
		return err
	}

	fmt.Println("Get some info from clusters")
	cluster1Info, err := getClusterInfo(restClient1, objectKey1)
	if err != nil {
		return err
	}

	cluster2Info, err := getClusterInfo(restClient2, objectKey2)
	if err != nil {
		return err
	}

	fmt.Println("Sync resources between clusters")
	create(cluster1Info.restClient, &cluster2Info.secret)
	create(cluster1Info.restClient, &cluster2Info.cluster)
	create(cluster2Info.restClient, &cluster1Info.secret)
	create(cluster2Info.restClient, &cluster1Info.cluster)

	fmt.Println("Attach completed!")
	return nil
}

// Detach is get the secret and cluster objects and delete on the another cluster so break the sync after that
func Detach(kubeconfig1 *string, kubeconfig2 *string, namespace1 string, namespace2 string) error {
	fmt.Println("Detach process started!")

	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	restClient1, err := createCustomClient(namespace1, kubeconfig1)
	if err != nil {
		return err
	}
	restClient2, err := createCustomClient(namespace2, kubeconfig2)
	if err != nil {
		return err
	}

	fmt.Println("Get clusters and secrets info, please wait...")

	cluster1Info, err1 := getClusterInfo(restClient1, objectKey1)
	if err != nil {
		fmt.Printf("%s not here on the main cluster.\n", objectKey1.Name)
	}
	cluster1Info2, err2 := getClusterInfo(restClient1, objectKey2)
	if err != nil {
		fmt.Printf("%s not here on the main cluster.\n", objectKey2.Name)
	}

	cluster2Info, err3 := getClusterInfo(restClient2, objectKey1)
	if err != nil {
		fmt.Printf("%s not here on the secondary cluster.\n", objectKey1.Name)
	}
	cluster2Info2, err4 := getClusterInfo(restClient2, objectKey2)
	if err != nil {
		fmt.Printf("%s not here on the secondary cluster.\n", objectKey2.Name)
	}

	if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
		for _, ci := range []clusterInfo{cluster1Info, cluster1Info2, cluster2Info, cluster2Info2} {
			delete(ci.restClient, &ci.cluster)
			delete(ci.restClient, &ci.secret)
			fmt.Println("Cluster or secret objects are removed.")
		}
	}
	fmt.Println("Detach completed!")
	return nil
}

// getClusterInfo is return the secret and cluster object from the given REST client cluster
func getClusterInfo(restClient client.Client, objectKey client.ObjectKey) (clusterInfo, error) {
	var clusterInfoObj clusterInfo
	var timeout = 3

	for {
		if clusterInfoObj.secret.CreationTimestamp.IsZero() {
			err := restClient.Get(context.TODO(), objectKey, &clusterInfoObj.secret)
			if err != nil {
				return clusterInfoObj, err
			}
		}

		if clusterInfoObj.cluster.CreationTimestamp.IsZero() {
			err := restClient.Get(context.TODO(), objectKey, &clusterInfoObj.cluster)
			if err != nil {
				return clusterInfoObj, err
			}
		}

		if !clusterInfoObj.secret.CreationTimestamp.IsZero() && !clusterInfoObj.cluster.CreationTimestamp.IsZero() {
			break
		} else if timeout == 0 {
			err := errors.New("timeout reached")
			return clusterInfoObj, err
		}
		time.Sleep(1 * time.Second)
		timeout--
	}

	clusterInfoObj.secret.ResourceVersion = ""
	clusterInfoObj.cluster.ResourceVersion = ""
	clusterInfoObj.restClient = restClient

	return clusterInfoObj, nil
}
