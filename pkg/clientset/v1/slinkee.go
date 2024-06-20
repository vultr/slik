package v1

import (
	"context"

	v1 "github.com/vultr/slinkee/pkg/api/types/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type SlinkeeInterface interface {
	List(opts metav1.ListOptions) (*v1.SlinkeeList, error)
	Get(name string, options metav1.GetOptions) (*v1.Slinkee, error)
	Create(*v1.Slinkee) (*v1.Slinkee, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Update(slinkee *v1.Slinkee, options metav1.UpdateOptions) (*v1.Slinkee, error)
	UpdateStatus(slinkee *v1.Slinkee, options metav1.UpdateOptions) (*v1.Slinkee, error)
	Delete(name string, slinkee *v1.Slinkee, options metav1.DeleteOptions) error
	// ...
}

type slinkeeClient struct {
	restClient rest.Interface
	ctx        context.Context
}

func (c *slinkeeClient) List(opts metav1.ListOptions) (*v1.SlinkeeList, error) {
	result := v1.SlinkeeList{}

	err := c.restClient.
		Get().
		Resource("slinkees").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slinkeeClient) Get(name string, opts metav1.GetOptions) (*v1.Slinkee, error) {
	result := v1.Slinkee{}

	err := c.restClient.
		Get().
		Resource("slinkees").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slinkeeClient) Create(slinkee *v1.Slinkee) (*v1.Slinkee, error) {
	result := v1.Slinkee{}

	err := c.restClient.
		Post().
		Resource("slinkees").
		Body(slinkee).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slinkeeClient) Update(slinkee *v1.Slinkee, options metav1.UpdateOptions) (*v1.Slinkee, error) {
	result := v1.Slinkee{}

	err := c.restClient.Put().
		Namespace(slinkee.Spec.Namespace).
		Resource("slinkees").
		Name(slinkee.Name).
		Body(slinkee).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slinkeeClient) UpdateStatus(slinkee *v1.Slinkee, options metav1.UpdateOptions) (*v1.Slinkee, error) {
	result := v1.Slinkee{}

	err := c.restClient.Put().
		Namespace(slinkee.Spec.Namespace).
		Resource("slinkees").
		Name(slinkee.Name).
		SubResource("status").
		Body(slinkee).
		Do(c.ctx).
		Into(&result)

	return &result, err
}

func (c *slinkeeClient) Delete(name string, slinkee *v1.Slinkee, options metav1.DeleteOptions) error {
	err := c.restClient.Delete().
		Namespace(slinkee.Spec.Namespace).
		Resource("slinkees").
		Name(name).
		Body(&options).
		Do(c.ctx).
		Error()

	return err
}

func (c *slinkeeClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true

	return c.restClient.
		Get().
		Resource("slinkees").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch(c.ctx)
}
