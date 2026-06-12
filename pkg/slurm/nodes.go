package slurm

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
)

const (
	nodeLabelCPUs           = "slik.vultr.com/cpus"
	nodeLabelRealMemory     = "slik.vultr.com/real_memory"
	nodeLabelThreadsPerCore = "slik.vultr.com/threads_per_core"
)

func waitForSlurmableNodes(client kubernetes.Interface) error {
	log := zap.L().Sugar()
	deadline := time.Now().Add(time.Duration(SlurmablerWaitTimeoutSec) * time.Second)

	for {
		nodes, err := GetAllNodes(client)
		if err != nil {
			return err
		}

		pending := pendingSlurmableLabels(nodes.Items)
		if len(pending) == 0 {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for slurmabler labels on nodes: %v", pending)
		}

		log.Infof("waiting for slurmabler labels on nodes: %v", pending)
		time.Sleep(1 * time.Second)
	}
}

func pendingSlurmableLabels(nodes []corev1.Node) []string {
	pending := []string{}
	for i := range nodes {
		if !isSlurmableNode(&nodes[i]) || hasSlurmLabels(&nodes[i]) {
			continue
		}

		pending = append(pending, nodes[i].Name)
	}

	return pending
}

func slurmNodes(client kubernetes.Interface) ([]corev1.Node, error) {
	nodes, err := GetAllNodes(client)
	if err != nil {
		return nil, err
	}

	result := []corev1.Node{}
	for i := range nodes.Items {
		if isSlurmableNode(&nodes.Items[i]) && hasSlurmLabels(&nodes.Items[i]) {
			result = append(result, nodes.Items[i])
		}
	}

	return result, nil
}

func isSlurmableNode(node *corev1.Node) bool {
	if node.Spec.Unschedulable {
		return false
	}

	for i := range node.Spec.Taints {
		switch node.Spec.Taints[i].Effect {
		case corev1.TaintEffectNoSchedule, corev1.TaintEffectNoExecute:
			return false
		}
	}

	return true
}

func hasSlurmLabels(node *corev1.Node) bool {
	labels := node.GetLabels()
	if labels == nil {
		return false
	}

	_, hasCPUs := labels[nodeLabelCPUs]
	_, hasMemory := labels[nodeLabelRealMemory]
	_, hasThreadsPerCore := labels[nodeLabelThreadsPerCore]

	return hasCPUs && hasMemory && hasThreadsPerCore
}
