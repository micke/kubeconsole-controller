package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

var verbose bool

func main() {
	_, verbose = os.LookupEnv("VERBOSE")

	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	podsClient := clientset.CoreV1().Pods("")

	for {
		verboseLog("Looking for kubeconsole pods...")

		pods, err := podsClient.List(context.TODO(), metav1.ListOptions{
			LabelSelector: "kubeconsole.garbagecollect=true",
		})
		if err != nil {
			panic(err.Error())
		}

		verboseLog("There are %d kubconsole pods in the cluster", len(pods.Items))

		for _, pod := range pods.Items {
			verboseLog("Found pod %s", pod.Name)

			heartbeat, err := time.Parse(time.RFC3339, pod.Annotations["kubeconsole.heartbeat"])
			if err != nil {
				fmt.Printf("Error parsing heartbeat: %v\n", err)
				continue
			}

			timeoutMinutes, err := strconv.Atoi(pod.Annotations["kubeconsole.timeout"])
			if err != nil {
				fmt.Printf("Error parsing timeout: %v\n", err)
				continue
			}

			timeout := heartbeat.Add(time.Minute * time.Duration(timeoutMinutes))

			if time.Now().After(timeout) {
				err := clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Printf("Error deleting pod(%s): %v\n", pod.Name, err)
				} else {
					verboseLog("Deleted %s", pod.Name)
				}
			} else {
				verboseLog("Skipping %s", pod.Name)
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func verboseLog(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", args)
	}
}
