package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

type Row struct {
	Title           string
	Status          string
	UpdateOn        string
	ScheduleStatus  string
	ScheduleUpdated string
}

type K8sClient struct {
	Client kubernetes.Interface
}

func getK8sClient() *K8sClient {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("error getting user home dir: %v\n", err)
		os.Exit(1)
	}
	kubeConfigPath := filepath.Join(userHomeDir, ".kube", "config")
	fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		err := fmt.Errorf("Error getting kubernetes config: %v\n", err)
		log.Fatal(err.Error)
	}
	client, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		err := fmt.Errorf("error getting kubernetes config: %v\n", err)
		log.Fatal(err.Error)
	}

	fmt.Printf("%T\n", client)
	return &K8sClient{
		Client: client,
	}
}

func (c *K8sClient) ScaleDeployment(name, namespace string, replica int32) {
	scaleObj, err := c.Client.AppsV1().Deployments(namespace).GetScale(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("error getting scale object: %v\n", err)
		os.Exit(1)
	}
	sd := *scaleObj
	if sd.Spec.Replicas == replica || replica < 0 {
		fmt.Printf("Deployment %s replicas %d, no changes applied\n", name, replica)
		return
	} else if sd.Spec.Replicas > replica {
		fmt.Printf("Scale down Deployment %s from %d to %d replicas\n", name, sd.Spec.Replicas, replica)
	} else {
		fmt.Printf("Scale Up Deployment %s from %d to %d replicas\n", name, sd.Spec.Replicas, replica)
	}
	sd.Spec.Replicas = replica
	scaleDeployment, err := c.Client.AppsV1().Deployments(namespace).UpdateScale(context.Background(), name, &sd, metav1.UpdateOptions{})
	if err != nil {
		fmt.Printf("error updating scale object: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully scaled deployment %s to %d replicas", name, scaleDeployment.Spec.Replicas)
}

func (c *K8sClient) ListDeployment(namespace string) (*appsv1.DeploymentList, error) {
	fmt.Println("List Deployments")
	deployments, err := c.Client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("error listing deployments: %v\n", err)
		return nil, err
	}
	return deployments, nil
}

func (c *K8sClient) GetDeployment(name, namespace string) (*appsv1.Deployment, error) { // => fmt.Printf("Created deployment %q.\n", result.GetName())
	fmt.Println("Get Deployment in namespace", namespace)
	result, err := c.Client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})

	if err != nil {
		fmt.Printf("Failed to getting Deployment: %v\n", err)
		return nil, err
	}
	return result, nil
}

func ListPods(namespace string, client kubernetes.Interface) (*v1.PodList, error) {
	fmt.Println("Get Kubernetes Pods")
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		err = fmt.Errorf("error getting pods: %v\n", err)
		return nil, err
	}
	return pods, nil
}

// php bin/magento indexer:status
// php bin/magento cache:clean
// php bin/magento cache:flush

func ExecCmdExample(client kubernetes.Interface, config *restclient.Config, podName string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	cmd := []string{
		"sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace("default").SubResource("exec")
	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}

	return nil
}

func GetPodNameByLabel(client kubernetes.Interface, namespace, labelSelector string) (string, error) {
	pods, err := client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "magento") {
			return pod.Name, nil
		}
	}
	return "", fmt.Errorf("No pod with label '%s' containing 'magento' found", labelSelector)
}

func GetDeploymentNameByLabel(client kubernetes.Interface, namespace, labelSelector string) (string, error) {
	deployments, err := client.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return "", err
	}
	for _, deployment := range deployments.Items {
		if strings.Contains(deployment.Name, "magento-cron") {
			return deployment.Name, nil
		}
	}
	return "", fmt.Errorf("No deployment with label '%s' containing 'magento-cron' found", labelSelector)
}

func main() {
	// client := getK8sClient()

	client := getK8sClient() // Initialize the Kubernetes client

	// Initialize Kubernetes client and config
	config, err := restclient.InClusterConfig()
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	namespace := "deafult"
	desiredReplicas := int32(0) // Scale to 0 replicas

	deploymentName, err := GetDeploymentNameByLabel(clientset, namespace, "app=magento-cron")
	if err != nil {
		fmt.Printf("Error getting deployment name: %v\n", err)
		os.Exit(1)
	}

	client.ScaleDeployment(deploymentName, namespace, desiredReplicas)

	command := "php bin/magento indexer:status"
	podName, err := GetPodNameByLabel(clientset, namespace, "app=magento")
	if err != nil {
		fmt.Printf("Error getting pod name: %v\n", err)
		os.Exit(1)
	}

	// Execute the Magento indexer status check command
	var stdout, stderr strings.Builder
	err = ExecCmdExample(clientset, config, podName, command, nil, &stdout, &stderr)
	if err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	}

	// Check the output for indexer status
	indexerStatus := strings.TrimSpace(stdout.String())
	if indexerStatus != "ready" {
		fmt.Println("Magento indexer status is not ready. Halting Kubernetes installation.")
		os.Exit(1)
	}

	// Continue with Kubernetes installation
	fmt.Println("Magento indexer status is ready. Proceeding with Kubernetes installation.")
}

// rows := []Row{
// 	{Title: "Catalog Product Rule", Status: "Reindex required", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// 	{Title: "Catalog Rule Product", Status: "Reindex required", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// 	{Title: "Catalog Search", Status: "Ready", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// 	{Title: "Category Products", Status: "Reindex required", UpdateOn: "Schedule", ScheduleStatus: "idle (0 in backlog)", ScheduleUpdated: "2021-06-28 09:45:53"},
// 	{Title: "Customer Grid", Status: "Ready", UpdateOn: "Schedule", ScheduleStatus: "idle (0 in backlog)", ScheduleUpdated: "2021-06-28 09:45:52"},
// 	{Title: "Design Config Grid", Status: "Ready", UpdateOn: "Schedule", ScheduleStatus: "idle (0 in backlog)", ScheduleUpdated: "2018-06-28 09:45:52"},
// 	{Title: "Inventory", Status: "Ready", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// 	{Title: "Product Categories", Status: "Reindex required", UpdateOn: "Schedule", ScheduleStatus: "idle (0 in backlog)", ScheduleUpdated: "2021-06-28 09:45:53"},
// 	{Title: "Product EAV", Status: "Reindex required", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// 	{Title: "Product Price", Status: "Reindex required", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// 	{Title: "Stock", Status: "Reindex required", UpdateOn: "Save", ScheduleStatus: "", ScheduleUpdated: ""},
// }

// var statuses []string
// for _, row := range rows {
// 	statuses = append(statuses, row.Status)
// }

// fmt.Println(statuses)
