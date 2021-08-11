// Package reconciler provide runtime client function wrap
package reconciler

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
			return fmt.Errorf("SetControllerReference error: %s", err.Error())
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
		if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
			return fmt.Errorf("SetControllerReference error: %s", err.Error())
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
		if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
			return fmt.Errorf("SetControllerReference error: %s", err.Error())
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
		if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
			return fmt.Errorf("SetControllerReference error: %s", err.Error())
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
		if err := ctrl.SetControllerReference(parentObject, data, scheme); err != nil {
			return fmt.Errorf("SetControllerReference error: %s", err.Error())
		}
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// AddPVCRetentionMark add delete deadline annottion to pvc
func AddPVCRetentionMark() {

}

// RemovePVCRetentionMark remove delete deadline annottion to pvc
func RemovePVCRetentionMark() {

}
