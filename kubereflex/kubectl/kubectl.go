package kubectl

import (
	"context"
	"fmt"
	"github.com/arpad-csepi/KLI/kubereflex/io"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
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

func Apply(path string, kubeconfig *string) {
	var clientset = setKubeClient(kubeconfig)
	var controlPlane = io.ReadYAMLResourceFile(path)

	groupResources, err := restmapper.GetAPIGroupResources(clientset.Discovery())
	if err != nil {
		panic(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	// Get some metadata needed to make the REST request.
	gvk := controlPlane.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}

	mapping, err := mapper.RESTMapping(gk, gvk.Version)
	if err != nil {
		panic(err)
	}

	helper := resource.NewHelper(clientset.CoreV1().RESTClient(), mapping)

	_, err = helper.Create(controlPlane.Namespace, false, &controlPlane)
	if err != nil {
		panic(err)
	}
}
