package reconciler

import (
	"fmt"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type PVCCleaner struct {
}

// Create is called in response to an create event - e.g. Pod Creation.
func (t *PVCCleaner) Create(event.CreateEvent, workqueue.RateLimitingInterface) {}

// Update is called in response to an update event -  e.g. Pod Updated.
func (t *PVCCleaner) Update(evt event.UpdateEvent, wq workqueue.RateLimitingInterface) {
	fmt.Println("-------", evt.ObjectOld, evt.ObjectNew)
}

// Delete is called in response to a delete event - e.g. Pod Deleted.
func (t *PVCCleaner) Delete(event.DeleteEvent, workqueue.RateLimitingInterface) {}

// Generic is called in response to an event of an unknown type or a synthetic event triggered as a cron or
// external trigger request - e.g. reconcile Autoscaling, or a Webhook.
func (t *PVCCleaner) Generic(event.GenericEvent, workqueue.RateLimitingInterface) {}
