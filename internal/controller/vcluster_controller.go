/*
Copyright 2024.

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
	"time"

	helmv1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	vclustersv1alpha1 "openvirtualcluster.dev/openvirtualcluster/api/v1alpha1"
)

const (
	VclusterHelmChartRepo    = "https://charts.loft.sh"
	VclusterHelmChart        = "vcluster"
	VclusterHelmChartVersion = "0.18.1"
	finalizer                = "opencv.dev/finalizer"
)

// VClusterReconciler reconciles a VCluster object
type VClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *VClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var vclusterInstallationHelmChart helmv1.HelmChart

	// TODO: Break these into multiple reconcile functions and after each step completes successfully,
	// update status of the CR to ensure that the execution moves to the next reconciliation function
	log := log.FromContext(ctx)

	log.Info("Received reconcile request for VCluster", "Vcluster name", req.Name, "Vcluster namespace", req.Namespace)

	var vcluster vclustersv1alpha1.VCluster
	if err := r.Get(ctx, req.NamespacedName, &vcluster); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("VCluster deleted", "NamespacedName", req.NamespacedName)
		} else {
			log.Error(err, "Failed to get VCluster", "NamespacedName", req.NamespacedName)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	vclusterInstallationHelmChart = helmv1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vcluster.Name,
			Namespace: vcluster.Namespace,
		},
		Spec: helmv1.HelmChartSpec{
			Repo:        VclusterHelmChartRepo,
			HelmVersion: VclusterHelmChartVersion,
			Chart:       VclusterHelmChart,
			Set:         vcluster.Spec.Set,
		},
		Status: helmv1.HelmChartStatus{},
	}

	err := r.Get(ctx,
		types.NamespacedName{
			Name:      vcluster.Name,
			Namespace: vcluster.Namespace,
		},
		&vclusterInstallationHelmChart,
	)
	if err != nil && k8serrors.IsNotFound(err) {
		// Create the chart

		if err := controllerutil.SetControllerReference(&vcluster, &vclusterInstallationHelmChart, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		// Create the helm chart CR
		err := r.Create(context.Background(), &vclusterInstallationHelmChart)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Created VCluster helm chart", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

		controllerutil.AddFinalizer(&vcluster, finalizer)
		if err := r.Update(context.Background(), &vcluster); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	} else if err != nil {
		log.Error(err, "error in getting VCluster helm chart")
		return ctrl.Result{}, err
	} else {
		// Update helm chart
		log.Info("Updating helm chart for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		err = r.Update(ctx, &vclusterInstallationHelmChart)
		if err != nil {
			log.Error(err, "error updating helm chart for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		}
	}
	// Update vcluster helm chart info
	vcluster.Status.HelmChartName = vclusterInstallationHelmChart.Name
	vcluster.Status.HelmChartNamespace = vclusterInstallationHelmChart.Namespace
	if err := updateVClusterStatus(ctx, r, vcluster); err != nil {
		log.Error(err, "error updating helm chart info for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	}

	// Handle delete
	if !vcluster.DeletionTimestamp.IsZero() {
		log.Info("VCluster is being deleted. Cleaning up resources", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

		err := r.Delete(context.Background(), &vclusterInstallationHelmChart)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("VCluster helm chart has been deleted", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

		controllerutil.RemoveFinalizer(&vcluster, finalizer)
		if err := r.Update(context.Background(), &vcluster); err != nil {
			log.Error(err, "error removing finalizer from VCluster object", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{}, err
		}
	}

	if !vcluster.Status.JobCompleted {
		err = checkHelmInstallJobStatus(ctx, r, vclusterInstallationHelmChart.Name, vclusterInstallationHelmChart.Namespace)
		if err != nil {
			log.Info("Requeue-ing after 10s as VCluster helm install job has not completed yet", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
		}
		vcluster.Status.JobCompleted = true
		err = updateVClusterStatus(ctx, r, vcluster)
		if err != nil {
			log.Error(err, "error updating vcluster job completion status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{}, err
		}
	}
	log.Info("Helm install job for VCluster has completed", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

	if !vcluster.Status.KubeconfigCreated {
		kubeconfig, err := checkVClusterSecret(ctx, r, vcluster.Name, vcluster.Namespace)
		if err != nil {
			log.Info("Requeue-ing after 30s as VCluster access secret has not been created yet", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		log.Info("VCluster access secret has been created. Updating CR status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		// Update kubeconfig in VCluster CR
		if kubeconfig != nil {
			vcluster.Status.KubeconfigSecretReference = *kubeconfig
			vcluster.Status.KubeconfigCreated = true

			// Update VCluster
			log.Info("Updating VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			if err := updateVClusterStatus(ctx, r, vcluster); err != nil {
				log.Error(err, "error updating VCluster Kubeconfig status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
				return ctrl.Result{}, err
			}
		}
	}
	log.Info("Kubeconfig for VCluster has been created", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

	log.Info("Done reconciling VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	return ctrl.Result{}, nil
}

func updateVClusterStatus(ctx context.Context, r *VClusterReconciler, vcluster vclustersv1alpha1.VCluster) error {
	if err := r.Status().Update(ctx, &vcluster); err != nil {
		return err
	}
	return nil
}

func checkVClusterSecret(ctx context.Context, r *VClusterReconciler, vclusterName string, vclusterNamespace string) (*corev1.SecretReference, error) {
	var vclusterSecret corev1.Secret
	var kubeconfig corev1.SecretReference

	vclusterSecretName := fmt.Sprintf("vc-%s", vclusterName)

	if err := r.Get(ctx, types.NamespacedName{Name: vclusterSecretName, Namespace: vclusterNamespace}, &vclusterSecret); err != nil {
		// Check after 30s if secret doesn't exist
		return nil, err
	}

	// Got VCluster. Now update the secretRef
	kubeconfig = corev1.SecretReference{
		Name:      vclusterSecret.Name,
		Namespace: vclusterSecret.Namespace,
	}

	return &kubeconfig, nil
}

func checkHelmInstallJobStatus(ctx context.Context, r *VClusterReconciler, helmChartName, helmChartNamespace string) error {
	var helmChart helmv1.HelmChart
	var vclusterHelmInstallJob batchv1.Job

	// Get latest Helm chart for status check
	if err := r.Get(ctx, types.NamespacedName{Name: helmChartName, Namespace: helmChartNamespace}, &helmChart); err != nil {
		// The Chart object should exist by now
		return err
	}

	vclusterInstallJobName := helmChart.Status.JobName
	if err := r.Get(ctx, types.NamespacedName{Name: vclusterInstallJobName, Namespace: helmChartNamespace}, &vclusterHelmInstallJob); err != nil {
		return err
	}

	if vclusterHelmInstallJob.Status.Active > 0 {
		return fmt.Errorf("VCluster helm install job has not completed yet. Job name: %s", vclusterInstallJobName)
	}

	if vclusterHelmInstallJob.Status.Failed > 0 {
		return fmt.Errorf("VCluster helm install job failed. Job name: %s", vclusterInstallJobName)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vclustersv1alpha1.VCluster{}).
		Owns(&helmv1.HelmChart{}).
		Complete(r)
}
