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

func buildSlurmdbdDeployment(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	ds := client.AppsV1().Deployments(wl.Namespace)

	aff, err := mkAffinity(wl)
	if err != nil {
		return err
	}

	mungeCont := mkMungeContainer(wl)
	slurmdbdCont := mkSlurmdbdContainer(wl)

	log.Infof("munged container: %+v", *mungeCont)
	log.Infof("slurmdbd container: %+v", *slurmdbdCont)

	if aff != nil {
		log.Infof("affinity: %+v", *aff)
	}

	var defaultMode int32 = 0600

	slurmdbdDep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-slurmdbd", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          fmt.Sprintf("%s-slurmdbd", wl.Name),
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": fmt.Sprintf("%s-slurmdbd", wl.Name),
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wl.Name,
					Namespace: wl.Namespace,
					Labels: map[string]string{
						"app":                          fmt.Sprintf("%s-slurmdbd", wl.Name),
						"app.kubernetes.io/managed-by": "slik",
					},
				},
				Spec: v1.PodSpec{
					Hostname: fmt.Sprintf("%s-slurmdbd", wl.Name), // MUST be set or slurmdbd will NOT start
					Affinity: aff,
					InitContainers: []v1.Container{
						*mungeCont,
					},
					Containers: []v1.Container{
						*slurmdbdCont,
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
							Name: "slurmdbd",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: fmt.Sprintf("%s-slurmdbd", wl.Name),
									},
									DefaultMode: &defaultMode, // MUST be set to 0600 or will slurmdbd will not start
								},
							},
						},
					},
				},
			},
		},
	}

	log.Infof("slurmdbd deployment: %+v", slurmdbdDep)

	_, err2 := ds.Create(context.TODO(), slurmdbdDep, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	log.Infof("slurmdbd deployments %s created", wl.Name)

	return nil
}

func mkSlurmdbdContainer(wl *v1s.Slik) *v1.Container {
	c := v1.Container{
		Name:  "slurmdbd",
		Image: config.GetSlurmSlurmdbdImage(),
	}

	c.VolumeMounts = []v1.VolumeMount{
		{
			Name:      "munge",
			MountPath: "/etc/munge",
		},
		{
			Name:      "shared-data",
			MountPath: "/run/munge",
		},
		{
			Name:      "slurmdbd",
			MountPath: "/etc/slurm",
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
			Name:          "slurmdbd",
			ContainerPort: 6819,
		},
	}

	return &c
}

func buildSlurmdbdService(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	svc := client.CoreV1().Services(wl.Namespace)

	svcSpec := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-slurmdbd", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          fmt.Sprintf("%s-slurmdbd", wl.Name),
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "slurmdbd",
					Port:       6819,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromString("slurmdbd"),
				},
			},
			Selector: map[string]string{
				"app": fmt.Sprintf("%s-slurmdbd", wl.Name),
			},
		},
	}

	log.Infof("slurmdbd service: %+v", svcSpec)

	_, err2 := svc.Create(context.TODO(), svcSpec, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	log.Infof("slurmdbd service %s created", wl.Name)

	return nil
}
