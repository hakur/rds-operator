package reconciler

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeleteSubResource delete CR instance sub resources
// Notice: argument [ list ] will store data, so don't make reference to them for gc
func DeleteSubResource(c client.Client, ctx context.Context, cr metav1.Object, list ...client.ObjectList) (err error) {
	for _, v := range list {
		labels := map[string]string{}

		err = c.List(ctx, v, client.InNamespace(cr.GetName()), client.MatchingLabels(labels))

		if err != nil && client.IgnoreNotFound(err) != nil {
			err = fmt.Errorf("delete sub resource failed,[namespace=%s] [finalizer=%s] [cr=%s] , err is -> %s",
				cr.GetNamespace(),
				cr.GetName(),
				strings.Join(cr.GetFinalizers(), ","),
				err.Error(),
			)
			return err
		} else {
			for {
				err = fmt.Errorf("delete sub resource failed,[namespace=%s] [finalizer=%s] [cr=%s] [subResourceName=%s], err is -> %s",
					cr.GetNamespace(),
					cr.GetName(),
					strings.Join(cr.GetFinalizers(), ","),
					"",
					err.Error(),
				)
				return err
			}
		}
	}

	return nil
}
