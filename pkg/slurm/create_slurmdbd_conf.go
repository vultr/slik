package slurm

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"

	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SlurmdbConf slurmdb conf
type SlurmdbConf struct {
	SlikName string

	User string
	Pass string
}

// NewSlurmdbdConf bilds SlurmdbConf for templating out slurmdb.conf
func NewSlurmdbdConf(client kubernetes.Interface, wl *v1s.Slik) (*SlurmdbConf, error) {
	log := zap.L().Sugar()

	var conf SlurmdbConf
	conf.SlikName = wl.Name
	conf.User = "slurm"
	conf.Pass = "slurm"

	log.Infof("slurmdbconf: %+v", conf)

	return &conf, nil
}

// buildSlurmconfConfigMap creates slurmdbd.conf configmap with slurmdbd.conf
func buildSlurmdbdConfigMap(client kubernetes.Interface, wl *v1s.Slik) error {
	log := zap.L().Sugar()

	cm := client.CoreV1().ConfigMaps(wl.Namespace)

	conf, err := NewSlurmdbdConf(client, wl)
	if err != nil {
		return err
	}

	tpl, err := template.New("slurmdbd_conf").Funcs(
		template.FuncMap{"StringsJoin": strings.Join},
	).Parse(slurmdbdConfTpl)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, *conf); err != nil {
		return err
	}

	name := fmt.Sprintf("%s-slurmdbd", wl.Name)
	cmSpec := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wl.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "slik",
			},
		},
		Data: map[string]string{
			"slurmdbd.conf": buf.String(),
		},
	}

	log.Infof("configmap (slurmdbd.conf): %+v", cmSpec)

	_, err2 := cm.Create(context.TODO(), cmSpec, metav1.CreateOptions{})
	if err2 != nil {
		return err2
	}

	WaitForConfigMap(client, name, wl.Namespace)

	return nil
}
