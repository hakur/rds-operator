/*
MIT License

Copyright (c) 2021 Software Authors

Software Authors are:
    Xing Yu, email: yuxing951@gmail.com,yuxing951@hotmail.com
    Yi Zhou, email: 6098550@qq.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMysqls implements MysqlInterface
type FakeMysqls struct {
	Fake *FakeApisV1alpha1
	ns   string
}

var mysqlsResource = schema.GroupVersionResource{Group: "apis", Version: "v1alpha1", Resource: "mysqls"}

var mysqlsKind = schema.GroupVersionKind{Group: "apis", Version: "v1alpha1", Kind: "Mysql"}

// Get takes name of the mysql, and returns the corresponding mysql object, and an error if there is any.
func (c *FakeMysqls) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Mysql, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mysqlsResource, c.ns, name), &v1alpha1.Mysql{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Mysql), err
}

// List takes label and field selectors, and returns the list of Mysqls that match those selectors.
func (c *FakeMysqls) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MysqlList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mysqlsResource, mysqlsKind, c.ns, opts), &v1alpha1.MysqlList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MysqlList{ListMeta: obj.(*v1alpha1.MysqlList).ListMeta}
	for _, item := range obj.(*v1alpha1.MysqlList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mysqls.
func (c *FakeMysqls) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mysqlsResource, c.ns, opts))

}

// Create takes the representation of a mysql and creates it.  Returns the server's representation of the mysql, and an error, if there is any.
func (c *FakeMysqls) Create(ctx context.Context, mysql *v1alpha1.Mysql, opts v1.CreateOptions) (result *v1alpha1.Mysql, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mysqlsResource, c.ns, mysql), &v1alpha1.Mysql{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Mysql), err
}

// Update takes the representation of a mysql and updates it. Returns the server's representation of the mysql, and an error, if there is any.
func (c *FakeMysqls) Update(ctx context.Context, mysql *v1alpha1.Mysql, opts v1.UpdateOptions) (result *v1alpha1.Mysql, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mysqlsResource, c.ns, mysql), &v1alpha1.Mysql{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Mysql), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMysqls) UpdateStatus(ctx context.Context, mysql *v1alpha1.Mysql, opts v1.UpdateOptions) (*v1alpha1.Mysql, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mysqlsResource, "status", c.ns, mysql), &v1alpha1.Mysql{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Mysql), err
}

// Delete takes name of the mysql and deletes it. Returns an error if one occurs.
func (c *FakeMysqls) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(mysqlsResource, c.ns, name, opts), &v1alpha1.Mysql{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMysqls) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mysqlsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MysqlList{})
	return err
}

// Patch applies the patch and returns the patched mysql.
func (c *FakeMysqls) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Mysql, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mysqlsResource, c.ns, name, pt, data, subresources...), &v1alpha1.Mysql{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Mysql), err
}
