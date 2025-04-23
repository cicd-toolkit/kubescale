/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ScalerReconciler) handleReplicatedResource(
	ctx context.Context,
	meta *meta.ObjectMeta,
	replicas *int32,
	updateFunc func(int32) error,
) {
	log := ctrllog.FromContext(ctx)
	annotations := meta.Annotations
	if annotations == nil {
		log.Info("No annotations found for resource", "namespace", meta.Namespace, "name", meta.Name)
		return
	}
	if shouldSkipResource(meta) {
		log.Info("Skipping resource", "namespace", meta.Namespace, "name", meta.Name)
		return
	}

	var inUptime, inDowntime bool

	// Handle downtime first (it takes priority)
	if val, ok := annotations[DowntimeAnnotation]; ok {
		timerange, err := parseScalerAnnotation(val)
		if err == nil {
			inDowntime = timerange.isInRange(time.Time{})
		}
	}

	if !inDowntime {
		if val, ok := annotations[UptimeAnnotation]; ok {
			timerange, err := parseScalerAnnotation(val)
			if err == nil {
				inUptime = timerange.isInRange(time.Time{})
			}
		}
	}

	// scale to 0 if in downtime
	if inDowntime && *replicas != 0 {
		// Save current replica count
		meta.Annotations[PreviousReplicasAnnotation] = fmt.Sprintf("%d", *replicas)
		log.Info("Scaling down resource", "namespace", meta.Namespace, "name", meta.Name)
		_ = updateFunc(0)
		_ = r.Client.Update(ctx, &appsv1.Deployment{
			ObjectMeta: *meta,
		})
		return
	}

	// restore if not in downtime and in uptime
	if !inDowntime && inUptime && *replicas == 0 {
		restore := int32(1)
		if val, ok := annotations[PreviousReplicasAnnotation]; ok {
			if prev, err := strconv.Atoi(val); err == nil && prev > 0 {
				restore = int32(prev)
			}
		}
		delete(meta.Annotations, PreviousReplicasAnnotation)
		log.Info("Restoring resource", "namespace", meta.Namespace, "name", meta.Name)
		_ = updateFunc(restore)
	}
}
