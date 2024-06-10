package slurm

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type Fixture6 struct {
	name         string
	workloadType string

	result      error
	description string
}

func TestWorkloadDelete(t *testing.T) {
	client := fake.NewSimpleClientset(
		&batchv1.JobList{
			Items: jobsFixture,
		},
		&batchv1.CronJobList{
			Items: cronJobsFixture,
		},
		&appsv1.DeploymentList{
			Items: deploymentFixture,
		},
		&v1.PodList{
			Items: podsFixture,
		},
	)

	fixtures := []Fixture6{
		{
			name: "job1",

			result:      nil,
			description: "happy path for job 1",
		},
		{
			name: "dne",

			result:      nil,
			description: "happy path for job 2",
		},
		{
			name: "cronjob1",

			result:      nil,
			description: "happy path for cron job 1",
		},
		{
			name: "dne",

			result:      nil,
			description: "happy path for cron job 2",
		},
		{
			name: "deployment1",

			result:      nil,
			description: "happy path for deployment 1",
		},
		{
			name: "dne",

			result:      nil,
			description: "happy path for deployment 2",
		},
		{
			name: "pod1",

			result:      nil,
			description: "happy path for pod 1",
		},
		{
			name: "dne",

			result:      nil,
			description: "happy path for pod 2",
		},
		{
			name:         "dne",
			workloadType: "idklol",

			result:      nil,
			description: "bad path",
		},
	}

	for _, fixture := range fixtures {
		result := SlurmDelete(client, fixture.name, "default")

		if result != fixture.result {
			t.Errorf("\n%s\nexpect: %s\nactual: %s", fixture.description, fixture.result, result)
		}
	}
}
