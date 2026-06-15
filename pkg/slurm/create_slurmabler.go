package slurm

import (
	"fmt"

	"github.com/vultr/slik/cmd/slik/config"
	v1s "github.com/vultr/slik/pkg/api/types/v1"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func buildSlurmablerDaemonSet(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	aff, err := mkAffinity(wl)
	if err != nil {
		return err
	}

	slurmablerCont := mkSlurmablerContainer(wl)

	log.Infof("slurmabler container: %+v", *slurmablerCont)

	if aff != nil {
		log.Infof("affinity: %+v", *aff)
	}

	slurmDSSpec := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-slurmabler", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          "slurmabler",
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "slurmabler",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wl.Name,
					Namespace: wl.Namespace,
					Labels: map[string]string{
						"app":                          "slurmabler",
						"app.kubernetes.io/managed-by": "slik",
					},
				},
				Spec: v1.PodSpec{
					ServiceAccountName: config.GetSlurmSlurmablerServiceAccount(),
					Affinity:           aff,
					Containers: []v1.Container{
						*slurmablerCont,
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}

	log.Infof("slurmabler daemonset: %+v", slurmDSSpec)

	if err := applyDaemonSet(client, slurmDSSpec); err != nil {
		return err
	}

	log.Infof("slurmabler daemonset %s created", wl.Name)

	if err := waitForSlurmableNodes(client); err != nil {
		return err
	}

	return nil
}

func mkSlurmablerContainer(wl *v1s.Slik) *v1.Container {
	var root int64 = 0
	var privileged bool = true

	c := v1.Container{
		Name:  "slurmabler",
		Image: config.GetSlurmSlurmablerImage(),
	}

	c.VolumeMounts = []v1.VolumeMount{}

	c.Env = []v1.EnvVar{
		{
			Name:  "X_VULTR_SLURM_ID",
			Value: wl.Name,
		},
		{
			Name: "HOSTNAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{
					FieldPath: "spec.nodeName",
				},
			},
		},
	}

	c.SecurityContext = &v1.SecurityContext{
		RunAsUser:  &root,
		Privileged: &privileged,
	}

	return &c
}
