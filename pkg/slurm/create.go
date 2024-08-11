package slurm

import (
	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateSlurm launches a slurm cluster on the k8s cluster
func CreateSlurm(client kubernetes.Interface, wl *v1s.Slik) error {
	// namespace
	if !NamespaceExists(client, wl.Namespace) {
		if err := buildNamespace(client, wl); err != nil {
			return err
		}
	}

	// node labeler for slurm.conf generation
	if err := buildSlurmablerDaemonSet(client, wl); err != nil {
		return err
	}

	// munge.key
	if err := buildMungedConfigMap(client, wl); err != nil {
		return err
	}

	// slurm.conf
	if err := buildSlurmconfConfigMap(client, wl); err != nil {
		return err
	}

	// slurmdbd.conf
	if err := buildSlurmdbdConfigMap(client, wl); err != nil {
		return err
	}

	// mariadb
	if err := buildMariaDBConfigMap(client, wl); err != nil {
		return err
	}

	// slurmctld
	if err := buildSlurmctlDeployment(client, wl); err != nil {
		return err
	}

	if err := buildSlurmctlService(client, wl); err != nil {
		return err
	}

	// slurmd
	if err := buildSlurmdDeployments(client, wl); err != nil {
		return err
	}

	if err := buildSlurmdService(client, wl); err != nil {
		return err
	}

	// slurm-toolbox
	if err := buildToolboxDeployment(client, wl); err != nil {
		return err
	}

	// slurmdbd and mariadb
	if wl.Spec.Slurmdbd {
		// mariadb
		if err := buildMariaDBStatefulSet(client, wl); err != nil {
			return err
		}

		if err := buildMariaDBService(client, wl); err != nil {
			return err
		}

		// slurmdbd
		if err := buildSlurmdbdDeployment(client, wl); err != nil {
			return err
		}

		if err := buildSlurmdbdService(client, wl); err != nil {
			return err
		}
	}

	// slurmrestd
	if wl.Spec.Slurmdbd && wl.Spec.Slurmrestd {
		if err := buildSlurmrestdDeployment(client, wl); err != nil {
			return err
		}

		if err := buildSlurmrestdService(client, wl); err != nil {
			return err
		}
	}

	return nil
}

func mkAffinity(wl *v1s.Slik) (*v1.Affinity, error) {
	return nil, nil
}
