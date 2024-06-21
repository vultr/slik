// Package config configures the application on start, exports config, initialization, etc
package config

// GetConfig returns config
func GetConfig() *Config {
	return &cfg
}

// GetProbesAPIListen returns probes api listen addr
func GetProbesAPIListen() string {
	return cfg.ProbesAPI.Listen
}

// GetProbesAPIPort returns probes api listen port
func GetProbesAPIPort() uint16 {
	return cfg.ProbesAPI.Port
}

// GetLoggingPath returns logging path
func GetLoggingPath() string {
	return cfg.Logging.Path
}

// GetSlurmSlurmablerImage returns the slurmabler image
func GetSlurmSlurmablerImage() string {
	return cfg.Slurm.Slurmabler.Image
}

// GetSlurmSlurmablerServiceAccount returns the slurmabler service account
func GetSlurmSlurmablerServiceAccount() string {
	return cfg.Slurm.Slurmabler.ServiceAccount
}

// GetSlurmMungedImage returns the munge image
func GetSlurmMungedImage() string {
	return cfg.Slurm.Munged.Image
}

// GetSlurmSlurmctldImage returns the slurmctl image
func GetSlurmSlurmctldImage() string {
	return cfg.Slurm.Slurmctld.Image
}

// GetSlurmSlurmdImage returns the slurmd image
func GetSlurmSlurmdImage() string {
	return cfg.Slurm.Slurmd.Image
}

// GetSlurmSlurmToolboxImage returns slurm toolbox image
func GetSlurmSlurmToolboxImage() string {
	return cfg.Slurm.SlurmToolbox.Image
}

// GetSlurmMariaDBImage returns mariadb image
func GetSlurmMariaDBImage() string {
	return cfg.Slurm.MariaDB.Image
}

// GetSlurmSlurmdbdImage returns slurmdbd image
func GetSlurmSlurmdbdImage() string {
	return cfg.Slurm.Slurmdbd.Image
}

// GetSlurmSlurmrestdImage returns slurmrestd image
func GetSlurmSlurmrestdImage() string {
	return cfg.Slurm.Slurmrestd.Image
}
