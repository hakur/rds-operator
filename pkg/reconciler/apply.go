// Package reconciler provide runtime client function wrap
package reconciler

import (
	"context"
	"strconv"
	"time"

	"github.com/hakur/rds-operator/pkg/types"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApplyService apply service
func ApplyService(c client.Client, ctx context.Context, data *corev1.Service, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData corev1.Service
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			//if service not exists, create it
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if service exists, update it now
		data.ResourceVersion = oldData.ResourceVersion
		data.Spec.ClusterIP = oldData.Spec.ClusterIP

		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}
	return nil
}

// ApplySecret apply secret
func ApplySecret(c client.Client, ctx context.Context, data *corev1.Secret, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData corev1.Secret
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if secret not exists, create it now
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if secret exists, update it now
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// ApplyStatefulSet  apply statefulset
func ApplyStatefulSet(c client.Client, ctx context.Context, data *appsv1.StatefulSet, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData appsv1.StatefulSet
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if deployment not exist, create it
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if deployment exists, update it
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// ApplyConfigMap apply statefulset
func ApplyConfigMap(c client.Client, ctx context.Context, data *corev1.ConfigMap, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData corev1.ConfigMap
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if configMap not exists, create it now
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if configMap exists, update it now
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
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if deployment exists, update it
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// ApplyCronJob apply conjob
func ApplyCronJob(c client.Client, ctx context.Context, data *batchv1.CronJob, parentObject metav1.Object, scheme *runtime.Scheme) (err error) {
	var oldData batchv1.CronJob
	if err := c.Get(ctx, client.ObjectKeyFromObject(data), &oldData); err != nil {
		if err := client.IgnoreNotFound(err); err == nil {
			// if cronjob not exist, create it
			if err := c.Create(ctx, data); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// if cronjob exists, update it
		if err := c.Update(ctx, data); err != nil {
			return err
		}
	}

	return nil
}

// AddPVCRetentionMark add deadline annottion to pvc
func AddPVCRetentionMark(c client.Client, ctx context.Context, namespace string, labelSelector map[string]string) (err error) {
	var pvcs corev1.PersistentVolumeClaimList
	if err = c.List(ctx, &pvcs, client.InNamespace(namespace), client.MatchingLabels(labelSelector)); err == nil {
		for _, pvc := range pvcs.Items {
			if _, ok := pvc.Annotations[types.PVCDeleteDateAnnotationName]; !ok {
				pvc.Annotations[types.PVCDeleteDateAnnotationName] = strconv.FormatInt(time.Now().Unix()+types.PVCDeleteRetentionSeconds, 10)
				if err = c.Update(ctx, &pvc); err != nil {
					return err
				}
			}
		}
	}

	return err
}

// RemovePVCRetentionMark delete deadline annottion from pvc
func RemovePVCRetentionMark(c client.Client, ctx context.Context, namespace string, labelSelector map[string]string) (err error) {
	var pvcs corev1.PersistentVolumeClaimList
	if err = c.List(ctx, &pvcs, client.InNamespace(namespace), client.MatchingLabels(labelSelector)); err == nil {
		for _, pvc := range pvcs.Items {
			if _, ok := pvc.Annotations[types.PVCDeleteDateAnnotationName]; ok {
				delete(pvc.Annotations, types.PVCDeleteDateAnnotationName)
				if err = c.Update(ctx, &pvc); err != nil {
					return err
				}
			}
		}
	}

	return err
}

// BuildCRPVCLabels generate CR subresource pvc labels
func BuildCRPVCLabels(metaObj metav1.Object, obj runtime.Object) map[string]string {
	return map[string]string{
		"cr-name":          metaObj.GetName(),
		"cr-group-version": obj.GetObjectKind().GroupVersionKind().Group + "___" + obj.GetObjectKind().GroupVersionKind().Version,
	}
}

// CheckOwnerRefExists check onwer reference exists
func CheckOwnerRefExists(owner metav1.Object, refs []metav1.OwnerReference) (refExists bool) {
	for _, ref := range refs {
		refEqual := true
		if ref.UID != owner.GetUID() {
			refEqual = false
		}

		if ref.Name != owner.GetName() {
			refEqual = false
		}

		if refEqual {
			refExists = true
			break
		}
	}

	return refExists
}
