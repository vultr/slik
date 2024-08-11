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

func buildSlurmdService(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	svc := client.CoreV1().Services(wl.Namespace)

	nodes, err := GetAllNodes(client)
	if err != nil {
		return err
	}

	for i := range nodes.Items {
		svcSpec := &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", wl.Name, nodes.Items[i].Name),
				Namespace: wl.Namespace,
				Labels: map[string]string{
					"app":                          fmt.Sprintf("%s-slurmd", wl.Name),
					"app.kubernetes.io/managed-by": "slik",
				},
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeClusterIP,
				Ports: []v1.ServicePort{
					{
						Name:       "slurmd",
						Port:       6818,
						Protocol:   v1.ProtocolTCP,
						TargetPort: intstr.FromString("slurmd"),
					},
				},
				Selector: map[string]string{
					"app":  fmt.Sprintf("%s-slurmd", wl.Name),
					"host": nodes.Items[i].Name,
				},
			},
		}

		log.Infof("slurmd service: %+v", svcSpec)

		_, err2 := svc.Create(context.TODO(), svcSpec, metav1.CreateOptions{})
		if err2 != nil {
			return err2
		}

		log.Infof("slurmd service %s created", wl.Name)
	}

	return nil
}

func buildSlurmdDeployments(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	nodes, err := GetAllNodes(client)
	if err != nil {
		return err
	}

	for i := range nodes.Items {
		ds := client.AppsV1().Deployments(wl.Namespace)

		aff, err := mkAffinity(wl)
		if err != nil {
			return err
		}

		mungeCont := mkMungeContainer(wl)
		slurmdCont := mkSlurmdContainer(wl)

		log.Infof("munged container: %+v", *mungeCont)
		log.Infof("slurmd container: %+v", *slurmdCont)

		if aff != nil {
			log.Infof("affinity: %+v", *aff)
		}

		slurmDepSpec := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s", wl.Name, nodes.Items[i].Name),
				Namespace: wl.Namespace,
				Labels: map[string]string{
					"app":                          fmt.Sprintf("%s-slurmd", wl.Name),
					"app.kubernetes.io/managed-by": "slik",
					"host":                         nodes.Items[i].Name,
				},
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app":  fmt.Sprintf("%s-slurmd", wl.Name),
						"host": nodes.Items[i].Name,
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name:      wl.Name,
						Namespace: wl.Namespace,
						Labels: map[string]string{
							"app":                          fmt.Sprintf("%s-slurmd", wl.Name),
							"app.kubernetes.io/managed-by": "slik",
							"host":                         nodes.Items[i].Name,
						},
					},
					Spec: v1.PodSpec{
						Hostname: fmt.Sprintf("%s-%s", wl.Name, nodes.Items[i].Name),
						Affinity: aff,
						InitContainers: []v1.Container{
							*mungeCont,
						},
						Containers: []v1.Container{
							*slurmdCont,
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
						NodeSelector: map[string]string{
							"kubernetes.io/hostname": nodes.Items[i].Name,
						},
					},
				},
			},
		}

		log.Infof("slurmd deployment: %+v", slurmDepSpec)

		_, err2 := ds.Create(context.TODO(), slurmDepSpec, metav1.CreateOptions{})
		if err2 != nil {
			return err2
		}
	}

	log.Infof("slurmd deployments %s created", wl.Name)

	return nil
}

func mkSlurmdContainer(wl *v1s.Slik) *v1.Container {
	c := v1.Container{
		Name:  "slurmd",
		Image: config.GetSlurmSlurmdImage(),
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
			Name:          "slurmd",
			ContainerPort: 6818,
		},
	}

	var root int64 = 0
	priv := true
	c.SecurityContext = &v1.SecurityContext{
		RunAsUser:  &root,
		Privileged: &priv,
	}

	return &c
}
