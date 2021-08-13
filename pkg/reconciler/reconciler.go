// Package reconciler provide runtime client function wrap
package reconciler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hakur/rds-operator/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyService apply service and set parent gc
func ApplyService(c client.Client, ctx context.Context, data *corev1.Service, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData corev1.Service
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// set gc reference
			if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
				return fmt.Errorf("SetControllerReference error: %s", err.Error())
			}
			//if service not exists, create it
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if service exists, update it
		// set gc reference
		if exists, err := CheckOwnerRefExists(parentObject, data); err != nil {
			if !exists {
				if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
					return fmt.Errorf("SetControllerReference error: %s", err.Error())
				}
			}
		} else {
			return err
		}
		data.ResourceVersion = oldData.ResourceVersion
		data.Spec.ClusterIP = oldData.Spec.ClusterIP
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

// ApplySecret apply secret and set parent gc
func ApplySecret(c client.Client, ctx context.Context, data *corev1.Secret, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData corev1.Secret
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if secret not exists, create it now
			// set gc reference
			if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
				return fmt.Errorf("SetControllerReference error: %s", err.Error())
			}
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if secret exists, update it now
		// set gc reference
		if exists, err := CheckOwnerRefExists(parentObject, data); err != nil {
			if !exists {
				if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
					return fmt.Errorf("SetControllerReference error: %s", err.Error())
				}
			}
		} else {
			return err
		}
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// ApplyStatefulSet  apply statefulset and set parent gc
func ApplyStatefulSet(c client.Client, ctx context.Context, data *appsv1.StatefulSet, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData appsv1.StatefulSet
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if deployment not exist, create it
			// set gc reference
			if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
				return fmt.Errorf("SetControllerReference error: %s", err.Error())
			}
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if deployment exists, update it
		// set gc reference
		if exists, err := CheckOwnerRefExists(parentObject, data); err != nil {
			if !exists {
				if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
					return fmt.Errorf("SetControllerReference error: %s", err.Error())
				}
			}
		} else {
			return err
		}
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// ApplyConfigMap apply statefulset and set parent gc
func ApplyConfigMap(c client.Client, ctx context.Context, data *corev1.ConfigMap, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData corev1.ConfigMap
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if configMap not exists, create it now
			// set gc reference
			if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
				return fmt.Errorf("SetControllerReference error: %s", err.Error())
			}
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if configMap exists, update it now
		// set gc reference
		if exists, err := CheckOwnerRefExists(parentObject, data); err != nil {
			if !exists {
				if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
					return fmt.Errorf("SetControllerReference error: %s", err.Error())
				}
			}
		} else {
			return err
		}
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

func ApplyDeployment(c client.Client, ctx context.Context, data *appsv1.Deployment, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData appsv1.Deployment
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if deployment not exist, create it
			// set gc reference
			if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
				return fmt.Errorf("SetControllerReference error: %s", err.Error())
			}
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if deployment exists, update it
		// set gc reference
		if exists, err := CheckOwnerRefExists(parentObject, data); err != nil {
			if !exists {
				if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
					return fmt.Errorf("SetControllerReference error: %s", err.Error())
				}
			}
		} else {
			return err
		}

		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// AddPVCRetentionMark add deadline annottion to pvc
func AddPVCRetentionMark(c client.Client, ctx context.Context, namespace string, labelSet map[string]string) (err error) {
	var pvcs corev1.PersistentVolumeClaimList
	if err = c.List(ctx, &pvcs, client.InNamespace(namespace), client.MatchingLabels(labelSet)); err != nil {
		for _, pvc := range pvcs.Items {
			if _, ok := pvc.Annotations[types.PVCDeleteDateAnnotationName]; !ok {
				pvc.Annotations[types.PVCDeleteDateAnnotationName] = strconv.FormatInt(time.Now().Unix()+types.PVCDeleteRetentionSeconds, 10)
				if err = c.Update(ctx, &pvc); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// RemovePVCRetentionMark delete deadline annottion from pvc
func RemovePVCRetentionMark(c client.Client, ctx context.Context, namespace string, labelSet map[string]string) (err error) {
	var pvcs corev1.PersistentVolumeClaimList
	if err = c.List(ctx, &pvcs, client.InNamespace(namespace), client.MatchingLabels(labelSet)); err != nil {
		for _, pvc := range pvcs.Items {
			if _, ok := pvc.Annotations[types.PVCDeleteDateAnnotationName]; !ok {
				delete(pvc.Annotations, types.PVCDeleteDateAnnotationName)
				if err = c.Update(ctx, &pvc); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// BuildCRPVCLabels generate CR subresource pvc labels
func BuildCRPVCLabels(crName, crGroupVersion string) map[string]string {
	return map[string]string{
		"cr-name":          crName,
		"cr-group-version": crGroupVersion,
	}
}

// CheckOwnerRefExists check onwer reference exists
func CheckOwnerRefExists(owner metav1.Object, controlled metav1.Object) (refExists bool, err error) {
	var obkind schema.ObjectKind
	if runtimeObj, ok := owner.(runtime.Object); !ok {
		return false, fmt.Errorf("owner [%s:%s/%s] cannot assert to a runtime.Object", owner.GetResourceVersion(), owner.GetNamespace(), owner.GetName())
	} else {
		obkind = runtimeObj.GetObjectKind()
	}

	refExists = true

	for _, ref := range controlled.GetOwnerReferences() {

		if ref.UID != owner.GetUID() {
			refExists = false
		}

		if ref.APIVersion != obkind.GroupVersionKind().Version {
			refExists = false
		}

		if ref.Kind != obkind.GroupVersionKind().Kind {
			refExists = false
		}

		if ref.Name != owner.GetName() {

			refExists = false
		}

		if refExists {
			break
		}
	}

	return refExists, nil
}
