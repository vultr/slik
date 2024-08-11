package slurm

import (
	"os"
	"testing"

	v1s "github.com/AhmedTremo/slik/pkg/api/types/v1"

	"k8s.io/client-go/kubernetes/fake"
)

type Fixture5 struct {
	name string

	wl *v1s.Slik

	result      error
	description string
}

func TestCreateSlurm(t *testing.T) {
	os.Args = append(os.Args, "-config=../../cmd/slik/config.yaml")

	client := fake.NewSimpleClientset()

	fixtures := []Fixture5{
		{
			name: "wl1",

			wl: &v1s.Slik{},

			result:      nil,
			description: "happy path for deployment",
		},
		{
			name: "wl2",

			wl: &v1s.Slik{},

			result:      nil,
			description: "happy path for job",
		},
		{
			name: "wl3",

			wl: &v1s.Slik{},

			result:      nil,
			description: "happy path for cron job",
		},
		{
			name: "wl4",

			wl: &v1s.Slik{},

			result:      nil,
			description: "happy path for pod",
		},
		{
			name: "wl1",

			wl: &v1s.Slik{},

			result:      nil,
			description: "happy path for job with gpu",
		},
		{
			name: "wl1",

			wl: &v1s.Slik{},

			result:      nil,
			description: "happy path for command and args",
		},
		{
			name: "wl1",

			wl: &v1s.Slik{},

			result:      nil,
			description: "quantity failure 1",
		},
	}

	for _, fixture := range fixtures {
		result := CreateSlurm(client, fixture.wl)

		if result != fixture.result {
			t.Errorf("\n%s\nexpect: %s\nactual: %s", fixture.description, fixture.result, result)
		}
	}
}
