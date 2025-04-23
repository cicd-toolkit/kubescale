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
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ScalerReconciler reconciles a Scaler object
type ScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	BaseAnnotation             = "kubescale"
	UptimeAnnotation           = BaseAnnotation + "/uptime"
	DowntimeAnnotation         = BaseAnnotation + "/downtime"
	PreviousReplicasAnnotation = BaseAnnotation + "/previous-replicas"
	CustomReplicaAnnotation    = BaseAnnotation + "/replicas"
	ExcludeAnnotation          = BaseAnnotation + "/exclude"
	ExcludeUntilAnnotation     = BaseAnnotation + "/exclude-until"
	UpDurationAnnotation       = BaseAnnotation + "/up"
	DownDurationAnnotation     = BaseAnnotation + "/down"
)

// +kubebuilder:rbac:groups=autoscale.kubescale.io,resources=scalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscale.kubescale.io,resources=scalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscale.kubescale.io,resources=scalers/finalizers,verbs=update

// Reconcile is part of the main Kubernetes reconciliation loop
func (r *ScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	go func() {
		time.Sleep(10 * time.Second) // Wait for the controller to be fully initialized
		for {
			if err := r.checkResources(); err != nil {
				fmt.Printf("Error checking resources: %v\n", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).   // Watches deployments
		Owns(&appsv1.StatefulSet{}). // Watches statefulsets
		Owns(&batchv1.CronJob{}).    // Watches cronjobs
		Complete(r)
}

func (r *ScalerReconciler) checkResources() error {
	ctx := context.Background()
	log := ctrllog.FromContext(ctx)
	now := time.Now().UTC()

	// Fetch namespace annotations
	nsMapAnnotations := make(map[string]map[string]string)
	var nsList corev1.NamespaceList
	if err := r.Client.List(ctx, &nsList); err == nil { // List all namespaces
		for _, ns := range nsList.Items {
			if nsAnn := ns.GetAnnotations(); nsAnn != nil {
				nsMapAnnotations[ns.Name] = nsAnn
			}
		}
	} else {
		log.Error(err, "Error listing namespaces")
	}

	// --- Deployments ---
	var deployList appsv1.DeploymentList
	if err := r.Client.List(ctx, &deployList, client.InNamespace("")); err == nil { // Fetch all namespaces
		for _, dep := range deployList.Items {
			r.transformAnnotations(ctx, &dep, now)
			nsAnnotations := nsMapAnnotations[dep.GetNamespace()]
			r.handleReplicatedResource(ctx, &dep.ObjectMeta, nsAnnotations, dep.Spec.Replicas, func(newReplicas int32) error {
				dep.Spec.Replicas = &newReplicas
				return r.Client.Update(ctx, &dep)
			})
		}
	} else {
		log.Error(err, "Error listing deployments")
	}

	// --- StatefulSets ---
	var stsList appsv1.StatefulSetList
	if err := r.Client.List(ctx, &stsList, client.InNamespace("")); err == nil { // Fetch all namespaces
		for _, sts := range stsList.Items {
			r.transformAnnotations(ctx, &sts, now)
			nsAnnotations := nsMapAnnotations[sts.GetNamespace()]
			r.handleReplicatedResource(ctx, &sts.ObjectMeta, nsAnnotations, sts.Spec.Replicas, func(newReplicas int32) error {
				sts.Spec.Replicas = &newReplicas
				return r.Client.Update(ctx, &sts)
			})
		}
	} else {
		log.Error(err, "Error listing statefulsets")
	}

	// --- DaemonSets ---
	var dsList appsv1.DaemonSetList
	if err := r.Client.List(ctx, &dsList, client.InNamespace("")); err == nil { // Fetch all namespaces
		for _, ds := range dsList.Items {
			r.transformAnnotations(ctx, &ds, now)
			nsAnnotations := nsMapAnnotations[ds.GetNamespace()]
			r.handleDaemonSets(ctx, nsAnnotations, &ds, func(newReplicas int32) error {
				// DaemonSets do not have replicas, so we don't need to update them
				return nil
			})
		}
	} else {
		log.Error(err, "Error listing daemonsets")
	}

	// --- CronJobs ---
	var cjList batchv1.CronJobList
	if err := r.Client.List(ctx, &cjList, client.InNamespace("")); err == nil { // Fetch all namespaces
		for _, cj := range cjList.Items {
			r.transformAnnotations(ctx, &cj, now)
			nsAnnotations := nsMapAnnotations[cj.GetNamespace()]
			r.handleCronJob(ctx, nsAnnotations, &cj)
		}
	} else {
		log.Error(err, "Error listing cronjobs")
	}

	return nil
}

func (r *ScalerReconciler) transformAnnotations(ctx context.Context, obj client.Object, now time.Time) {
	log := ctrllog.FromContext(ctx)
	ann := obj.GetAnnotations()
	if ann == nil {
		return
	}

	tz := "UTC"
	start := now
	foundAnnotations := false

	// Handle 'kubescale/up'
	if val, ok := ann[UpDurationAnnotation]; ok && val != "" {

		foundAnnotations = true
		duration, err := parseHumanDuration(val)
		if err != nil {
			log.Error(err, "Invalid duration format ", UpDurationAnnotation, obj.GetNamespace(), "/", obj.GetName())
			return
		}

		end := now.Add(duration)
		day := end.Weekday().String()[:3]
		ann[UptimeAnnotation] = fmt.Sprintf("%s-%s %s-%s %s", day, day, start.Format("15:04"), end.Format("15:04"), tz)
		delete(ann, UpDurationAnnotation)
	}

	// Handle 'kubescale/down'
	if val, ok := ann[DownDurationAnnotation]; ok && val != "" {
		foundAnnotations = true
		duration, err := parseHumanDuration(val)
		if err != nil {
			log.Error(err, "Invalid duration format ", DownDurationAnnotation, obj.GetNamespace(), "/", obj.GetName())
			return
		}
		start := now
		end := now.Add(duration)
		day := end.Weekday().String()[:3]
		ann[DowntimeAnnotation] = fmt.Sprintf("%s-%s %s-%s %s", day, day, start.Format("15:04"), end.Format("15:04"), tz)
		delete(ann, DownDurationAnnotation)
	}

	if foundAnnotations {
		obj.SetAnnotations(ann)
		// Patch the resource with updated annotations
		err := r.Client.Update(ctx, obj)
		if err != nil {
			log.Error(err, "Failed to patch annotations on ", DownDurationAnnotation, obj.GetNamespace(), "/", obj.GetName())
		} else {
			log.Info("Transformed annotations for ", obj.GetNamespace(), "/", obj.GetName())
		}
	}
}

func shouldSkipResource(meta *meta.ObjectMeta) bool {
	ann := meta.Annotations
	if ann == nil {
		return false
	}

	// Hard exclude
	if val, ok := ann[ExcludeAnnotation]; ok && strings.ToLower(val) == "true" {
		return true
	}

	// Time-based exclude
	if until, ok := ann[ExcludeUntilAnnotation]; ok {
		t, err := time.Parse(time.RFC3339, until)
		if err == nil && time.Now().Before(t) {
			return true
		}
	}

	return false
}
