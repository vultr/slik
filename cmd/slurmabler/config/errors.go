// Package config configures the application on start, exports config, initialization, etc
package config

import "errors"

var (
	ErrPostgresSSLModeInvalid = errors.New("postgres.sslmode must be any of: disable, require, verify-ca, verify-full")

	ErrLoggingEncodingInvalid = errors.New("logging.encoding must be either json or console")

	// slurm
	ErrSlurmSlurmablerImageNotSet = errors.New("slurm.slurmabler.image not set")
	ErrSlurmMungedImageNotSet     = errors.New("slurm.munged.image not set")
	ErrSlurmSlurmctldImageNotSet  = errors.New("slurm.slurmctld.image not set")
	ErrSlurmSlurmdImageNotSet     = errors.New("slurm.slurmd.image not set")
	ErrSlurmNamespaceNotSet       = errors.New("slurm.namespace not set")
)
