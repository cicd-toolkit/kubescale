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

	batchv1 "k8s.io/api/batch/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *ScalerReconciler) handleCronJob(ctx context.Context, cj *batchv1.CronJob) {
	log := ctrllog.FromContext(ctx)
	annotations := cj.Annotations
	if annotations == nil {
		return
	}
	if shouldSkipResource(&cj.ObjectMeta) {
		log.Info("Skipping CronJob", "namespace", cj.Namespace, "name", cj.Name)
		return
	}

	var inUptime, inDowntime bool

	if val, ok := annotations[DowntimeAnnotation]; ok {
		sd, ed, st, et, loc, err := parseScalerAnnotation(val)
		if err == nil {
			inDowntime = isNowInUptime(sd, ed, st, et, loc)
		}
	}

	if !inDowntime {
		if val, ok := annotations[UptimeAnnotation]; ok {
			sd, ed, st, et, loc, err := parseScalerAnnotation(val)
			if err == nil {
				inUptime = isNowInUptime(sd, ed, st, et, loc)
			}
		}
	}

	// Suspend if in downtime
	if inDowntime && (cj.Spec.Suspend == nil || !*cj.Spec.Suspend) {
		log.Info("Suspending CronJob", "namespace", cj.Namespace, "name", cj.Name)
		s := true
		cj.Spec.Suspend = &s
		_ = r.Client.Update(ctx, cj)
		return
	}

	// Resume if in uptime and not in downtime
	if !inDowntime && inUptime && (cj.Spec.Suspend != nil && *cj.Spec.Suspend) {
		log.Info("Resuming CronJob", "namespace", cj.Namespace, "name", cj.Name)
		s := false
		cj.Spec.Suspend = &s
		_ = r.Client.Update(ctx, cj)
	}
}
