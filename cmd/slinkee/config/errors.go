// Package config configures the application on start, exports config, initialization, etc
package config

import "errors"

var (
	ErrLoggingEncodingInvalid = errors.New("logging.encoding must be either json or console")

	// slurm
	ErrSlurmSlurmablerImageNotSet          = errors.New("slurm.slurmabler.image not set")
	ErrSlurmSlurmablerServiceAccountNotSet = errors.New("slurm.slurmabler.service_account not set")
	ErrSlurmMungedImageNotSet              = errors.New("slurm.munged.image not set")
	ErrSlurmSlurmctldImageNotSet           = errors.New("slurm.slurmctld.image not set")
	ErrSlurmSlurmdImageNotSet              = errors.New("slurm.slurmd.image not set")
	ErrSlurmSlurmToolboxImageNotSet        = errors.New("slurm.slurm_toolbox.image not set")
	ErrSlurmMariaDBNotSet                  = errors.New("slurm.mariadb.image not set")
	ErrSlurmSlurmdbdImageNotSet            = errors.New("slurm.slurmdbd.image not set")
	ErrSlurmSlurmrestdImageNotSet          = errors.New("slurm.slurmrestd.image not set")
)
