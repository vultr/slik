package slurm

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SlurmDelete deletes slurm
func SlurmDelete(client kubernetes.Interface, name, namespace string) error {
	// slurmd deployments
	nodes, err := GetAllNodes(client)
	if err != nil {
		return err
	}

	for i := range nodes.Items {
		res := fmt.Sprintf("%s-%s", name, nodes.Items[i].Name)
		if err := ServiceDelete(client, res, namespace); err != nil {
			return err
		}

		if err := DeploymentDelete(client, res, namespace); err != nil {
			return err
		}
	}

	if err := DaemonSetDelete(client, fmt.Sprintf("%s-slurmabler", name), namespace); err != nil {
		return err
	}

	if err := DeploymentDelete(client, fmt.Sprintf("%s-slurm-toolbox", name), namespace); err != nil {
		return err
	}

	if err := DeploymentDelete(client, fmt.Sprintf("%s-slurmrestd", name), namespace); err != nil {
		return err
	}

	if err := ServiceDelete(client, fmt.Sprintf("%s-slurmrestd", name), namespace); err != nil {
		return err
	}

	if err := StatefulSetDelete(client, fmt.Sprintf("%s-mariadb", name), namespace); err != nil {
		return err
	}

	if err := ServiceDelete(client, fmt.Sprintf("%s-mariadb", name), namespace); err != nil {
		return err
	}

	if err := DeploymentDelete(client, fmt.Sprintf("%s-slurmdbd", name), namespace); err != nil {
		return err
	}

	if err := ServiceDelete(client, fmt.Sprintf("%s-slurmdbd", name), namespace); err != nil {
		return err
	}

	if err := DeploymentDelete(client, fmt.Sprintf("%s-slurmctld", name), namespace); err != nil {
		return err
	}

	if err := ServiceDelete(client, fmt.Sprintf("%s-slurmctld", name), namespace); err != nil {
		return err
	}

	// config layer
	mungedCM := fmt.Sprintf("%s-munged", name)
	if err := ConfigMapDelete(client, mungedCM, namespace); err != nil {
		return err
	}

	slurmCM := fmt.Sprintf("%s-slurm", name)
	if err := ConfigMapDelete(client, slurmCM, namespace); err != nil {
		return err
	}

	slurmdbdCM := fmt.Sprintf("%s-slurmdbd", name)
	if err := ConfigMapDelete(client, slurmdbdCM, namespace); err != nil {
		return err
	}

	if err := ConfigMapDelete(client, fmt.Sprintf("%s-mariadb-config", name), namespace); err != nil {
		return err
	}

	if err := ConfigMapDelete(client, fmt.Sprintf("%s-mariadb-init", name), namespace); err != nil {
		return err
	}

	switch namespace {
	case "default", "kube-system":
		// no-op, we don't touch default or kube-system namespaces
	default:
		if err := NamespaceDelete(client, namespace); err != nil {
			return err
		}
	}

	return nil
}

// DeploymentDelete deletes deployment if it exists
func DeploymentDelete(client kubernetes.Interface, name, namespace string) error {
	log := zap.L().Sugar()

	if DeploymentExists(client, name, namespace) {
		if err := client.AppsV1().Deployments(namespace).Delete(context.TODO(), name, v1.DeleteOptions{}); err != nil {
			return err
		}

		log.Infof("deployment %s deleted", name)
	}

	return nil
}

// DaemonSetDelete deletes DaemonSet if it exists
func DaemonSetDelete(client kubernetes.Interface, name, namespace string) error {
	log := zap.L().Sugar()

	if DaemonSetExists(client, name, namespace) {
		if err := client.AppsV1().DaemonSets(namespace).Delete(context.TODO(), name, v1.DeleteOptions{}); err != nil {
			return err
		}

		log.Infof("daemonset %s deleted", name)
	}

	return nil
}

// ConfigMapDelete deletes configmap if it exists
func ConfigMapDelete(client kubernetes.Interface, name, namespace string) error {
	log := zap.L().Sugar()

	if ConfigMapExists(client, name, namespace) {
		if err := client.CoreV1().ConfigMaps(namespace).Delete(context.TODO(), name, v1.DeleteOptions{}); err != nil {
			return err
		}

		log.Infof("configmap %s deleted", name)
	}

	return nil
}

// ServiceDelete deletes deployment if it exists
func ServiceDelete(client kubernetes.Interface, name, namespace string) error {
	log := zap.L().Sugar()

	if ServiceExists(client, name, namespace) {
		if err := client.CoreV1().Services(namespace).Delete(context.TODO(), name, v1.DeleteOptions{}); err != nil {
			return err
		}

		log.Infof("service %s deleted", name)
	}

	return nil
}

// StatefulSetDelete deletes statefulset if it exists
func StatefulSetDelete(client kubernetes.Interface, name, namespace string) error {
	log := zap.L().Sugar()

	if StatefulsetExists(client, name, namespace) {
		if err := client.AppsV1().StatefulSets(namespace).Delete(context.TODO(), name, v1.DeleteOptions{}); err != nil {
			return err
		}

		log.Infof("statefulset %s deleted", name)
	}

	return nil
}

// NamespaceDelete deletes namespace if it exists
func NamespaceDelete(client kubernetes.Interface, namespace string) error {
	log := zap.L().Sugar()

	if NamespaceExists(client, namespace) {
		if err := client.CoreV1().Namespaces().Delete(context.TODO(), namespace, v1.DeleteOptions{}); err != nil {
			return err
		}

		log.Infof("namespace %s deleted", namespace)
	}

	return nil
}
