package v1

import (
	"context"

	v1 "github.com/AhmedTremo/slik/pkg/api/types/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type SlikInterface interface {
	List(opts metav1.ListOptions) (*v1.SlikList, error)
	Get(name string, options metav1.GetOptions) (*v1.Slik, error)
	Create(*v1.Slik) (*v1.Slik, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Update(slik *v1.Slik, options metav1.UpdateOptions) (*v1.Slik, error)
	UpdateStatus(slik *v1.Slik, options metav1.UpdateOptions) (*v1.Slik, error)
	Delete(name string, slik *v1.Slik, options metav1.DeleteOptions) error
	// ...
}

type slikClient struct {
	restClient rest.Interface
	ctx        context.Context
}

func (c *slikClient) List(opts metav1.ListOptions) (*v1.SlikList, error) {
	result := v1.SlikList{}

	err := c.restClient.
		Get().
		Resource("sliks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slikClient) Get(name string, opts metav1.GetOptions) (*v1.Slik, error) {
	result := v1.Slik{}

	err := c.restClient.
		Get().
		Resource("sliks").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slikClient) Create(slik *v1.Slik) (*v1.Slik, error) {
	result := v1.Slik{}

	err := c.restClient.
		Post().
		Resource("sliks").
		Body(slik).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slikClient) Update(slik *v1.Slik, options metav1.UpdateOptions) (*v1.Slik, error) {
	result := v1.Slik{}

	err := c.restClient.Put().
		Namespace(slik.Spec.Namespace).
		Resource("sliks").
		Name(slik.Name).
		Body(slik).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slikClient) UpdateStatus(slik *v1.Slik, options metav1.UpdateOptions) (*v1.Slik, error) {
	result := v1.Slik{}

	err := c.restClient.Put().
		Namespace(slik.Spec.Namespace).
		Resource("sliks").
		Name(slik.Name).
		SubResource("status").
		Body(slik).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slikClient) Delete(name string, slik *v1.Slik, options metav1.DeleteOptions) error {
	err := c.restClient.Delete().
		Namespace(slik.Spec.Namespace).
		Resource("sliks").
		Name(name).
		Body(&options).
		Do(c.ctx).
		Error()

	return err
}

func (c *slikClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true

	return c.restClient.
		Get().
		Resource("sliks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(c.ctx)
}
