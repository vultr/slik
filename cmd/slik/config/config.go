// Package config configures the application on start, exports config, initialization, etc
package config

import (
	"flag"
	"fmt"
	"net"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

var cfg Config

// Config is the CLI options wrapped in a struct
type Config struct {
	Name       string
	ConfigFile string

	Logging   Logging   `yaml:"logging"`
	ProbesAPI ProbesAPI `yaml:"probes_api"`

	Slurm Slurm `yaml:"slurm"`
}

// Logging logging definition
type Logging struct {
	Debug    bool   `yaml:"debug"`
	Path     string `yaml:"path"`
	Encoding string `yaml:"encoding"`
}

// RestAPI rest API definition
type RestAPI struct {
	Listen string `yaml:"listen"`
	Port   uint16 `yaml:"port"`
	APIKey string `yaml:"api_key"`
}

// ProbesAPI probes API definition
type ProbesAPI struct {
	Listen string `yaml:"listen"`
	Port   uint16 `yaml:"port"`
}

// Slurm config
type Slurm struct {
	Slurmabler   Slurmabler   `yaml:"slurmabler"`
	Munged       Munged       `yaml:"munged"`
	Slurmctld    Slurmctld    `yaml:"slurmctld"`
	Slurmd       Slurmd       `yaml:"slurmd"`
	SlurmToolbox SlurmToolbox `yaml:"slurm_toolbox"`
	MariaDB      MariaDB      `yaml:"mariadb"`
	Slurmdbd     Slurmdbd     `yaml:"slurmdbd"`
	Slurmrestd   Slurmrestd   `yaml:"slurmrestd"`
}

// Slurmabler config
type Slurmabler struct {
	Image string `yaml:"image"`

	ServiceAccount string `yaml:"service_account"`
}

// Munged config
type Munged struct {
	Image string `yaml:"image"`
}

// Slurmctld config
type Slurmctld struct {
	Image string `yaml:"image"`
}

// Slurmd config
type Slurmd struct {
	Image string `yaml:"image"`
}

// SlurmToolbox config
type SlurmToolbox struct {
	Image string `yaml:"image"`
}

// MariaDB config
type MariaDB struct {
	Image string `yaml:"image"`
}

// Slurmdbd config
type Slurmdbd struct {
	Image string `yaml:"image"`
}

// Slurmrestd config
type Slurmrestd struct {
	Image string `yaml:"image"`
}

// NewConfig returns a Config struct that can be used to reference configuration
// NewConfig does the following:
//   - Runs initCLI (sets and read CLI switches)
//   - Runs initConfig (reads config from files)
func NewConfig(name string) (*Config, error) {
	// Setup CLI flags
	initCLI(&cfg)

	// initialize environment
	if err := initConf(name, cfg.ConfigFile); err != nil {
		return nil, fmt.Errorf("config.NewConfig: %w", err)
	}

	if err := checkConfig(); err != nil {
		return nil, err
	}

	// initialize logging
	logfile := fmt.Sprintf("%s/slik.log", GetLoggingPath())

	if err := initLogging(logfile, true); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// initCLI initializes CLI switches
func initCLI(config *Config) {
	flag.StringVar(&config.ConfigFile, "config", "./config.yaml", "Path for the config.yaml configuration file")
	flag.Parse()
}

// initConf initializes the configuration
func initConf(name, cfgFile string) error {
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("os.ReadFile failed: %w", err)
	}

	if err1 := yaml.Unmarshal(data, &cfg); err1 != nil {
		return fmt.Errorf("yaml.Unmarshal failed: %w", err1)
	}

	cfg.Name = name
	cfg.ConfigFile = cfgFile

	return nil
}

// initLogging initializes logging
func initLogging(logfile string, stdout bool) error { //nolint
	var stdoutPaths []string
	var stderrPaths []string

	if stdout {
		stdoutPaths = []string{logfile, "stdout"}
		stderrPaths = []string{logfile, "stderr"}
	} else {
		stdoutPaths = []string{logfile}
		stderrPaths = []string{logfile}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	h, err := mainIP()
	if err != nil {
		return err
	}

	zapConfig := &zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Encoding:         cfg.Logging.Encoding,
		OutputPaths:      stdoutPaths,
		ErrorOutputPaths: stderrPaths,
		InitialFields: map[string]interface{}{
			"hostname": hostname,
			"host":     h,
			"pid":      os.Getpid(),
		},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "name",
			CallerKey:      "caller",
			FunctionKey:    "function",
			MessageKey:     "message",
			StacktraceKey:  "stack",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	if cfg.Logging.Debug {
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	zapLogger := zap.Must(zapConfig.Build())
	zap.ReplaceGlobals(zapLogger)

	return nil
}

func mainIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return "", err
	}

	return localAddr.IP.String(), nil
}
