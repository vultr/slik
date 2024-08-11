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
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

func buildSlurmrestdDeployment(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	ds := client.AppsV1().Deployments(wl.Namespace)

	aff, err := mkAffinity(wl)
	if err != nil {
		return err
	}

	mungeCont := mkMungeContainer(wl)
	slurmrestdCont := mkSlurmrestdContainer(wl)

	log.Infof("munged container: %+v", *mungeCont)
	log.Infof("slurmrestd container: %+v", *slurmrestdCont)

	if aff != nil {
		log.Infof("affinity: %+v", *aff)
	}

	slurmrestdDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-slurmrestd", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          fmt.Sprintf("%s-slurmrestd", wl.Name),
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": fmt.Sprintf("%s-slurmrestd", wl.Name),
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wl.Name,
					Namespace: wl.Namespace,
					Labels: map[string]string{
						"app":                          fmt.Sprintf("%s-slurmrestd", wl.Name),
						"app.kubernetes.io/managed-by": "slik",
					},
				},
				Spec: v1.PodSpec{
					Affinity: aff,
					InitContainers: []v1.Container{
						*mungeCont,
					},
					Containers: []v1.Container{
						*slurmrestdCont,
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

	log.Infof("slurmrestd deployment: %+v", slurmrestdDep)

	_, err2 := ds.Create(context.TODO(), slurmrestdDep, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	log.Infof("slurmrestd deployments %s created", wl.Name)

	return nil
}

func buildSlurmrestdService(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	svc := client.CoreV1().Services(wl.Namespace)

	svcSpec := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-slurmrestd", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          fmt.Sprintf("%s-slurmrestd", wl.Name),
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "slurmrestd",
					Port:       6820,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromString("slurmrestd"),
				},
			},
			Selector: map[string]string{
				"app": fmt.Sprintf("%s-slurmrestd", wl.Name),
			},
		},
	}

	log.Infof("slurmrestd service: %+v", svcSpec)

	_, err2 := svc.Create(context.TODO(), svcSpec, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	log.Infof("slurmrestd service %s created", wl.Name)

	return nil
}

func mkSlurmrestdContainer(wl *v1s.Slik) *v1.Container {
	c := v1.Container{
		Name:  "slurmrestd",
		Image: config.GetSlurmSlurmrestdImage(),
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
			Name:          "slurmrestd",
			ContainerPort: 6820,
		},
	}

	// slurmrestd will NOT run as root, pull uid/gid from container for the slurm user/group
	var runAsUser int64 = 64030
	var runasGroup int64 = 64030
	c.SecurityContext = &v1.SecurityContext{
		RunAsUser:  &runAsUser,
		RunAsGroup: &runasGroup,
	}

	return &c
}
