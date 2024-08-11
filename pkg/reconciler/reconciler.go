package reconciler

import (
	"context"
	"time"

	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"
	"github.com/AhmedTremo/slik/pkg/connectors"
	"github.com/AhmedTremo/slik/pkg/slurm"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reconciler type
type Reconciler struct {
	Interval int

	shutdown bool
}

// Run runs the reconciler loop
func (r *Reconciler) Run() {
	log := zap.L().Sugar()

	counter := 0
	for {
		counter++

		if r.shutdown {
			return
		}

		if (counter%r.Interval) == 0 || counter == 1 {
			if err := loop(); err != nil {
				log.Error(err)
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// Shutdown exits the primary loop
func (r *Reconciler) Shutdown() {
	r.shutdown = true
}

// NewReconciler creates a new reconciler
func NewReconciler() *Reconciler {
	return &Reconciler{
		Interval: LoopInterval,
	}
}

func loop() error {
	log := zap.L().Sugar()

	log.Info("reconciling...")

	slurmcs, err := connectors.GetSlikClientset()
	if err != nil {
		return err
	}

	sliks, err := slurmcs.Slik(context.TODO()).List(v1.ListOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	for i := range sliks.Items {
		log.Infof("on slik cluster: %s, %+v",
			sliks.Items[i].Name,
			sliks.Items[i],
		)

		s := sliks.Items[i]

		if s.DeletionTimestamp != nil { // delete resource if it's not null
			log.Infof("deleting slurm cluster: %s", s.Name)

			cs, err := connectors.GetKubernetesConn()
			if err != nil {
				log.Error(err)

				continue
			}

			if err := slurm.SlurmDelete(cs, s.Name, s.Spec.Namespace); err != nil {
				log.Error(err)

				continue
			}

			s.ObjectMeta.Finalizers = []string{}

			if _, err := slurmcs.Slik(context.TODO()).Update(&s, v1.UpdateOptions{}); err != nil {
				log.Error(err)

				continue
			}

			continue // item is deleted, next pls
		}

		switch s.Status.State {
		case "":
			log.Infof("slurm cluster initializing: %s", s.Name)

			switch checks(&s) {
			case true:
				s.Status.State = StatePending
				s2, err := slurmcs.Slik(context.TODO()).UpdateStatus(&s, v1.UpdateOptions{})
				if err != nil {
					log.Error(err)

					continue
				}

				s2.ObjectMeta.Finalizers = []string{
					"sliks.hpc.AhmedTremo.com",
				}

				if _, err := slurmcs.Slik(context.TODO()).Update(s2, v1.UpdateOptions{}); err != nil {
					log.Error(err)

					continue
				}
			case false:
				s.Status.State = StateFailed

				if _, err := slurmcs.Slik(context.TODO()).UpdateStatus(&s, v1.UpdateOptions{}); err != nil {
					log.Error(err)

					continue
				}
			}
		case StatePending:
			log.Infof("creating slurm cluster: %s", s.Name)

			cs, err := connectors.GetKubernetesConn()
			if err != nil {
				log.Error(err)

				continue
			}

			if err := slurm.CreateSlurm(cs, &s); err != nil {
				log.Error(err)

				continue
			}

			s.Status.State = StateActive
			if _, err := slurmcs.Slik(context.TODO()).UpdateStatus(&s, v1.UpdateOptions{}); err != nil {
				log.Error(err)

				continue
			}
		case StateActive:

		case StateFailed:
			log.Infof("checking failed slurm cluster: %s", s.Name)

			switch checks(&s) {
			case true:
				s.Status.State = ""
				_, err := slurmcs.Slik(context.TODO()).UpdateStatus(&s, v1.UpdateOptions{})
				if err != nil {
					log.Error(err)

					continue
				}
			case false:
				log.Infof("slurm cluster %s still failing checks", s.Name)
			}
		}
	}

	return nil
}

// checks returns true if all checks pass
func checks(s *v1s.Slik) bool {
	log := zap.L().Sugar()

	// these checks only matter if a db is in use
	if s.Spec.Slurmdbd {
		q, err := resource.ParseQuantity(s.Spec.MariaDB.StorageSize)
		if err != nil {
			log.Error(err)
			log.Warnf("mariadb.storage_size %s is not valid, setting cluster to failed state", s.Spec.MariaDB.StorageSize)

			return false
		}

		q2, _ := resource.ParseQuantity("45G")
		if q.Cmp(q2) == -1 {
			log.Warnf("mariadb.storage_size must be at least 45G")

			return false
		}
	}

	return true
}
