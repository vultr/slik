// Package helpers are helpers primarily used by main
package helpers

import (
	"context"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"

	"go.uber.org/zap"
)

// Signals catches signals
func Signals(cancel context.CancelFunc, counter *int) {
	log := zap.L().Sugar()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGUSR1)

	sig := <-c
	log.With(
		"context", "helpers",
	).Infof("signal %s received", sig.String())

	switch sig {
	case syscall.SIGINT, syscall.SIGTERM:
		*counter++
		cancel()

		if *counter > 1 {
			log.With(
				"context", "helpers",
			).Infof("force terminating")

			os.Exit(0) //nolint
		}
	case syscall.SIGUSR1:
		if err := dumpHeap(); err != nil {
			log.With(
				"context", "helpers",
			).Error(err)
		}
	}
}

func dumpHeap() error {
	log := zap.L().Sugar()

	tmpfile, err := os.CreateTemp(".", "heap-profile-*.pprof")
	if err != nil {
		return err
	}
	defer tmpfile.Close()

	log.With(
		"context", "helpers",
	).Infof("Writing heap profile to: %q", tmpfile.Name())

	if err := pprof.WriteHeapProfile(tmpfile); err != nil {
		return err
	}

	return nil
}
