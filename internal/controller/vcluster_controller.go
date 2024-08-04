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
	"encoding/json"
	"fmt"
	"github.com/loft-sh/vcluster-config/config"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"os"
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

	Vclustersv1alpha1 "github.com/OpenVirtualCluster/virtual-cluster-operator/api/v1alpha1"
)

// VclusterReconciler reconciles a Vcluster object
type VclusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	VclusterHelmChartRepo    = "https://charts.loft.sh"
	VclusterHelmChart        = "Vcluster"
	VclusterHelmChartVersion = "0.19.4"
	finalizer                = "openvirtualcluster.dev/finalizer"
	labelManagedBy           = "Vcluster.loft.sh/managed-by"
	labelHelmReleaseName     = "helm.toolkit.fluxcd.io/name"
)

// +kubebuilder:rbac:groups=Vclusters.openvirtualcluster.dev,resources=Vclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=Vclusters.openvirtualcluster.dev,resources=Vclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=Vclusters.openvirtualcluster.dev,resources=Vclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Vcluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *VclusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var VclusterHelmRelease helmv2.HelmRelease

	log := log.FromContext(ctx)
	log.Info("Received reconcile request for Vcluster", "Vcluster name", req.Name, "Vcluster namespace", req.Namespace)

	var Vcluster Vclustersv1alpha1.Vcluster
	if err := r.Get(ctx, req.NamespacedName, &Vcluster); err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("Vcluster deleted", "NamespacedName", req.NamespacedName)
		} else {
			log.Error(err, "Failed to get Vcluster", "NamespacedName", req.NamespacedName)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Ensure conditions field is initialized
	if Vcluster.Status.Conditions == nil {
		Vcluster.Status.Conditions = []metav1.Condition{}
	}

	helmRepo := &sourcev1.HelmRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "Vcluster",
			Namespace: Vcluster.Namespace,
		},
		Spec: sourcev1.HelmRepositorySpec{
			URL: VclusterHelmChartRepo,
		},
	}
	if err := r.CreateOrUpdateHelmRepository(ctx, helmRepo, &Vcluster); err != nil {
		return ctrl.Result{}, err
	}

	if Vcluster.Spec.Config != nil {
		configJson, err := getJSONFromConfig(Vcluster.Spec.Config)
		if err != nil {
			return ctrl.Result{}, err
		}

		VclusterHelmRelease = helmv2.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Vcluster.Name,
				Namespace: Vcluster.Namespace,
			},
			Spec: helmv2.HelmReleaseSpec{
				Chart: &helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Chart:   VclusterHelmChart,
						Version: VclusterHelmChartVersion,
						SourceRef: helmv2.CrossNamespaceObjectReference{
							Kind:      sourcev1.HelmRepositoryKind,
							Name:      "Vcluster",
							Namespace: Vcluster.Namespace,
						},
					},
				},
				Values: configJson,
			},
			Status: helmv2.HelmReleaseStatus{},
		}

	} else {
		vclusterValuesFile, err := r.getVclusterValuesFile(ctx, Vcluster)
		if err != nil {
			return ctrl.Result{}, err
		}

		// defer the deletion of the values file once the reconcile is completed
		defer func(fileLoc string) {
			err := deleteFile(fileLoc)
			if err != nil {
				log.Error(err, "Error deleting Vcluster values file", "File location", fileLoc)
			}
		}(vclusterValuesFile)

		VclusterHelmRelease = helmv2.HelmRelease{
			ObjectMeta: metav1.ObjectMeta{
				Name:      Vcluster.Name,
				Namespace: Vcluster.Namespace,
			},
			Spec: helmv2.HelmReleaseSpec{
				Chart: &helmv2.HelmChartTemplate{
					Spec: helmv2.HelmChartTemplateSpec{
						Chart:   VclusterHelmChart,
						Version: VclusterHelmChartVersion,
						SourceRef: helmv2.CrossNamespaceObjectReference{
							Kind:      sourcev1.HelmRepositoryKind,
							Name:      "Vcluster",
							Namespace: Vcluster.Namespace,
						},
						ValuesFile: vclusterValuesFile,
					},
				},
			},
			Status: helmv2.HelmReleaseStatus{},
		}
	}

	err := r.Get(ctx, types.NamespacedName{Name: Vcluster.Name, Namespace: Vcluster.Namespace}, &VclusterHelmRelease)
	if err != nil && k8serrors.IsNotFound(err) {
		if err := controllerutil.SetControllerReference(&Vcluster, &VclusterHelmRelease, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, &VclusterHelmRelease); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Created Vcluster HelmRelease", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)

		controllerutil.AddFinalizer(&Vcluster, finalizer)
		if err := r.Update(ctx, &Vcluster); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Added finalizer to Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	} else if err != nil {
		log.Error(err, "Error getting Vcluster HelmRelease")
		return ctrl.Result{}, err
	} else {
		log.Info("Updating HelmRelease for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
		err = r.Update(ctx, &VclusterHelmRelease)
		if err != nil {
			log.Error(err, "Error updating HelmRelease for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
		}
	}

	if !Vcluster.DeletionTimestamp.IsZero() {
		log.Info("Vcluster is being deleted. Cleaning up resources", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)

		err := r.Delete(ctx, &VclusterHelmRelease)
		if err != nil {
			return ctrl.Result{}, err
		}
		log.Info("Vcluster HelmRelease has been deleted", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)

		controllerutil.RemoveFinalizer(&Vcluster, finalizer)
		if err := r.Update(ctx, &Vcluster); err != nil {
			log.Error(err, "Error removing finalizer from Vcluster object", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
			return ctrl.Result{}, err
		}
	}

	// Check if the Vcluster should sleep
	if Vcluster.Spec.Sleep {
		log.Info("Vcluster is set to sleep. Scaling down related statefulsets and deleting related pods", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
		if err := r.ScaleDownVclusterResources(ctx, Vcluster); err != nil {
			log.Error(err, "Error scaling down resources for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Vcluster is awake. Scaling up related statefulsets", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
		if err := r.ScaleUpVclusterResources(ctx, Vcluster); err != nil {
			log.Error(err, "Error scaling up resources for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
			return ctrl.Result{}, err
		}
	}

	if !Vcluster.Status.KubeconfigCreated {
		kubeconfig, err := checkVclusterSecret(ctx, r, Vcluster.Name, Vcluster.Namespace)
		if err != nil {
			log.Info("Requeue-ing after 30s as Vcluster access secret has not been created yet", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
		log.Info("Vcluster access secret has been created. Updating CR status", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
		if kubeconfig != nil {
			Vcluster.Status.KubeconfigSecretReference = kubeconfig
			Vcluster.Status.KubeconfigCreated = true

			log.Info("Updating Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
			if err := updateVclusterStatus(ctx, r, Vcluster); err != nil {
				log.Error(err, "Error updating Vcluster Kubeconfig status", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
				return ctrl.Result{}, err
			}
		}
	}
	log.Info("Kubeconfig for Vcluster has been created", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)

	// Update Vcluster status to mirror HelmRelease status
	Vcluster.Status.Conditions = []metav1.Condition{}
	for _, condition := range VclusterHelmRelease.Status.Conditions {
		Vcluster.Status.Conditions = append(Vcluster.Status.Conditions, metav1.Condition{
			Type:               "Helm" + condition.Type,
			Status:             condition.Status,
			LastTransitionTime: condition.LastTransitionTime,
			Reason:             condition.Reason,
			Message:            condition.Message,
		})
	}

	log.Info("Updating Vcluster status to mirror HelmRelease status", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	if err := updateVclusterStatus(ctx, r, Vcluster); err != nil {
		log.Error(err, "Error updating Vcluster status", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
		return ctrl.Result{}, err
	}

	log.Info("Done reconciling Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	return ctrl.Result{}, nil
}

func getJSONFromConfig(config *config.Config) (*apiextensionsv1.JSON, error) {
	configJson, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return &apiextensionsv1.JSON{Raw: configJson}, nil
}

func deleteFile(fileLoc string) error {
	if err := os.Remove(fileLoc); err != nil {
		return err
	}
	return nil
}

func (r *VclusterReconciler) getVclusterValuesFile(ctx context.Context, Vcluster Vclustersv1alpha1.Vcluster) (string, error) {
	vclusterValuesFile := fmt.Sprintf("/tmp/%s-values.yaml", Vcluster.Name)

	file, err := os.OpenFile(vclusterValuesFile, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if Vcluster.Spec.ValuesString != "" {
		_, err = file.WriteString(Vcluster.Spec.ValuesString)
		if err != nil {
			return "", err
		}
	}

	if Vcluster.Spec.ValuesFromConfigMap != nil {
		var configMap corev1.ConfigMap
		if err = r.Get(ctx, types.NamespacedName{Name: Vcluster.Spec.ValuesFromConfigMap.Name, Namespace: Vcluster.Namespace}, &configMap); err != nil {
			return "", err
		}

		_, err = file.WriteString(configMap.Data[Vcluster.Spec.ValuesFromConfigMap.Key])
		if err != nil {
			return "", err
		}
	}

	return vclusterValuesFile, nil
}

func (r *VclusterReconciler) ScaleDownVclusterResources(ctx context.Context, Vcluster Vclustersv1alpha1.Vcluster) error {
	log := log.FromContext(ctx)

	// Scale down the StatefulSet with the labelHelmReleaseName label
	var statefulsets appsv1.StatefulSetList
	labels := client.MatchingLabels{labelHelmReleaseName: Vcluster.Name}
	if err := r.List(ctx, &statefulsets, client.InNamespace(Vcluster.Namespace), labels); err != nil {
		return err
	}

	if len(statefulsets.Items) == 0 {
		log.Info("No statefulsets found for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	} else {
		for _, statefulset := range statefulsets.Items {
			statefulset.Spec.Replicas = new(int32) // Set replicas to 0
			if err := r.Update(ctx, &statefulset); err != nil {
				return err
			}
		}
		log.Info("Scaled down statefulsets for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	}

	// Delete Pods with the labelManagedBy label
	var pods corev1.PodList
	labels = client.MatchingLabels{labelManagedBy: Vcluster.Name}
	if err := r.List(ctx, &pods, client.InNamespace(Vcluster.Namespace), labels); err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		log.Info("No pods found for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	} else {
		for _, pod := range pods.Items {
			if err := r.Delete(ctx, &pod); err != nil {
				return err
			}
		}
		log.Info("Deleted pods for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	}

	return nil
}

func (r *VclusterReconciler) ScaleUpVclusterResources(ctx context.Context, Vcluster Vclustersv1alpha1.Vcluster) error {
	log := log.FromContext(ctx)

	// Scale up the StatefulSet with the labelHelmReleaseName label
	var statefulsets appsv1.StatefulSetList
	labels := client.MatchingLabels{labelHelmReleaseName: Vcluster.Name}
	if err := r.List(ctx, &statefulsets, client.InNamespace(Vcluster.Namespace), labels); err != nil {
		return err
	}

	if len(statefulsets.Items) == 0 {
		log.Info("No statefulsets found for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	} else {
		for _, statefulset := range statefulsets.Items {
			replicas := int32(1) // Set desired number of replicas
			statefulset.Spec.Replicas = &replicas
			if err := r.Update(ctx, &statefulset); err != nil {
				return err
			}
		}
		log.Info("Scaled up statefulsets for Vcluster", "Vcluster name", Vcluster.Name, "Vcluster namespace", Vcluster.Namespace)
	}

	return nil
}

func updateVclusterStatus(ctx context.Context, r *VclusterReconciler, Vcluster Vclustersv1alpha1.Vcluster) error {
	if err := r.Status().Update(ctx, &Vcluster); err != nil {
		return err
	}
	return nil
}

func checkVclusterSecret(ctx context.Context, r *VclusterReconciler, VclusterName string, VclusterNamespace string) (*corev1.SecretReference, error) {
	var VclusterSecret corev1.Secret
	var kubeconfig corev1.SecretReference

	VclusterSecretName := fmt.Sprintf("vc-%s", VclusterName)

	if err := r.Get(ctx, types.NamespacedName{Name: VclusterSecretName, Namespace: VclusterNamespace}, &VclusterSecret); err != nil {
		return nil, err
	}

	kubeconfig = corev1.SecretReference{
		Name:      VclusterSecret.Name,
		Namespace: VclusterSecret.Namespace,
	}

	return &kubeconfig, nil
}

func (r *VclusterReconciler) CreateOrUpdateHelmRepository(ctx context.Context, helmRepo *sourcev1.HelmRepository, owner *Vclustersv1alpha1.Vcluster) error {
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

// SetupWithManager sets up the controller with the Manager.
func (r *VclusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&Vclustersv1alpha1.Vcluster{}).
		Complete(r)
}
