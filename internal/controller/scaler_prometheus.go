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

// apiVersion: monitoring.coreos.com/v1
// kind: Prometheus

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

var PrometheusGVR = schema.GroupVersionResource{
	Group:    "monitoring.coreos.com",
	Version:  "v1",
	Resource: "prometheuses",
}

func (r *ScalerReconciler) handlePrometheus(ctx context.Context, nsAnnotations map[string]string, p *unstructured.Unstructured) error {
	log := ctrllog.FromContext(ctx)
	annotations := MergeAnnotations(nsAnnotations, p.GetAnnotations())
	if annotations == nil {
		return nil
	}

	objectMeta := &metav1.ObjectMeta{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(p.Object, objectMeta); err != nil {
		return fmt.Errorf("failed to convert metadata to ObjectMeta: %v", err)
	}

	if shouldSkipResource(objectMeta) {
		log.Info("Skipping Prometheus", "namespace", p.GetNamespace(), "name", p.GetName())
		return nil
	}

	var inUptime, inDowntime bool

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
	// log.Info("Prometheus Annotations", "namespace", p.GetNamespace(), "name", p.GetName(), "annotations", annotations)
	// fmt.Println("inUptime:", inUptime)
	// fmt.Println("inDowntime:", inDowntime)
	currentReplicas, _, err := unstructured.NestedInt64(p.Object, "spec", "replicas")
	if err != nil {
		// log.Error("failed to get replicas")
		return fmt.Errorf("failed to get replicas: %v", err)
	}
	dynamicClient, err := dynamic.NewForConfig(config.GetConfigOrDie())
	if err != nil {
		return fmt.Errorf("failed to create dynamic client: %w", err)
	}

	// Suspend if in downtime
	if inDowntime && currentReplicas > 0 {
		log.Info("Suspending Prometheus", "namespace", p.GetNamespace(), "name", p.GetName())

		err = unstructured.SetNestedField(p.Object, int64(0), "spec", "replicas")
		if err != nil {
			return fmt.Errorf("failed to set replicas: %v", err)
		}
		// Add annotation to store the current replicas
		annotations[PreviousReplicasAnnotation] = strconv.Itoa(int(currentReplicas))
		p.SetAnnotations(annotations)
		// / Update the Prometheus object
		_, err = dynamicClient.Resource(PrometheusGVR).Namespace(p.GetNamespace()).Update(ctx, p, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update Prometheus: %v", err)
		}
	}

	// Resume if in uptime and not in downtime
	if !inDowntime && inUptime && currentReplicas == 0 {
		log.Info("Resuming Prometheus", "namespace", p.GetNamespace(), "name", p.GetName())
		restore := int64(1)
		if val, ok := annotations[PreviousReplicasAnnotation]; ok {
			if prev, err := strconv.Atoi(val); err == nil && prev > 0 {
				restore = int64(prev)
			}
		}
		delete(annotations, PreviousReplicasAnnotation)
		p.SetAnnotations(annotations)
		err = unstructured.SetNestedField(p.Object, restore, "spec", "replicas")
		if err != nil {
			return fmt.Errorf("failed to set replicas: %v", err)
		}
		// / Update the Prometheus object
		_, err = dynamicClient.Resource(PrometheusGVR).Namespace(p.GetNamespace()).Update(ctx, p, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update Prometheus: %v", err)
		}
	}

	return nil
}
