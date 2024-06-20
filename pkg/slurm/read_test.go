package slurm

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var jobsFixture = []batchv1.Job{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job1",
			Namespace: "default",
		},
	},
}

var cronJobsFixture = []batchv1.CronJob{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cronjob1",
			Namespace: "default",
		},
	},
}

var deploymentFixture = []appsv1.Deployment{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deployment1",
			Namespace: "default",
		},
	},
}

var podsFixture = []v1.Pod{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "job1",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "job1",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodFailed,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod4",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "job1",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod5",
			Namespace: "default",
			Labels: map[string]string{
				"job-name": "job1",
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	},
}

type Fixture1 struct {
	name         string
	workloadType string

	result      bool
	description string
}

func TestSlurmExists(t *testing.T) {
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

	fixtures := []Fixture1{
		{
			name: "job1",

			result:      true,
			description: "happy path",
		},
		{
			name: "does_not_exist",

			result:      false,
			description: "bad path",
		},
		{
			name: "cronjob1",

			result:      true,
			description: "happy path",
		},
		{
			name: "does_not_exist",

			result:      false,
			description: "bad path",
		},
		{
			name: "deployment1",

			result:      true,
			description: "happy path",
		},
		{
			name: "does_not_exist",

			result:      false,
			description: "bad path",
		},
		{
			name: "pod1",

			result:      true,
			description: "happy path",
		},
		{
			name: "does_not_exist",

			result:      false,
			description: "bad path",
		},
		{
			name:         "does_not_exist",
			workloadType: "asdf",

			result:      true,
			description: "bad path",
		},
	}

	for _, fixture := range fixtures {
		result := SlurmExists(client, fixture.name, fixture.workloadType)

		if result != fixture.result {
			t.Errorf("\n%s\nexpect: %t\nactual: %t", fixture.description, fixture.result, result)
		}
	}
}

type Fixture2 struct {
	name string

	podFixture []v1.Pod

	result      string
	description string
}

func TestGetPodStatus(t *testing.T) {
	fixtures := []Fixture2{
		{
			name: "pod1",

			podFixture: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
						Labels: map[string]string{
							"job-name": "job1",
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						Conditions: []v1.PodCondition{
							{
								Message: "test",
							},
						},
					},
				},
			},

			result:      "test",
			description: "PodRunning path",
		},
		{
			name: "pod1",

			podFixture: []v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod1",
						Namespace: "default",
						Labels: map[string]string{
							"job-name": "job1",
						},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},

			result:      "Unknown",
			description: "Unknown path",
		},
	}

	for _, fixture := range fixtures {
		client := fake.NewSimpleClientset(
			&v1.PodList{
				Items: fixture.podFixture,
			},
		)

		result := GetPodStatus(client, fixture.name, "default")

		if result != fixture.result {
			t.Errorf("\n%s\nexpect: %s\nactual: %s", fixture.description, fixture.result, result)
		}
	}
}

type Fixture3 struct {
	name string

	nodeFixture []v1.Node

	result      int
	description string
}

func TestGetAllNodes(t *testing.T) {
	fixtures := []Fixture3{
		{
			name: "node1",

			nodeFixture: []v1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node1",
					},
				},
			},

			result:      1,
			description: "PodRunning path",
		},
	}

	for _, fixture := range fixtures {
		client := fake.NewSimpleClientset(
			&v1.NodeList{
				Items: fixture.nodeFixture,
			},
		)

		nl, _ := GetAllNodes(client)

		if len(nl.Items) != fixture.result {
			t.Errorf("\n%s\nexpect: %d\nactual: %d", fixture.description, fixture.result, len(nl.Items))
		}
	}
}

type Fixture4 struct {
	name string

	deploymentFixture []appsv1.Deployment

	result      string
	description string
}

func TestGetDeploymentStatus(t *testing.T) {
	fixtures := []Fixture4{
		{
			name: "deployment1",

			deploymentFixture: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "deployment1",
						Namespace: "default",
					},
					Status: appsv1.DeploymentStatus{
						Conditions: []appsv1.DeploymentCondition{
							{
								Message: "test",
							},
						},
					},
				},
			},

			result:      "test",
			description: "Message path",
		},
		{
			name: "deployment1",

			deploymentFixture: []appsv1.Deployment{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "deployment1",
					},
				},
			},

			result:      "Unknown",
			description: "unknown path",
		},
		{
			name: "deployment1",

			deploymentFixture: nil,

			result:      "Unknown",
			description: "nil path",
		},
	}

	for _, fixture := range fixtures {
		client := fake.NewSimpleClientset(
			&appsv1.DeploymentList{
				Items: fixture.deploymentFixture,
			},
		)

		status := GetDeploymentStatus(client, fixture.name, "default")

		if status != fixture.result {
			t.Errorf("\n%s\nexpect: %s\nactual: %s", fixture.description, fixture.result, status)
		}
	}
}
