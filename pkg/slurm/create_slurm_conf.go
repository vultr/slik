package slurm

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SlurmConf configuration for generation of slurm.conf
type SlurmConf struct {
	SlikName string

	SlurmdNodes []SlurmdNode
	Slurmdbd    bool
}

// SlurmdNode for generation of the nodes section in slurm.conf
type SlurmdNode struct {
	NodeName       string
	CPUs           int
	ThreadsPerCore int
	RealMemory     int
}

// NewSlurmConf bilds SlurmConf for templating out slurm.conf
func NewSlurmConf(client kubernetes.Interface, wl *v1s.Slik) (*SlurmConf, error) {
	log := zap.L().Sugar()

	nodes, err := GetAllNodes(client)
	if err != nil {
		return nil, err
	}

	var conf SlurmConf
	for i := range nodes.Items {
		labels := nodes.Items[i].GetLabels()
		cpusS, ok := labels["slik.AhmedTremo.com/cpus"]
		if !ok {
			continue
		}

		memoryS, ok := labels["slik.AhmedTremo.com/real_memory"]
		if !ok {
			continue
		}

		threadsPerCoreS, ok := labels["slik.AhmedTremo.com/threads_per_core"]
		if !ok {
			continue
		}

		cpus, err := strconv.Atoi(cpusS)
		if err != nil {
			return nil, err
		}

		memory, err := strconv.Atoi(memoryS)
		if err != nil {
			return nil, err
		}

		threadsPerCore, err := strconv.Atoi(threadsPerCoreS)
		if err != nil {
			return nil, err
		}

		log.Infof("Node: %s, CPU: %d, Memory: %d, ThreadsPerCore: %d", nodes.Items[i].Name, cpus, memory, threadsPerCore)

		conf.SlurmdNodes = append(conf.SlurmdNodes, SlurmdNode{
			NodeName:       nodes.Items[i].Name,
			CPUs:           cpus,
			ThreadsPerCore: threadsPerCore,
			RealMemory:     memory,
		})
	}

	conf.SlikName = wl.Name
	conf.Slurmdbd = wl.Spec.Slurmdbd

	log.Infof("slurmconf: %+v", conf)

	return &conf, nil
}

// buildSlurmconfConfigMap creates slurm.conf configmap with slurm.conf
func buildSlurmconfConfigMap(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	cm := client.CoreV1().ConfigMaps(wl.Namespace)

	conf, err := NewSlurmConf(client, wl)
	if err != nil {
		return err
	}

	tpl, err := template.New("slurm_conf").Funcs(
		template.FuncMap{"StringsJoin": strings.Join},
	).Parse(slurmConfTpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, *conf); err != nil {
		return err
	}

	name := fmt.Sprintf("%s-slurm", wl.Name)
	cmSpec := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Data: map[string]string{
			"slurm.conf": buf.String(),
		},
	}

	log.Infof("configmap (slurm.conf): %+v", cmSpec)

	_, err2 := cm.Create(context.TODO(), cmSpec, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	WaitForConfigMap(client, name, wl.Namespace)

	return nil
}
