// Package config configures the application on start, exports config, initialization, etc
package config

// checkConfig checks the config for validity
func checkConfig() error {
	switch cfg.Logging.Encoding {
	case "json", "console":
		// no-op
	default:
		return ErrLoggingEncodingInvalid
	}

	// slurmabler checks
	if cfg.Slurm.Slurmabler.Image == "" {
		return ErrSlurmSlurmablerImageNotSet
	}

	if cfg.Slurm.Slurmabler.ServiceAccount == "" {
		return ErrSlurmSlurmablerServiceAccountNotSet
	}

	// munged checks
	if cfg.Slurm.Munged.Image == "" {
		return ErrSlurmMungedImageNotSet
	}

	// slurmctld checks
	if cfg.Slurm.Slurmctld.Image == "" {
		return ErrSlurmSlurmctldImageNotSet
	}

	// slurmd checks
	if cfg.Slurm.Slurmd.Image == "" {
		return ErrSlurmSlurmdImageNotSet
	}

	// slurm toolbox
	if cfg.Slurm.SlurmToolbox.Image == "" {
		return ErrSlurmSlurmToolboxImageNotSet
	}

	// mariadb
	if cfg.Slurm.MariaDB.Image == "" {
		return ErrSlurmMariaDBNotSet
	}

	// slurmdbd
	if cfg.Slurm.Slurmdbd.Image == "" {
		return ErrSlurmSlurmdbdImageNotSet
	}

	// slurmrestd
	if cfg.Slurm.Slurmrestd.Image == "" {
		return ErrSlurmSlurmrestdImageNotSet
	}

	return nil
}
