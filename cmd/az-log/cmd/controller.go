/*
Copyright 2022 The Kubernetes Authors.

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
	"errors"
	// "fmt"
	"strings"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	consts "sigs.k8s.io/azuredisk-csi-driver/pkg/azureconstants"
	v1 "k8s.io/api/core/v1"
)

// controllerCmd represents the controller command
var controllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		volumes, nodes, requestIds, since, sinceTime, isFollow, isPrevious := GetFlags(cmd)
		config := getConfig()
		clientsetK8s := getKubernetesClientset(config)

		for {
			pod := GetAzuredisk2Controller(clientsetK8s)
			GetLogsByAzDriverPod(clientsetK8s, pod.Name, AzureDiskContainer, volumes, nodes, requestIds, since, sinceTime, isFollow, isPrevious)
			// If in watch mode (--follow) and the pod failover and restarts, keep watching logs from new controller
			if isFollow == false || pod.Status.Conditions[1].Status == v1.ConditionTrue {
				break
			}
		}

	},
}

func init() {
	getCmd.AddCommand(controllerCmd)
}

func GetAzuredisk2Controller(clientsetK8s kubernetes.Interface) *v1.Pod {
	//Get node that leader controller pod is running in
	lease, err := clientsetK8s.CoordinationV1().Leases(consts.DefaultAzureDiskCrdNamespace).Get(context.Background(), "default", metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	holder := *lease.Spec.HolderIdentity
	node := strings.Split(holder, "_")[0]

	// Get pod name of the leader controller
	pods, err := clientsetK8s.CoreV1().Pods(consts.ReleaseNamespace).List(context.Background(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node,
		LabelSelector: "app=csi-azuredisk2-controller",
	})
	if err != nil {
		panic(err.Error())
	}
	if len(pods.Items) > 1 {
		panic(errors.New("more than one controller pods were found"))
	}
	//fmt.Println(pods.Items[0].Status.Conditions[1].Status) // Type == v1.PodReady // == v1.ConditionTrue
	return &pods.Items[0]
}
