package slurm

import (
	"context"
	"fmt"

	"github.com/AhmedTremo/slik/cmd/slik/config"
	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func buildToolboxDeployment(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	ds := client.AppsV1().Deployments(wl.Namespace)

	aff, err := mkAffinity(wl)
	if err != nil {
		return err
	}

	mungeCont := mkMungeContainer(wl)
	slurmToolboxCont := mkSlurmToolboxContainer(wl)

	log.Infof("munged container: %+v", *mungeCont)
	log.Infof("slurm-toolbox container: %+v", *slurmToolboxCont)

	if aff != nil {
		log.Infof("affinity: %+v", *aff)
	}

	slurmToolboxDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-slurm-toolbox", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          "slurm-toolbox",
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "slurm-toolbox",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wl.Name,
					Namespace: wl.Namespace,
					Labels: map[string]string{
						"app":                          "slurm-toolbox",
						"app.kubernetes.io/managed-by": "slik",
					},
				},
				Spec: v1.PodSpec{
					Affinity: aff,
					InitContainers: []v1.Container{
						*mungeCont,
					},
					Containers: []v1.Container{
						*slurmToolboxCont,
					},
					RestartPolicy:    v1.RestartPolicyAlways,
					ImagePullSecrets: []v1.LocalObjectReference{},
					Volumes: []v1.Volume{
						{
							Name: "shared-data",
							VolumeSource: v1.VolumeSource{
								EmptyDir: &v1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "munge",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: fmt.Sprintf("%s-munged", wl.Name),
									},
								},
							},
						},
						{
							Name: "slurm",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: fmt.Sprintf("%s-slurm", wl.Name),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	log.Infof("slurm_toolbox deployment: %+v", slurmToolboxDep)

	_, err2 := ds.Create(context.TODO(), slurmToolboxDep, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	log.Infof("slurm_toolbox deployments %s created", wl.Name)

	return nil
}

func mkSlurmToolboxContainer(wl *v1s.Slik) *v1.Container {
	c := v1.Container{
		Name:  "slurm-toolbox",
		Image: config.GetSlurmSlurmToolboxImage(),
	}

	c.VolumeMounts = []v1.VolumeMount{
		{
			Name:      "munge",
			MountPath: "/etc/munge",
		},
		{
			Name:      "slurm",
			MountPath: "/etc/slurm",
		},
		{
			Name:      "shared-data",
			MountPath: "/run/munge",
		},
	}

	c.Env = []v1.EnvVar{
		{
			Name:  "X_AhmedTremo_SLURM_ID",
			Value: wl.Name,
		},
	}

	c.Ports = []v1.ContainerPort{
		{
			Name:          "slurmctld",
			ContainerPort: 6817,
		},
		{
			Name:          "slurmd",
			ContainerPort: 6818,
		},
	}

	return &c
}
