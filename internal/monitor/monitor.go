package monitor

import (
	"net/http"
	"context"
	"strings"
	"time"
	"fmt"
	"log"

	"github.com/kerochan-web/sentinel/internal/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Check returns true if the service is healthy, false otherwise.
func Check(s config.Service) bool {
	switch s.Type {
        case "k8s":
                return checkKubernetesHealth(s.Target)
	case "http":
		return checkHTTP(s.Target)
	case "systemd":
		// For now, we'll just log that we aren't supporting this yet
		// Tiny steps!
		fmt.Printf("[Monitor] systemd check for %s not yet implemented\n", s.Name)
		return true
	default:
		fmt.Printf("[Monitor] Unknown service type: %s\n", s.Type)
		return false
	}
}

// checkKubernetesHealth connects to the local cluster API and audits statuses
func checkKubernetesHealth(target string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1. Authenticate using the in-cluster token provided by the ServiceAccount
	k8sConfig, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("[K8s Error] Failed to load in-cluster auth: %v", err)
		return false
	}

	clientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Printf("[K8s Error] Failed to create API client: %v", err)
		return false
	}

	// Target format expected in config: "namespace/deployment-name" (e.g., "todo-app/todo-app")
	parts := strings.Split(target, "/")
	if len(parts) != 2 {
		log.Printf("[K8s Error] Invalid target format '%s'. Use 'namespace/deployment'", target)
		return false
	}
	namespace, deployName := parts[0], parts[1]

	// 2. Check Deployment availability
	deploy, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deployName, metav1.GetOptions{})
	if err != nil {
		log.Printf("[K8s Error] Failed to fetch deployment %s: %v", deployName, err)
		return false
	}

	isHealthy := true

	// Evaluate if the deployment is fully available
	if deploy.Status.AvailableReplicas < deploy.Status.Replicas {
		log.Printf("[K8s Alert] Deployment unavailable: %s (%d/%d replicas active)", 
			deployName, deploy.Status.AvailableReplicas, deploy.Status.Replicas)
		isHealthy = false
	}

	// 3. Drill down into the Pods owned by this deployment
	// We dynamically convert the deployment's map selector into a string query
	selector := metav1.FormatLabelSelector(deploy.Spec.Selector)
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		log.Printf("[K8s Error] Failed to list pods for %s: %v", deployName, err)
		return isHealthy
	}

	for _, pod := range pods.Items {
		// Check for Pod failure states
		if pod.Status.Phase == "Failed" {
			log.Printf("[K8s Alert] Pod failed: %s", pod.Name)
			isHealthy = false
		}

		// Inspect individual container run states inside the pod
		for _, status := range pod.Status.ContainerStatuses {
			if status.RestartCount > 0 {
				log.Printf("[K8s Alert] Pod restarted: %s (Count: %d)", pod.Name, status.RestartCount)
				// A historic restart doesn't necessarily mean it's dead right now, 
				// but let's consider it unhealthy for simulation insight if it's currently crashlooping.
			}

			if status.State.Waiting != nil {
				reason := status.State.Waiting.Reason
				if reason == "CrashLoopBackOff" || reason == "ImagePullBackOff" {
					log.Printf("[K8s Alert] Pod trapped in failure loop: %s (Reason: %s)", pod.Name, reason)
					isHealthy = false
				}
			}
		}
	}

	return isHealthy
}

func checkHTTP(url string) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider 200-299 as healthy
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
