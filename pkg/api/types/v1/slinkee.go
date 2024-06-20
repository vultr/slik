package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// to generate...
// 1. install: go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0
// 2a. generate: controller-gen object paths=./spec/api/types/v1/slinkee.go
// 2b. or: go generate ./...

//go:generate controller-gen object paths=$GOFILE

type SlinkeeSpec struct {
	Namespace  string `json:"namespace"`
	Slurmdbd   bool   `json:"slurmdbd"`
	Slurmrestd bool   `json:"slurmrestd"`

	MariaDB MariaDB `json:"mariadb"`
}

type MariaDB struct {
	StorageSize  string `json:"storage_size"`
	StorageClass string `json:"storage_class"`
}

type SlinkeeStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Slinkee struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlinkeeSpec   `json:"spec,omitempty"`
	Status SlinkeeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SlinkeeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Slinkee `json:"items"`
}

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Slinkee) DeepCopyInto(out *Slinkee) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = SlinkeeSpec{
		Namespace:  in.Spec.Namespace,
		Slurmdbd:   in.Spec.Slurmdbd,
		Slurmrestd: in.Spec.Slurmrestd,
		MariaDB: MariaDB{
			StorageSize:  in.Spec.MariaDB.StorageSize,
			StorageClass: in.Spec.MariaDB.StorageClass,
		},
	}
	out.Status = SlinkeeStatus{
		State: in.Status.State,
	}
}
