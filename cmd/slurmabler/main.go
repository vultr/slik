package main

import (
	"fmt"
	"math"
	"os/exec"
	"time"

	"github.com/vultr/slinkee/cmd/slurmabler/config"
	"github.com/vultr/slinkee/pkg/connectors"
	"github.com/vultr/slinkee/pkg/labeler"
	"github.com/vultr/slinkee/pkg/slurm"

	"go.uber.org/zap"
)

const (
	name string = "slurmabler"
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
		"context", "main",
		"version", version,
		"commit", commit,
		"date", date,
	).Info()

	// get node config
	labeler, err := getNodeConfig()
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("%+v", *labeler)

	// apply labels to node
	clientset, err := connectors.GetKubernetesConn()
	if err != nil {
		log.Fatal(err)
	}

	node, err := slurm.GetNode(clientset, labeler.NodeName)
	if err != nil {
		log.Fatal(err)
	}

	labels := node.GetLabels()
	labels["slinkee.vultr.com/nodename"] = labeler.NodeName
	labels["slinkee.vultr.com/cpus"] = fmt.Sprintf("%d", labeler.CPUs)
	labels["slinkee.vultr.com/boards"] = fmt.Sprintf("%d", labeler.Boards)
	labels["slinkee.vultr.com/sockets_per_board"] = fmt.Sprintf("%d", labeler.SocketsPerBoard)
	labels["slinkee.vultr.com/cores_per_socket"] = fmt.Sprintf("%d", labeler.CoresPerSocket)
	labels["slinkee.vultr.com/threads_per_core"] = fmt.Sprintf("%d", labeler.ThreadsPerCore)
	labels["slinkee.vultr.com/real_memory"] = fmt.Sprintf("%d", labeler.RealMemory)

	node.Labels = labels

	if err := slurm.UpdateNode(clientset, node); err != nil {
		log.Fatal(err)
	}

	log.Info("sleeping forever...")

	time.Sleep(math.MaxInt64)
}

func getNodeConfig() (*labeler.Labels, error) {
	out, err := exec.Command("slurmd", "-C").Output()
	if err != nil {
		return nil, err
	}

	labeler := labeler.NewLabeler(string(out))

	return labeler, nil
}
