package kubectl

import (
	"flag"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getKubeConfig() *string {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	return kubeconfig
}

func TestGetClusterInfo(t *testing.T) {
	objectKey := client.ObjectKey{Namespace: "cluster-registry", Name: "demo-active"}

	restClient := createCustomClient(objectKey.Namespace, getKubeConfig())

	secret, cluster := getClusterInfo(restClient, objectKey)

	if secret.CreationTimestamp.IsZero() {
        t.Errorf("Failed to get secret")
    }
	if cluster.CreationTimestamp.IsZero() {
        t.Errorf("Failed to get cluster")
    }
}
