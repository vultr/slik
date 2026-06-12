package slurm

import (
	"context"
	"time"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const configMapWaitAttempts = 30

// SlurmExists returns true if the workload exists
func SlurmExists(client kubernetes.Interface, name, namespace string) bool {
	return DeploymentExists(client, name, namespace)
}

// DeploymentExists returns true if the deployment exists
func DeploymentExists(client kubernetes.Interface, name, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetDeployment(client, name, namespace)
		return err
	})
}

// DaemonSetExists returns true if the daemonset exists
func DaemonSetExists(client kubernetes.Interface, name, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetDaemonSet(client, name, namespace)
		return err
	})
}

// ConfigMapExists returns true if the configmap exists
func ConfigMapExists(client kubernetes.Interface, name, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetConfigMap(client, name, namespace)
		return err
	})
}

// ServiceExists returns true if the service exists
func ServiceExists(client kubernetes.Interface, name, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetService(client, name, namespace)
		return err
	})
}

// StatefulsetExists returns true if the statefulset exists
func StatefulsetExists(client kubernetes.Interface, name, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetStatefulSet(client, name, namespace)
		return err
	})
}

func resourceExists(get func() error) bool {
	if err := get(); err != nil {
		if !errors.IsNotFound(err) {
			zap.L().Sugar().Error(err)
		}

		return false
	}

	return true
}

// GetNamespace returns the namespace if it exists
func GetNamespace(client kubernetes.Interface, namespace string) (*v1.Namespace, error) {
	return client.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
}

// GetDeployment returns the deployment if it exists
func GetDeployment(client kubernetes.Interface, name, namespace string) (*appsv1.Deployment, error) {
	return client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetDaemonSet returns the daemonset if it exists
func GetDaemonSet(client kubernetes.Interface, name, namespace string) (*appsv1.DaemonSet, error) {
	return client.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetService returns the service if it exists
func GetService(client kubernetes.Interface, name, namespace string) (*v1.Service, error) {
	return client.CoreV1().Services(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetConfigMap returns the configmap if it exists
func GetConfigMap(client kubernetes.Interface, name, namespace string) (*v1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetStatefulSet returns the statefulset if it exists
func GetStatefulSet(client kubernetes.Interface, name, namespace string) (*appsv1.StatefulSet, error) {
	return client.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetNode gets a node
func GetNode(client kubernetes.Interface, name string) (*v1.Node, error) {
	return client.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
}

// WaitForConfigMap waits for configmap to exist before returning
func WaitForConfigMap(client kubernetes.Interface, name, namespace string) {
	log := zap.L().Sugar()

	for i := 0; i < configMapWaitAttempts; i++ {
		if ConfigMapExists(client, name, namespace) {
			return
		}

		log.Debugf("configmap %s still not exist", name)

		time.Sleep(1 * time.Second)
	}
}

// GetDeploymentStatus returns the job status
func GetDeploymentStatus(client kubernetes.Interface, name, namespace string) string {
	log := zap.L().Sugar()

	job, err := GetDeployment(client, name, namespace)
	if err != nil {
		log.With(
			"workload", name,
		).Error(err)

		return WorkloadStatusUnknown
	}

	if len(job.Status.Conditions) > 0 {
		return job.Status.Conditions[0].Message
	}

	return WorkloadStatusUnknown
}

// PodExists returns true if post exists
func PodExists(client kubernetes.Interface, name, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetPod(client, name, namespace)
		return err
	})
}

// NamespaceExists returns true if the namespace exists
func NamespaceExists(client kubernetes.Interface, namespace string) bool {
	return resourceExists(func() error {
		_, err := GetNamespace(client, namespace)
		return err
	})
}

// GetPod returns the pod if it exists
func GetPod(client kubernetes.Interface, name, namespace string) (*v1.Pod, error) {
	return client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

// GetPodStatus returns the job status
func GetPodStatus(client kubernetes.Interface, name, namespace string) string {
	log := zap.L().Sugar()

	job, err := GetPod(client, name, namespace)
	if err != nil {
		log.With(
			"workload", name,
		).Error(err)

		return WorkloadStatusUnknown
	}

	if len(job.Status.Conditions) > 0 {
		return job.Status.Conditions[0].Message
	}

	return WorkloadStatusUnknown
}

// GetAllNodes returns all nodes
func GetAllNodes(client kubernetes.Interface) (*v1.NodeList, error) {
	return client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
}
