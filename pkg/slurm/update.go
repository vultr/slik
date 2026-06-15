package slurm

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// UpdateNode updates a node
func UpdateNode(client kubernetes.Interface, node *v1.Node) error {
	for {
		_, err := client.CoreV1().Nodes().Update(context.TODO(), node, metav1.UpdateOptions{})
		if err != nil {
			if errors.IsConflict(err) {
				time.Sleep(time.Duration(ConflictRetryIntervalSec) * time.Second)

				continue
			} else if errors.IsNotFound(err) {
				return err
			}
		}

		return nil
	}
}

func configChecksumAnnotations(client kubernetes.Interface, namespace string, names ...string) map[string]string {
	annotations := map[string]string{}
	for _, name := range names {
		cm, err := GetConfigMap(client, name, namespace)
		if err != nil {
			continue
		}

		h := sha256.New()
		keys := make([]string, 0, len(cm.Data))
		for key := range cm.Data {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			h.Write([]byte(key))
			h.Write([]byte(cm.Data[key]))
		}

		binaryKeys := make([]string, 0, len(cm.BinaryData))
		for key := range cm.BinaryData {
			binaryKeys = append(binaryKeys, key)
		}
		sort.Strings(binaryKeys)

		for _, key := range binaryKeys {
			h.Write([]byte(key))
			h.Write(cm.BinaryData[key])
		}

		annotations[fmt.Sprintf("slik.vultr.com/checksum-%s", name)] = hex.EncodeToString(h.Sum(nil))
	}

	return annotations
}

func applyConfigMap(client kubernetes.Interface, desired *v1.ConfigMap) error {
	cm := client.CoreV1().ConfigMaps(desired.Namespace)
	existing, err := cm.Get(context.TODO(), desired.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = cm.Create(context.TODO(), desired, metav1.CreateOptions{})
			return err
		}

		return err
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Data = desired.Data
	existing.BinaryData = desired.BinaryData
	_, err = cm.Update(context.TODO(), existing, metav1.UpdateOptions{})

	return err
}

func applyDeployment(client kubernetes.Interface, desired *appsv1.Deployment) error {
	dep := client.AppsV1().Deployments(desired.Namespace)
	existing, err := dep.Get(context.TODO(), desired.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = dep.Create(context.TODO(), desired, metav1.CreateOptions{})
			return err
		}

		return err
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Spec.Replicas = desired.Spec.Replicas
	existing.Spec.Template = desired.Spec.Template
	existing.Spec.Strategy = desired.Spec.Strategy
	_, err = dep.Update(context.TODO(), existing, metav1.UpdateOptions{})

	return err
}

func applyDaemonSet(client kubernetes.Interface, desired *appsv1.DaemonSet) error {
	ds := client.AppsV1().DaemonSets(desired.Namespace)
	existing, err := ds.Get(context.TODO(), desired.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = ds.Create(context.TODO(), desired, metav1.CreateOptions{})
			return err
		}

		return err
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Spec.Template = desired.Spec.Template
	existing.Spec.UpdateStrategy = desired.Spec.UpdateStrategy
	_, err = ds.Update(context.TODO(), existing, metav1.UpdateOptions{})

	return err
}

func applyService(client kubernetes.Interface, desired *v1.Service) error {
	svc := client.CoreV1().Services(desired.Namespace)
	existing, err := svc.Get(context.TODO(), desired.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = svc.Create(context.TODO(), desired, metav1.CreateOptions{})
			return err
		}

		return err
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Spec.Ports = desired.Spec.Ports
	existing.Spec.Selector = desired.Spec.Selector
	existing.Spec.Type = desired.Spec.Type
	_, err = svc.Update(context.TODO(), existing, metav1.UpdateOptions{})

	return err
}

func applyStatefulSet(client kubernetes.Interface, desired *appsv1.StatefulSet) error {
	sts := client.AppsV1().StatefulSets(desired.Namespace)
	existing, err := sts.Get(context.TODO(), desired.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err = sts.Create(context.TODO(), desired, metav1.CreateOptions{})
			return err
		}

		return err
	}

	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations
	existing.Spec.Replicas = desired.Spec.Replicas
	existing.Spec.Template = desired.Spec.Template
	existing.Spec.UpdateStrategy = desired.Spec.UpdateStrategy
	_, err = sts.Update(context.TODO(), existing, metav1.UpdateOptions{})

	return err
}

func waitForDeploymentAvailable(client kubernetes.Interface, namespace, name string) error {
	deadline := time.Now().Add(time.Duration(SlurmablerWaitTimeoutSec) * time.Second)
	for {
		dep, err := GetDeployment(client, name, namespace)
		if err != nil {
			return err
		}

		if dep.Status.AvailableReplicas > 0 {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for deployment %s/%s to become available", namespace, name)
		}

		time.Sleep(1 * time.Second)
	}
}

func updatePVCStorage(client kubernetes.Interface, name, namespace, storageSize string) error {
	if storageSize == "" {
		return nil
	}

	desired, err := resource.ParseQuantity(storageSize)
	if err != nil {
		return err
	}

	pvc, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}

		return err
	}

	current := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	if current.Cmp(desired) >= 0 {
		return nil
	}

	pvc.Spec.Resources.Requests[v1.ResourceStorage] = desired
	_, err = client.CoreV1().PersistentVolumeClaims(namespace).Update(context.TODO(), pvc, metav1.UpdateOptions{})

	return err
}

func reconcileDisabledComponents(client kubernetes.Interface, name, namespace string, slurmdbd, slurmrestd bool) error {
	if !slurmrestd {
		if err := DeploymentDelete(client, fmt.Sprintf("%s-slurmrestd", name), namespace); err != nil {
			return err
		}

		if err := ServiceDelete(client, fmt.Sprintf("%s-slurmrestd", name), namespace); err != nil {
			return err
		}
	}

	if slurmdbd {
		return nil
	}

	if err := DeploymentDelete(client, fmt.Sprintf("%s-slurmdbd", name), namespace); err != nil {
		return err
	}

	if err := ServiceDelete(client, fmt.Sprintf("%s-slurmdbd", name), namespace); err != nil {
		return err
	}

	if err := StatefulSetDelete(client, fmt.Sprintf("%s-mariadb", name), namespace); err != nil {
		return err
	}

	if err := ServiceDelete(client, fmt.Sprintf("%s-mariadb", name), namespace); err != nil {
		return err
	}

	for _, cm := range []string{
		fmt.Sprintf("%s-slurmdbd", name),
		fmt.Sprintf("%s-mariadb-config", name),
		fmt.Sprintf("%s-mariadb-init", name),
	} {
		if err := ConfigMapDelete(client, cm, namespace); err != nil {
			return err
		}
	}

	return nil
}
