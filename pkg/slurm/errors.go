package slurm

import apierrors "k8s.io/apimachinery/pkg/api/errors"

func ignoreAlreadyExists(err error) error {
	if apierrors.IsAlreadyExists(err) {
		return nil
	}

	return err
}
