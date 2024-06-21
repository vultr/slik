package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// to generate...
// 1. install: go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.15.0
// 2a. generate: controller-gen object paths=./spec/api/types/v1/slik.go
// 2b. or: go generate ./...

//go:generate controller-gen object paths=$GOFILE

type SlikSpec struct {
	Namespace  string `json:"namespace"`
	Slurmdbd   bool   `json:"slurmdbd"`
	Slurmrestd bool   `json:"slurmrestd"`

	MariaDB MariaDB `json:"mariadb"`
}

type MariaDB struct {
	StorageSize  string `json:"storage_size"`
	StorageClass string `json:"storage_class"`
}

type SlikStatus struct {
	State string `json:"state"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Slik struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlikSpec   `json:"spec,omitempty"`
	Status SlikStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type SlikList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Slik `json:"items"`
}

// DeepCopyInto copies all properties of this object into another object of the
// same type that is provided as a pointer.
func (in *Slik) DeepCopyInto(out *Slik) {
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = SlikSpec{
		Namespace:  in.Spec.Namespace,
		Slurmdbd:   in.Spec.Slurmdbd,
		Slurmrestd: in.Spec.Slurmrestd,
		MariaDB: MariaDB{
			StorageSize:  in.Spec.MariaDB.StorageSize,
			StorageClass: in.Spec.MariaDB.StorageClass,
		},
	}
	out.Status = SlikStatus{
		State: in.Status.State,
	}
}
