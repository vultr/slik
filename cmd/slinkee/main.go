// Package main is the v-cdn-server
package main

import (
	"context"
	"time"

	"github.com/vultr/slinkee/cmd/slinkee/config"
	"github.com/vultr/slinkee/cmd/slinkee/metrics"
	"github.com/vultr/slinkee/spec/helpers"
	"github.com/vultr/slinkee/spec/probes"
	"github.com/vultr/slinkee/spec/reconciler"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	name string = "slinkee"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() { //nolint
	_, err := config.NewConfig(name)
	if err != nil {
		logger, _ := zap.NewDevelopment()
		s := logger.Sugar()
		s.Fatal(err)
	}

	log := zap.L().Sugar()

	log.With(
		"context", "slinkee",
		"version", version,
		"commit", commit,
		"date", date,
	).Info()

	// Used for clean shutdown, catches signals
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		counter := 0
		for {
			helpers.Signals(cancel, &counter)
		}
	}()

	g, gCtx := errgroup.WithContext(ctx)

	metrics.NewMetrics()

	log.With(
		"context", "slinkee",
	).Info("initializing probes api")

	probe, err := probes.NewProbesAPI(name, config.GetProbesAPIListen(), config.GetProbesAPIPort())
	if err != nil {
		log.Fatal(err)
	}

	recon := reconciler.NewReconciler()

	// run http probes api
	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done():
				log.With(
					"context", "slinkee",
				).Info("probes: exited")

				return nil
			default:
				log.With(
					"context", "slinkee",
				).Info("probes: starting")

				if err2 := probe.Start(); err2 != nil {
					return err2
				}
			}
		}
	})

	// start reconcile loop
	g.Go(func() error {
		select {
		case <-gCtx.Done():
			log.With(
				"context", "slinkee",
			).Info("reconciler: exited")

			return nil
		default:
			log.With(
				"context", "slinkee",
			).Info("reconciler: starting")

			recon.Run()
		}

		return nil
	})

	probe.Success()

	// cleanly shutdown
	var start time.Time
	var delta time.Duration
	g.Go(func() error {
		<-gCtx.Done()
		start = time.Now()

		probe.Failed()

		if err := probe.Shutdown(); err != nil {
			log.With(
				"context", "slinkee",
			).Error(err)
		}

		recon.Shutdown()

		return nil
	})

	if err := g.Wait(); err != nil {
		log.With(
			"context", "slinkee",
		).Errorf("exit reason: %s", err)
	}

	delta = time.Since(start)

	log.With(
		"context", "slinkee",
	).Infof("shutdown in %s", delta.Round(time.Millisecond))
}
