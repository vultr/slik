package slurm

import (
	"fmt"

	"github.com/vultr/slik/cmd/slik/config"
	v1s "github.com/vultr/slik/pkg/api/types/v1"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// buildSlurmconfConfigMap creates mariadb configmap
func buildMariaDBConfigMap(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	cmCfgSpec := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb-config", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Data: map[string]string{
			"overrides.cnf": overridesCnf,
		},
	}

	log.Infof("configmap (mariadb-config): %+v", cmCfgSpec)

	if err := applyConfigMap(client, cmCfgSpec); err != nil {
		return err
	}

	WaitForConfigMap(client, fmt.Sprintf("%s-mariadb-config", wl.Name), wl.Namespace)

	cmInitSpec := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb-init", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Data: map[string]string{
			"slurm-init.sql": slurmInit,
		},
	}

	log.Infof("configmap (mariadb-init): %+v", cmInitSpec)

	if err := applyConfigMap(client, cmInitSpec); err != nil {
		return err
	}

	WaitForConfigMap(client, fmt.Sprintf("%s-mariadb-init", wl.Name), wl.Namespace)

	return nil
}

func buildMariaDBStatefulSet(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	aff, err := mkAffinity(wl)
	if err != nil {
		return err
	}

	mariaDBCont := mkMariaDBContainer(wl)
	annotations := configChecksumAnnotations(client, wl.Namespace,
		fmt.Sprintf("%s-mariadb-config", wl.Name),
		fmt.Sprintf("%s-mariadb-init", wl.Name),
	)

	log.Infof("mariadb container: %+v", *mariaDBCont)

	if aff != nil {
		log.Infof("affinity: %+v", *aff)
	}

	mariadbSTS := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          "mariadb",
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "mariadb",
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "mariadb",
					Namespace:   wl.Namespace,
					Annotations: annotations,
					Labels: map[string]string{
						"app":                          "mariadb",
						"app.kubernetes.io/managed-by": "slik",
					},
				},
				Spec: v1.PodSpec{
					Affinity: aff,
					Containers: []v1.Container{
						*mariaDBCont,
					},
					RestartPolicy:    v1.RestartPolicyAlways,
					ImagePullSecrets: []v1.LocalObjectReference{},
					Volumes: []v1.Volume{
						{
							Name: "mariadb-init",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: fmt.Sprintf("%s-mariadb-init", wl.Name),
									},
								},
							},
						},
						{
							Name: "mariadb-config",
							VolumeSource: v1.VolumeSource{
								ConfigMap: &v1.ConfigMapVolumeSource{
									LocalObjectReference: v1.LocalObjectReference{
										Name: fmt.Sprintf("%s-mariadb-config", wl.Name),
									},
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: fmt.Sprintf("%s-mariadb", wl.Name),
					},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						StorageClassName: &wl.Spec.MariaDB.StorageClass,
						Resources: v1.VolumeResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceName(v1.ResourceStorage): resource.MustParse(wl.Spec.MariaDB.StorageSize),
							},
						},
					},
				},
			},
		},
	}

	log.Infof("mariadb statefulset: %+v", mariadbSTS)

	if err := applyStatefulSet(client, mariadbSTS); err != nil {
		return err
	}

	pvcName := fmt.Sprintf("%s-mariadb-%s-mariadb-0", wl.Name, wl.Name)
	if err := updatePVCStorage(client, pvcName, wl.Namespace, wl.Spec.MariaDB.StorageSize); err != nil {
		return err
	}

	log.Infof("mariadb statefulset %s created", wl.Name)

	return nil
}

func mkMariaDBContainer(wl *v1s.Slik) *v1.Container {
	c := v1.Container{
		Name:  "mariadb",
		Image: config.GetSlurmMariaDBImage(),
	}

	c.VolumeMounts = []v1.VolumeMount{
		{
			Name:      fmt.Sprintf("%s-mariadb", wl.Name),
			MountPath: "/var/lib/mysql",
		},
		{
			Name:      "mariadb-init",
			MountPath: "/docker-entrypoint-initdb.d",
		},
		{
			Name:      "mariadb-config",
			MountPath: "/etc/mysql/conf.d",
		},
	}

	c.Env = []v1.EnvVar{
		{
			Name:  "X_VULTR_SLURM_ID",
			Value: wl.Name,
		},
		{
			Name:  "MARIADB_ALLOW_EMPTY_ROOT_PASSWORD",
			Value: "true",
		},
		{
			Name:  "MARIADB_DATABASE",
			Value: "slurmdbd",
		},
		{
			Name:  "MARIADB_USER",
			Value: "slurm",
		},
		{
			Name:  "MARIADB_PASSWORD",
			Value: "slurm",
		},
	}

	c.Ports = []v1.ContainerPort{
		{
			Name:          "mariadb",
			ContainerPort: 3306,
		},
	}

	return &c
}

func buildMariaDBService(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	svcSpec := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-mariadb", wl.Name),
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app":                          "mariadb",
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeClusterIP,
			Ports: []v1.ServicePort{
				{
					Name:       "mariadb",
					Port:       3306,
					Protocol:   v1.ProtocolTCP,
					TargetPort: intstr.FromString("mariadb"),
				},
			},
			Selector: map[string]string{
				"app": "mariadb",
			},
		},
	}

	log.Infof("mariadb service: %+v", svcSpec)

	if err := applyService(client, svcSpec); err != nil {
		return err
	}

	log.Infof("mariadb service %s created", wl.Name)

	return nil
}
