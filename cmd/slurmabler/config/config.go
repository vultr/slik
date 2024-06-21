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
}

// NewConfig returns a Config struct that can be used to reference configuration
// NewConfig does the following:
//   - Runs initCLI (sets and read CLI switches)
//   - Runs initConfig (reads config from files)
func NewConfig(name string) (*Config, error) {
	// Setup CLI flags
	initCLI(&cfg)

	// initialize logging
	logfile := fmt.Sprintf("%s/slurmabler.log", "/app")

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
		Encoding:         "console",
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

	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

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
