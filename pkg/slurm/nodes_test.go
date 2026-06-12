package slurm

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPendingSlurmableLabelsIgnoresBlockedNodes(t *testing.T) {
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "control-plane"},
			Spec: corev1.NodeSpec{
				Taints: []corev1.Taint{
					{Key: "node-role.kubernetes.io/control-plane", Effect: corev1.TaintEffectNoSchedule},
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "worker"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "labeled-worker",
				Labels: map[string]string{
					nodeLabelCPUs:           "2",
					nodeLabelRealMemory:     "1024",
					nodeLabelThreadsPerCore: "1",
				},
			},
		},
	}

	pending := pendingSlurmableLabels(nodes)
	if len(pending) != 1 || pending[0] != "worker" {
		t.Fatalf("expected only unlabeled worker to be pending, got %v", pending)
	}
}

func TestSlurmNodesReturnsOnlyEligibleLabeledNodes(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.NodeList{Items: []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "worker-1",
				Labels: map[string]string{
					nodeLabelCPUs:           "2",
					nodeLabelRealMemory:     "1024",
					nodeLabelThreadsPerCore: "1",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "worker-2"},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "control-plane",
				Labels: map[string]string{
					nodeLabelCPUs:           "2",
					nodeLabelRealMemory:     "1024",
					nodeLabelThreadsPerCore: "1",
				},
			},
			Spec: corev1.NodeSpec{
				Taints: []corev1.Taint{
					{Key: "node-role.kubernetes.io/control-plane", Effect: corev1.TaintEffectNoSchedule},
				},
			},
		},
	}})

	nodes, err := slurmNodes(client)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 1 || nodes[0].Name != "worker-1" {
		t.Fatalf("expected only eligible labeled node, got %v", nodes)
	}
}
