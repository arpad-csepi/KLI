package kubectl

import (
	"context"
	"errors"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	istio_operator "github.com/banzaicloud/istio-operator/api/v2/v1alpha1"
	cluster_registry "github.com/cisco-open/cluster-registry-controller/api/v1alpha1"
)

type Clientset struct {
	client client.Client
	config *rest.Config
	discovery *discovery.DiscoveryClient
}

type clusterInfo struct {
	restClient client.Client
	secret     corev1.Secret
	cluster    cluster_registry.Cluster
}

var Clients []Clientset

// CreateClient set up kubernetes REST client which scheme contains custom kubernetes types from banzaicloud and cisco-open
func CreateClient(kubeconfig ...*string) error {
	if len(kubeconfig) == 0 {
		return errors.New("no kubeconfig was definied")
	}

	for i := 0; i < len(kubeconfig); i++ {
		clientset := Clientset {}
		// REST configuration for creating custom client
		restConfig, err := clientcmd.BuildConfigFromFlags("", *kubeconfig[i])
		if err != nil {
			return err
		}

		clientset.config = restConfig

		// discoverClient discover server-supported API groups, versions and resources.
		discoveryClient, err := discovery.NewDiscoveryClientForConfig(restConfig)
		if err != nil {
			return err
		}

		clientset.discovery = discoveryClient

		// runtimeScheme contains already registered types in the API server
		runtimeScheme := scheme.Scheme

		apiextensions.AddToScheme(runtimeScheme)
		appsv1.AddToScheme(runtimeScheme)

		// Add custom types to the runtime scheme
		err = istio_operator.SchemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			return err
		}
		err = cluster_registry.SchemeBuilder.AddToScheme(runtimeScheme)
		if err != nil {
			return err
		}
		// mapper initializes a mapping between Kind and APIVersion to a resource name and back based on the objects in a runtime.Scheme and the Kubernetes API conventions.
		mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient))

		// restClient is the custom client which known the custom resource types
		customClient, err := client.New(restConfig, client.Options{Scheme: runtimeScheme, Mapper: mapper, Opts: client.WarningHandlerOptions{}})
		if err != nil {
			return err
		}

		clientset.client = customClient

		Clients = append(Clients, clientset)
	}

	return nil
}

// TODO: Deprecated due to helm can create namespace before install
// CreateNamespace create namespace to provided kubeconfig kubecontext
func CreateNamespace(namespace string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}

	err := Clients[0].client.Create(context.Background(), ns, &client.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func GetNamespace(namespace string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{}
	key := types.NamespacedName{
		Name: namespace,
	}

	err := Clients[0].client.Get(context.TODO(), key, ns, &client.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func DeleteNamespace(namespace string) error {
	ns, err := GetNamespace(namespace)
	if err != nil {
		return err
	}

	err = Clients[0].client.Delete(context.TODO(), ns, &client.DeleteOptions{})
	if err != nil {
		return err
	}

	return nil
}

// IsNamespaceExists check the given namespace is exists already or not
func IsNamespaceExists(namespace string) (bool, error) {
	nsList := &corev1.NamespaceList{}

	err := Clients[0].client.List(context.TODO(), nsList, &client.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, namespaceItem := range nsList.Items {
		if namespaceItem.Name == namespace {
			return true, nil
		}
	}

	return false, nil
}

// Verify check release status until the given time
func Verify(deploymentName string, namespace string, timeout time.Duration) error {
	animation := [7]string{"_", "-", "`", "'", "´", "-", "_"}
	frame := 0

	key := types.NamespacedName{
		Namespace: namespace,
		Name:      deploymentName,
	}

	for start := time.Now(); ; {
		fmt.Printf("Verifing the %s deployment: [%s]", deploymentName, animation[frame])

		deployment := &appsv1.Deployment{}
		
		err := Clients[0].client.Get(context.TODO(), key, deployment, &client.GetOptions{})
		if err != nil {
			return err
		}
		if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
			fmt.Println("\nOk! Verify process was successful!")
			break
		}
		if time.Since(start) > timeout {
			fmt.Println("\nAww. One or more resource is not ready! Please check your cluster to more info.")
			return errors.New("resources are not ready")
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
func Apply(CRObject client.Object) error {
	fmt.Printf("Apply %s resource file to %s namespace\n", CRObject.GetName(), CRObject.GetNamespace())
	
	NamespacedClient := client.NewNamespacedClient(Clients[0].client, CRObject.GetNamespace())

	err := NamespacedClient.Create(context.TODO(), CRObject)
	if err != nil {
		return err
	}
	fmt.Printf("Yep, %s resource applied\n", CRObject.GetName())

	return nil
}

// Remove is read the custom resource definition and remove it with custom REST client
func Remove(CRObject client.Object) error {
	fmt.Printf("Remove resource based on %s\n", CRObject.GetName())

	NamespacedClient := client.NewNamespacedClient(Clients[0].client, CRObject.GetNamespace())

	err := NamespacedClient.DeleteAllOf(context.TODO(), CRObject)
	if err != nil {
		return err
	}

	fmt.Println("Resource deleted!")
	return nil
}

// GetAPIServerEndpoint is return with the API endpoint URL address
// TODO: Do check for non-valid values
func GetAPIServerEndpoint() (string, error) {
	endpoint := Clients[0].discovery.RESTClient().Get().URL().Host
	return endpoint, nil
}

// GetDeploymentName is search the deployment name based on the chart release name
func GetDeploymentName(releaseName string, namespace string) (string, error) {
	deployments := &appsv1.DeploymentList{}
	err := Clients[0].client.List(context.TODO(), deployments, &client.ListOptions{})
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
func Attach(namespace1 string, namespace2 string) error {
	fmt.Println("Attach process started")

	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	NamespacedClient1 := client.NewNamespacedClient(Clients[0].client, namespace1)
	NamespacedClient2 := client.NewNamespacedClient(Clients[1].client, namespace2)

	fmt.Println("Get some info from clusters")
	cluster1Info, err := getClusterInfo(NamespacedClient1, objectKey1)
	if err != nil {
		return err
	}

	cluster2Info, err := getClusterInfo(NamespacedClient2, objectKey2)
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
func Detach(namespace1 string, namespace2 string) error {
	fmt.Println("Detach process started!")

	objectKey1 := client.ObjectKey{Namespace: namespace1, Name: "demo-active"}
	objectKey2 := client.ObjectKey{Namespace: namespace2, Name: "demo-passive"}

	NamespacedClient1 := client.NewNamespacedClient(Clients[0].client, namespace1)
	NamespacedClient2 := client.NewNamespacedClient(Clients[1].client, namespace2)

	fmt.Println("Get clusters and secrets info, please wait...")

	//TODO: Make a better struct for more compact code
	cluster1Info, err1 := getClusterInfo(NamespacedClient1, objectKey1)
	if err1 != nil {
		fmt.Printf("%s not here on the main cluster.\n", objectKey1.Name)
	} else {
		delete(cluster1Info.restClient, &cluster1Info.cluster)
		delete(cluster1Info.restClient, &cluster1Info.secret)
	}
	cluster1Info2, err2 := getClusterInfo(NamespacedClient1, objectKey2)
	if err2 != nil {
		fmt.Printf("%s not here on the main cluster.\n", objectKey2.Name)
	} else {
		delete(cluster1Info2.restClient, &cluster1Info2.cluster)
		delete(cluster1Info2.restClient, &cluster1Info2.secret)
	}

	cluster2Info, err3 := getClusterInfo(NamespacedClient2, objectKey1)
	if err3 != nil {
		fmt.Printf("%s not here on the secondary cluster.\n", objectKey1.Name)
	} else {
		delete(cluster2Info.restClient, &cluster2Info.cluster)
		delete(cluster2Info.restClient, &cluster2Info.secret)
	}
	cluster2Info2, err4 := getClusterInfo(NamespacedClient2, objectKey2)
	if err4 != nil {
		fmt.Printf("%s not here on the secondary cluster.\n", objectKey2.Name)
	} else {
		delete(cluster2Info2.restClient, &cluster2Info2.cluster)
		delete(cluster2Info2.restClient, &cluster2Info2.secret)
	}

	fmt.Println("Cluster or secret objects are removed.\nDetach completed!")
	return nil
}

// getClusterInfo is return the secret and cluster object from the given REST client cluster
func getClusterInfo(clientset client.Client, objectKey client.ObjectKey) (clusterInfo, error) {
	var clusterInfoObj clusterInfo
	var timeout = 3

	for {
		if clusterInfoObj.secret.CreationTimestamp.IsZero() {
			err := clientset.Get(context.TODO(), objectKey, &clusterInfoObj.secret, &client.GetOptions{})
			if err != nil {
				return clusterInfoObj, err
			}
		}

		if clusterInfoObj.cluster.CreationTimestamp.IsZero() {
			err := clientset.Get(context.TODO(), objectKey, &clusterInfoObj.cluster)
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
	clusterInfoObj.restClient = clientset

	return clusterInfoObj, nil
}
