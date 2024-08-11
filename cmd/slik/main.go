// Package main is the v-cdn-server
package main

import (
	"context"
	"time"

	"github.com/AhmedTremo/slik/cmd/slik/config"
	"github.com/AhmedTremo/slik/cmd/slik/metrics"
	"github.com/AhmedTremo/slik/pkg/helpers"
	"github.com/AhmedTremo/slik/pkg/probes"
	"github.com/AhmedTremo/slik/pkg/reconciler"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	name string = "slik"
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
		"context", name,
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
		"context", name,
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
					"context", name,
				).Info("probes: exited")

				return nil
			default:
				log.With(
					"context", name,
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
				"context", name,
			).Info("reconciler: exited")

			return nil
		default:
			log.With(
				"context", name,
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
				"context", name,
			).Error(err)
		}

		recon.Shutdown()

		return nil
	})

	if err := g.Wait(); err != nil {
		log.With(
			"context", name,
		).Errorf("exit reason: %s", err)
	}

	delta = time.Since(start)

	log.With(
		"context", name,
	).Infof("shutdown in %s", delta.Round(time.Millisecond))
}
