package controller

import (
	"context"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
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
	VclusterHelmChartVersion = "0.19.4"
	finalizer                = "openvirtualcluster.dev/finalizer"
)

// VClusterReconciler reconciles a VCluster object
type VClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters/finalizers,verbs=update

func (r *VClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var vclusterHelmRelease helmv2.HelmRelease

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

	// Create or update the HelmRepository
	helmRepo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vcluster",
			Namespace: vcluster.Namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: VclusterHelmChartRepo,
		},
	}
	if err := r.CreateOrUpdateHelmRepository(ctx, helmRepo, &vcluster); err != nil {
		return ctrl.Result{}, err
	}

	vclusterHelmRelease = helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      vcluster.Name,
			Namespace: vcluster.Namespace,
		},
		Spec: helmv2.HelmReleaseSpec{
			Chart: &helmv2.HelmChartTemplate{
				Spec: helmv2.HelmChartTemplateSpec{
					Chart:   VclusterHelmChart,
					Version: VclusterHelmChartVersion,
					SourceRef: helmv2.CrossNamespaceObjectReference{
						Kind:      sourcev1.HelmRepositoryKind,
						Name:      "vcluster",
						Namespace: vcluster.Namespace,
					},
				},
			},
			//ValuesFrom: []helmv2.ValuesReference{
			//	{
			//		Kind: helmv2.ValuesKindConfigMap,
			//		Name: vcluster.Spec.ValuesConfigMap,
			//	},
			//},
		},
		Status: helmv2.HelmReleaseStatus{},
	}

	err := r.Get(ctx, types.NamespacedName{Name: vcluster.Name, Namespace: vcluster.Namespace}, &vclusterHelmRelease)
	if err != nil && k8serrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(&vcluster, &vclusterHelmRelease, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &vclusterHelmRelease); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Created VCluster HelmRelease", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

		controllerutil.AddFinalizer(&vcluster, finalizer)
		if err := r.Update(ctx, &vcluster); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	} else if err != nil {
		log.Error(err, "Error getting VCluster HelmRelease")
		return ctrl.Result{}, err
	} else {
		log.Info("Updating HelmRelease for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		err = r.Update(ctx, &vclusterHelmRelease)
		if err != nil {
			log.Error(err, "Error updating HelmRelease for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		}
	}

	//vcluster.Status.HelmReleaseName = vclusterHelmRelease.Name
	//vcluster.Status.HelmReleaseNamespace = vclusterHelmRelease.Namespace
	//if err := updateVClusterStatus(ctx, r, vcluster); err != nil {
	//	log.Error(err, "Error updating HelmRelease info for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	//}

	if !vcluster.DeletionTimestamp.IsZero() {
		log.Info("VCluster is being deleted. Cleaning up resources", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

		err := r.Delete(ctx, &vclusterHelmRelease)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("VCluster HelmRelease has been deleted", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

		controllerutil.RemoveFinalizer(&vcluster, finalizer)
		if err := r.Update(ctx, &vcluster); err != nil {
			log.Error(err, "Error removing finalizer from VCluster object", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{}, err
		}
	}

	if !vcluster.Status.KubeconfigCreated {
		kubeconfig, err := checkVClusterSecret(ctx, r, vcluster.Name, vcluster.Namespace)
		if err != nil {
			log.Info("Requeue-ing after 30s as VCluster access secret has not been created yet", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		log.Info("VCluster access secret has been created. Updating CR status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		if kubeconfig != nil {
			vcluster.Status.KubeconfigSecretReference = *kubeconfig
			vcluster.Status.KubeconfigCreated = true

			log.Info("Updating VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			if err := updateVClusterStatus(ctx, r, vcluster); err != nil {
				log.Error(err, "Error updating VCluster Kubeconfig status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
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
		return nil, err
	}

	kubeconfig = corev1.SecretReference{
		Name:      vclusterSecret.Name,
		Namespace: vclusterSecret.Namespace,
	}

	return &kubeconfig, nil
}

//func checkHelmInstallJobStatus(ctx context.Context, r *VClusterReconciler, helmReleaseName, helmReleaseNamespace string) error {
//	var helmRelease helmv2.HelmRelease
//	var vclusterHelmInstallJob batchv1.Job
//
//	if err := r.Get(ctx, types.NamespacedName{Name: helmReleaseName, Namespace: helmReleaseNamespace}, &helmRelease); err != nil {
//		return err
//	}
//
//	vclusterInstallJobName := helmRelease.Status.
//	if err := r.Get(ctx, types.NamespacedName{Name: vclusterInstallJobName, Namespace: helmReleaseNamespace}, &vclusterHelmInstallJob); err != nil {
//		return err
//	}
//
//	if vclusterHelmInstallJob.Status.Active > 0 {
//		return fmt.Errorf("VCluster helm install job has not completed yet. Job name: %s", vclusterInstallJobName)
//	}
//
//	if vclusterHelmInstallJob.Status.Failed > 0 {
//		return fmt.Errorf("VCluster helm install job failed. Job name: %s", vclusterInstallJobName)
//	}
//
//	return nil
//}

func (r *VClusterReconciler) CreateOrUpdateHelmRepository(ctx context.Context, helmRepo *sourcev1.HelmRepository, owner *vclustersv1alpha1.VCluster) error {
	var existingHelmRepo sourcev1.HelmRepository
	err := r.Get(ctx, types.NamespacedName{Name: helmRepo.Name, Namespace: helmRepo.Namespace}, &existingHelmRepo)
	if err != nil && k8serrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(owner, helmRepo, r.Scheme); err != nil {
			return err
		}
		if err := r.Create(ctx, helmRepo); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		helmRepo.ResourceVersion = existingHelmRepo.ResourceVersion
		if err := r.Update(ctx, helmRepo); err != nil {
			return err
		}
	}
	return nil
}

func (r *VClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vclustersv1alpha1.VCluster{}).
		Owns(&helmv2.HelmRelease{}).
		Complete(r)
}
