/*
Copyright © 2022 Árpád Csepi csepi.arpad@outlook.com

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"flag"
	"path/filepath"

	"k8s.io/client-go/util/homedir"

	kubereflex "github.com/arpad-csepi/kubereflex"
)

// TODO: Fill up long description
// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install istio-operator and cluster-registry-controller",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
    var kubeconfig = getKubeConfig()

		// Install istio-operator and cluster-registry-controller automaticly
		kubereflex.InstallHelmChart("https://kubernetes-charts.banzaicloud.com", "banzaicloud-stable", "istio-operator", "banzaicloud-stable", "istio-system", nil, kubeconfig)
  	kubereflex.InstallHelmChart("https://cisco-open.github.io/cluster-registry-controller", "cluster-registry", "cluster-registry", "cluster-registry", "cluster-registry", nil, kubeconfig)

		custom_resource_path, err := cmd.Flags().GetString("custom_resource_path")
		if err != nil {
			panic("Custom resource error")
		}
		if custom_resource_path != "" {
			fmt.Println("Custom resource has called")
			fmt.Println(custom_resource_path)
		}
	},
}

var custom_resource_path string

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	//TODO: Maybe seperate kubernetes pod install and helm chart install here later

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	installCmd.Flags().StringVarP(&custom_resource_path, "custom_resource_path", "c", "", "Specify custom resource file location")
}

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
