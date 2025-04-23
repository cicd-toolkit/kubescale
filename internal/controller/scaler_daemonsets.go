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
	"encoding/json"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ScalerReconciler) handleDaemonSets(
	ctx context.Context,
	nsAnnotations map[string]string,
	ds *appsv1.DaemonSet,
	updateFunc func(int32) error,
) {
	log := ctrllog.FromContext(ctx)
	meta := &ds.ObjectMeta
	annotations := MergeAnnotations(nsAnnotations, meta.Annotations)
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
	if inDowntime {
		// Save current replica count
		nodeSelectorJSON, err := json.Marshal(ds.Spec.Template.Spec.NodeSelector)
		if err != nil {
			log.Error(err, "Failed to serialize NodeSelector", "namespace", meta.Namespace, "name", meta.Name)
			return
		}
		ds.ObjectMeta.Annotations[PreviousReplicasAnnotation] = string(nodeSelectorJSON)
		log.Info("Force NodeSelector", "namespace", meta.Namespace, "name", meta.Name)
		ds.Spec.Template.Spec.NodeSelector = map[string]string{
			"kubescale-suspend-daemonset": "true",
		}
		_ = r.Client.Update(ctx, ds)
		return
	}

	// restore if not in downtime and in uptime
	if !inDowntime && inUptime {
		var nodeSelector map[string]string
		if val, ok := annotations[PreviousReplicasAnnotation]; ok {
			if err := json.Unmarshal([]byte(val), &nodeSelector); err != nil {
				log.Error(err, "Failed to deserialize NodeSelector", "namespace", meta.Namespace, "name", meta.Name)
				return
			}
		}
		ds.Spec.Template.Spec.NodeSelector = nodeSelector
		delete(ds.ObjectMeta.Annotations, PreviousReplicasAnnotation)
		log.Info("Restoring NodeSelector", "namespace", meta.Namespace, "name", meta.Name)
		_ = r.Client.Update(ctx, ds)
		return
	}
}
