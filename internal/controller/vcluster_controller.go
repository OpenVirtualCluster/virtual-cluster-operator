package controller

import (
	"context"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	appsv1 "k8s.io/api/apps/v1"
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
	labelManagedBy           = "vcluster.loft.sh/managed-by"
	labelHelmReleaseName     = "helm.toolkit.fluxcd.io/name"
)

// VClusterReconciler reconciles a VCluster object
type VClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vclusters.openvirtualcluster.dev,resources=vclusters/finalizers,verbs=update

// add the helm controller rbac
//+kubebuilder:rbac:groups=helm.toolkit.fluxcd.io,resources=helmreleases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=helm.toolkit.fluxcd.io,resources=helmreleases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=helm.toolkit.fluxcd.io,resources=helmreleases/finalizers,verbs=update

// add the source controller rbac
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=helmrepositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=helmrepositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=source.toolkit.fluxcd.io,resources=helmrepositories/finalizers,verbs=update

// add statefulset rbac
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

// add deployments rbac
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// add the ingress rbac
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// add networkpolicy rbac
//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

// add pods rbac
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete

// add services rbac
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// add secret rbac
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

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

	// Ensure conditions field is initialized
	if vcluster.Status.Conditions == nil {
		vcluster.Status.Conditions = []metav1.Condition{}
	}

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

	// Check if the VCluster should sleep
	if vcluster.Spec.Sleep {
		log.Info("VCluster is set to sleep. Scaling down related statefulsets and deleting related pods", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		if err := r.ScaleDownVClusterResources(ctx, vcluster); err != nil {
			log.Error(err, "Error scaling down resources for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			return ctrl.Result{}, err
		}
	} else {
		log.Info("VCluster is awake. Scaling up related statefulsets", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		if err := r.ScaleUpVClusterResources(ctx, vcluster); err != nil {
			log.Error(err, "Error scaling up resources for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
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
			vcluster.Status.KubeconfigSecretReference = kubeconfig
			vcluster.Status.KubeconfigCreated = true

			log.Info("Updating VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
			if err := updateVClusterStatus(ctx, r, vcluster); err != nil {
				log.Error(err, "Error updating VCluster Kubeconfig status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
				return ctrl.Result{}, err
			}
		}
	}
	log.Info("Kubeconfig for VCluster has been created", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)

	// Update VCluster status to mirror HelmRelease status
	vcluster.Status.Conditions = []metav1.Condition{}
	for _, condition := range vclusterHelmRelease.Status.Conditions {
		vcluster.Status.Conditions = append(vcluster.Status.Conditions, metav1.Condition{
			Type:               "Helm" + condition.Type,
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	log.Info("Updating VCluster status to mirror HelmRelease status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	if err := updateVClusterStatus(ctx, r, vcluster); err != nil {
		log.Error(err, "Error updating VCluster status", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
		return ctrl.Result{}, err
	}

	log.Info("Done reconciling VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	return ctrl.Result{}, nil
}

func (r *VClusterReconciler) ScaleDownVClusterResources(ctx context.Context, vcluster vclustersv1alpha1.VCluster) error {
	log := log.FromContext(ctx)

	// Scale down the StatefulSet with the labelHelmReleaseName label
	var statefulsets appsv1.StatefulSetList
	labels := client.MatchingLabels{labelHelmReleaseName: vcluster.Name}
	if err := r.List(ctx, &statefulsets, client.InNamespace(vcluster.Namespace), labels); err != nil {
		return err
	}

	if len(statefulsets.Items) == 0 {
		log.Info("No statefulsets found for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	} else {
		for _, statefulset := range statefulsets.Items {
			statefulset.Spec.Replicas = new(int32) // Set replicas to 0
			if err := r.Update(ctx, &statefulset); err != nil {
				return err
			}
		}
		log.Info("Scaled down statefulsets for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	}

	// Delete Pods with the labelManagedBy label
	var pods corev1.PodList
	labels = client.MatchingLabels{labelManagedBy: vcluster.Name}
	if err := r.List(ctx, &pods, client.InNamespace(vcluster.Namespace), labels); err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		log.Info("No pods found for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	} else {
		for _, pod := range pods.Items {
			if err := r.Delete(ctx, &pod); err != nil {
				return err
			}
		}
		log.Info("Deleted pods for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	}

	return nil
}

func (r *VClusterReconciler) ScaleUpVClusterResources(ctx context.Context, vcluster vclustersv1alpha1.VCluster) error {
	log := log.FromContext(ctx)

	// Scale up the StatefulSet with the labelHelmReleaseName label
	var statefulsets appsv1.StatefulSetList
	labels := client.MatchingLabels{labelHelmReleaseName: vcluster.Name}
	if err := r.List(ctx, &statefulsets, client.InNamespace(vcluster.Namespace), labels); err != nil {
		return err
	}

	if len(statefulsets.Items) == 0 {
		log.Info("No statefulsets found for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	} else {
		for _, statefulset := range statefulsets.Items {
			replicas := int32(1) // Set desired number of replicas
			statefulset.Spec.Replicas = &replicas
			if err := r.Update(ctx, &statefulset); err != nil {
				return err
			}
		}
		log.Info("Scaled up statefulsets for VCluster", "VCluster name", vcluster.Name, "VCluster namespace", vcluster.Namespace)
	}

	return nil
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
