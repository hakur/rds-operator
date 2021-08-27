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

// FakeMysqlBackups implements MysqlBackupInterface
type FakeMysqlBackups struct {
	Fake *FakeApisV1alpha1
	ns   string
}

var mysqlbackupsResource = schema.GroupVersionResource{Group: "apis", Version: "v1alpha1", Resource: "mysqlbackups"}

var mysqlbackupsKind = schema.GroupVersionKind{Group: "apis", Version: "v1alpha1", Kind: "MysqlBackup"}

// Get takes name of the mysqlBackup, and returns the corresponding mysqlBackup object, and an error if there is any.
func (c *FakeMysqlBackups) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.MysqlBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(mysqlbackupsResource, c.ns, name), &v1alpha1.MysqlBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MysqlBackup), err
}

// List takes label and field selectors, and returns the list of MysqlBackups that match those selectors.
func (c *FakeMysqlBackups) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MysqlBackupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(mysqlbackupsResource, mysqlbackupsKind, c.ns, opts), &v1alpha1.MysqlBackupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MysqlBackupList{ListMeta: obj.(*v1alpha1.MysqlBackupList).ListMeta}
	for _, item := range obj.(*v1alpha1.MysqlBackupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested mysqlBackups.
func (c *FakeMysqlBackups) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(mysqlbackupsResource, c.ns, opts))

}

// Create takes the representation of a mysqlBackup and creates it.  Returns the server's representation of the mysqlBackup, and an error, if there is any.
func (c *FakeMysqlBackups) Create(ctx context.Context, mysqlBackup *v1alpha1.MysqlBackup, opts v1.CreateOptions) (result *v1alpha1.MysqlBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(mysqlbackupsResource, c.ns, mysqlBackup), &v1alpha1.MysqlBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MysqlBackup), err
}

// Update takes the representation of a mysqlBackup and updates it. Returns the server's representation of the mysqlBackup, and an error, if there is any.
func (c *FakeMysqlBackups) Update(ctx context.Context, mysqlBackup *v1alpha1.MysqlBackup, opts v1.UpdateOptions) (result *v1alpha1.MysqlBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(mysqlbackupsResource, c.ns, mysqlBackup), &v1alpha1.MysqlBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MysqlBackup), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMysqlBackups) UpdateStatus(ctx context.Context, mysqlBackup *v1alpha1.MysqlBackup, opts v1.UpdateOptions) (*v1alpha1.MysqlBackup, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(mysqlbackupsResource, "status", c.ns, mysqlBackup), &v1alpha1.MysqlBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MysqlBackup), err
}

// Delete takes name of the mysqlBackup and deletes it. Returns an error if one occurs.
func (c *FakeMysqlBackups) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(mysqlbackupsResource, c.ns, name), &v1alpha1.MysqlBackup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMysqlBackups) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(mysqlbackupsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MysqlBackupList{})
	return err
}

// Patch applies the patch and returns the patched mysqlBackup.
func (c *FakeMysqlBackups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.MysqlBackup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(mysqlbackupsResource, c.ns, name, pt, data, subresources...), &v1alpha1.MysqlBackup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.MysqlBackup), err
}