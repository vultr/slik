package slurm

import (
	"context"

	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// buildNamespace creates a kubernetes namespace
func buildNamespace(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	ns := client.CoreV1().Namespaces()

	nsSpec := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: wl.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "slik",
			},
		},
	}

	log.Infof("namespace: %+v", nsSpec)

	_, err2 := ns.Create(context.TODO(), nsSpec, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	return nil
}
