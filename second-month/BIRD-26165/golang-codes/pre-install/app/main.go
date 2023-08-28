package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/retry"
)

func main() {
	var namespace string
	flag.StringVar(&namespace, "namespace", "default", "Namespace where the deployment is located")
	flag.Parse()

	// Get Kubernetes configuration
	config, err := getClientConfig()
	if err != nil {
		log.Fatalf("Error creating client config: %v", err)
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Get deployment list
	deployments, err := clientset.AppsV1().Deployments(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("Error getting deployments: %v", err)
	}

	// Iterate through deployments to find and scale the one with "magento-cron" in its name
	for _, deployment := range deployments.Items {
		if strings.Contains(deployment.ObjectMeta.Name, "magento-cron") {
			// Scale deployment to zero replicas
			retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				deployment.Spec.Replicas = int32Ptr(0)
				_, updateErr := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), &deployment, metav1.UpdateOptions{})
				return updateErr
			})
			if retryErr != nil {
				log.Fatalf("Error scaling deployment: %v", retryErr)
			}
			fmt.Printf("Scaled deployment %s to zero replicas.\n", deployment.ObjectMeta.Name)
		}
	}
}

func getClientConfig() (*rest.Config, error) {
	// Use in-cluster configuration if running inside a Kubernetes cluster
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster {
		return rest.InClusterConfig()
	}

	// Otherwise, use the kubeconfig file
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}

func int32Ptr(i int32) *int32 { return &i }
