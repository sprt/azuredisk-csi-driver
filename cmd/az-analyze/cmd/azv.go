/*
Copyright YEAR The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	//"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	v1beta1 "sigs.k8s.io/azuredisk-csi-driver/pkg/apis/azuredisk/v1beta1"
	azDiskClientSet "sigs.k8s.io/azuredisk-csi-driver/pkg/apis/client/clientset/versioned"
	consts "sigs.k8s.io/azuredisk-csi-driver/pkg/azureconstants"
)

// azvCmd represents the azv command
var azvCmd = &cobra.Command{
	Use:   "azv",
	Short: "Azure Volume",
	Long: `Azure Volume is a Kubernetes Custom Resource.`,
	Run: func(cmd *cobra.Command, args []string) {
		pod, _ := cmd.Flags().GetString("pod")

		var result []AzvResource
		result = GetAzVolumesByPod(pod)

		// display
		if len(result) != 0 {
			displayAzv(result)
		} else {
			// not found, display an error
			fmt.Println("azVolumes not found")
		}
		
	},
}

func init() {
	getCmd.AddCommand(azvCmd)
	azvCmd.PersistentFlags().StringP("pod", "p", "", "insert-pod-name")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// azvCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// azvCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type AzvResource struct {
	ResourceType string
	Namespace    string
	Name         string
	State        v1beta1.AzVolumeState
}

func GetAzVolumesByPod(podName string) []AzvResource {
	// implementation
	result := make([]AzvResource, 0)

	// access to config
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// get pvc claim name set of pod : assume default namesapce
	pvcSet := make(map[string]string)
	clientset1, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	} else {
		pods, err := clientset1.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}

		for _, pod := range pods.Items {
			// if pod flag isn't provided, print all pods
			if podName == "" || pod.Name == podName {
				for _, v := range pod.Spec.Volumes {
					if v.PersistentVolumeClaim != nil {
						pvcSet[v.PersistentVolumeClaim.ClaimName] = pod.Name
					}
				}
			}
		}
	}

	// get azVolumes with the same claim name in pvcSet
	clientset2, err := azDiskClientSet.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	} else {
		azVolumes, err := clientset2.DiskV1beta1().AzVolumes(consts.DefaultAzureDiskCrdNamespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		for _, azVolume := range azVolumes.Items {
			// pvName := azVolume.Status.PersistentVolume;
			pvcClaimName := azVolume.Spec.Parameters["csi.storage.k8s.io/pvc/name"]

			// if pvcName is contained in pvcSet, add the azVolume to result
			if pName, ok := pvcSet[pvcClaimName]; ok {
				result = append(result, AzvResource {
					ResourceType: pName,
					Namespace:    azVolume.Namespace,
					Name:         azVolume.Status.PersistentVolume,
					State:        azVolume.Status.State})
			}
		}
	}

	return result
}

func displayAzv(result []AzvResource) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"PODNAME", "NAMESPACE", "NAME", "STATE"})
	// var row []string
	// v := reflect.ValueOf(azv)

	// for i := 0; i < v.NumField(); i++ {
	// 	row = append(row, v.Field(i).Interface())
	// }
	for _, azv := range result {
		table.Append([]string{azv.ResourceType, azv.Namespace, azv.Name, string(azv.State)})
	}

	table.Render()
}
