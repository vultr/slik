package slurm

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
