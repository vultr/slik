package slurm

import (
	b64 "encoding/base64"
	"fmt"

	"github.com/vultr/slik/cmd/slik/config"
	v1s "github.com/vultr/slik/pkg/api/types/v1"
	"github.com/vultr/slik/pkg/munge"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// buildMungedConfigMap creates munged configmap with the munge.key
func buildMungedConfigMap(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	name := fmt.Sprintf("%s-munged", wl.Name)
	if ConfigMapExists(client, name, wl.Namespace) {
		return nil
	}

	mk, err := munge.NewMungeKey()
	if err != nil {
		return err
	}

	mungeKeyDec, err := b64.StdEncoding.DecodeString(mk.SecretBase64())
	if err != nil {
		return err
	}

	cmSpec := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		BinaryData: map[string][]byte{
			"munge.key": mungeKeyDec,
		},
	}

	log.Infof("configmap (munge): %+v", cmSpec)

	if err := applyConfigMap(client, cmSpec); err != nil {
		return err
	}

	WaitForConfigMap(client, name, wl.Namespace)

	return nil
}

func mkMungeContainer(wl *v1s.Slik) *v1.Container {
	c := v1.Container{
		Name:  "munged",
		Image: config.GetSlurmMungedImage(),
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
	}

	c.Env = []v1.EnvVar{
		{
			Name:  "X_VULTR_SLURM_ID",
			Value: wl.Name,
		},
	}

	// Always must be set for sidecar containers, munged is a sidecar container
	// munged MUST ALWAYS be running for slurmctld, slurmd, etc
	// it MUST be started BEFORE any other slurm procs
	always := v1.ContainerRestartPolicyAlways
	c.RestartPolicy = &always

	return &c
}
