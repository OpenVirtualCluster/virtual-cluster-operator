package v1alpha1

import (
	networkingv1 "k8s.io/api/networking/v1"
)

type Config struct {
	Sync             Sync             `json:"sync"`
	ControlPlane     ControlPlane     `json:"controlPlane"`
	ExportKubeConfig ExportKubeConfig `json:"exportKubeConfig"`
	Plugins          Plugins          `json:"plugins"`
	Experimental     Experimental     `json:"experimental"`
	Telemetry        Telemetry        `json:"telemetry"`
}

type Services struct {
	Enabled bool `json:"enabled"`
}
type Endpoints struct {
	Enabled bool `json:"enabled"`
}
type PersistentVolumeClaims struct {
	Enabled bool `json:"enabled"`
}
type ConfigMaps struct {
	Enabled bool `json:"enabled"`
	All     bool `json:"all"`
}
type Secrets struct {
	Enabled bool `json:"enabled"`
	All     bool `json:"all"`
}
type TranslateImage struct {
}
type Limits struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}
type Requests struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}
type Resources struct {
	Limits   Limits   `json:"limits"`
	Requests Requests `json:"requests"`
}
type InitContainer struct {
	Image     string    `json:"image"`
	Resources Resources `json:"resources"`
}
type RewriteHosts struct {
	Enabled       bool          `json:"enabled"`
	InitContainer InitContainer `json:"initContainer"`
}
type Pods struct {
	Enabled               bool           `json:"enabled"`
	TranslateImage        TranslateImage `json:"translateImage"`
	UseSecretsForSATokens bool           `json:"useSecretsForSATokens"`
	RewriteHosts          RewriteHosts   `json:"rewriteHosts"`
}
type Ingresses struct {
	Enabled bool `json:"enabled"`
}
type PriorityClasses struct {
	Enabled bool `json:"enabled"`
}
type NetworkPolicies struct {
	Enabled bool `json:"enabled"`
}
type VolumeSnapshots struct {
	Enabled bool `json:"enabled"`
}
type PodDisruptionBudgets struct {
	Enabled bool `json:"enabled"`
}
type ServiceAccounts struct {
	Enabled bool `json:"enabled"`
}
type StorageClasses struct {
	Enabled bool `json:"enabled"`
}
type PersistentVolumes struct {
	Enabled bool `json:"enabled"`
}
type ToHost struct {
	Services               Services               `json:"services"`
	Endpoints              Endpoints              `json:"endpoints"`
	PersistentVolumeClaims PersistentVolumeClaims `json:"persistentVolumeClaims"`
	ConfigMaps             ConfigMaps             `json:"configMaps"`
	Secrets                Secrets                `json:"secrets"`
	Pods                   Pods                   `json:"pods"`
	Ingresses              Ingresses              `json:"ingresses"`
	PriorityClasses        PriorityClasses        `json:"priorityClasses"`
	NetworkPolicies        NetworkPolicies        `json:"networkPolicies"`
	VolumeSnapshots        VolumeSnapshots        `json:"volumeSnapshots"`
	PodDisruptionBudgets   PodDisruptionBudgets   `json:"podDisruptionBudgets"`
	ServiceAccounts        ServiceAccounts        `json:"serviceAccounts"`
	StorageClasses         StorageClasses         `json:"storageClasses"`
	PersistentVolumes      PersistentVolumes      `json:"persistentVolumes"`
}
type Events struct {
	Enabled bool `json:"enabled"`
}
type CsiDrivers struct {
	Enabled string `json:"enabled"`
}
type CsiNodes struct {
	Enabled string `json:"enabled"`
}
type CsiStorageCapacities struct {
	Enabled string `json:"enabled"`
}
type IngressClasses struct {
	Enabled bool `json:"enabled"`
}
type Labels struct {
}
type Selector struct {
	All    bool   `json:"all"`
	Labels Labels `json:"labels"`
}
type Nodes struct {
	Enabled          bool     `json:"enabled"`
	SyncBackChanges  bool     `json:"syncBackChanges"`
	ClearImageStatus bool     `json:"clearImageStatus"`
	Selector         Selector `json:"selector"`
}
type FromHost struct {
	Events               Events               `json:"events"`
	CsiDrivers           CsiDrivers           `json:"csiDrivers"`
	CsiNodes             CsiNodes             `json:"csiNodes"`
	CsiStorageCapacities CsiStorageCapacities `json:"csiStorageCapacities"`
	StorageClasses       StorageClasses       `json:"storageClasses"`
	IngressClasses       IngressClasses       `json:"ingressClasses"`
	Nodes                Nodes                `json:"nodes"`
}
type Sync struct {
	ToHost   ToHost   `json:"toHost"`
	FromHost FromHost `json:"fromHost"`
}
type Image struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}
type APIServer struct {
	Enabled         bool     `json:"enabled"`
	Command         []string `json:"command"`
	ExtraArgs       []string `json:"extraArgs"`
	ImagePullPolicy string   `json:"imagePullPolicy"`
	Image           Image    `json:"image"`
}
type ControllerManager struct {
	Enabled         bool     `json:"enabled"`
	Command         []string `json:"command"`
	ExtraArgs       []string `json:"extraArgs"`
	ImagePullPolicy string   `json:"imagePullPolicy"`
	Image           Image    `json:"image"`
}
type Scheduler struct {
	Command         []string `json:"command"`
	ExtraArgs       []string `json:"extraArgs"`
	ImagePullPolicy string   `json:"imagePullPolicy"`
	Image           Image    `json:"image"`
}
type SecurityContext struct {
}
type K8S struct {
	Enabled           bool              `json:"enabled"`
	APIServer         APIServer         `json:"apiServer"`
	ControllerManager ControllerManager `json:"controllerManager"`
	Scheduler         Scheduler         `json:"scheduler"`
	SecurityContext   SecurityContext   `json:"securityContext"`
	Resources         Resources         `json:"resources"`
}
type K3S struct {
	Enabled         bool            `json:"enabled"`
	Command         []string        `json:"command"`
	ExtraArgs       []string        `json:"extraArgs"`
	ImagePullPolicy string          `json:"imagePullPolicy"`
	Image           Image           `json:"image"`
	SecurityContext SecurityContext `json:"securityContext"`
	Resources       Resources       `json:"resources"`
}
type K0S struct {
	Enabled         bool            `json:"enabled"`
	Config          string          `json:"config"`
	Command         []string        `json:"command"`
	ExtraArgs       []string        `json:"extraArgs"`
	ImagePullPolicy string          `json:"imagePullPolicy"`
	Image           Image           `json:"image"`
	SecurityContext SecurityContext `json:"securityContext"`
	Resources       Resources       `json:"resources"`
}
type Distro struct {
	K8S K8S `json:"k8s"`
	K3S K3S `json:"k3s"`
	K0S K0S `json:"k0s"`
}
type Embedded struct {
	Enabled bool `json:"enabled"`
}

type Database struct {
	Embedded Embedded `json:"embedded"`
	External External `json:"external"`
}
type ETCDEmbedded struct {
	Enabled                 bool `json:"enabled"`
	MigrateFromDeployedEtcd bool `json:"migrateFromDeployedEtcd"`
}

type CoreDNSDeploymentPods struct {
	Annotations Annotations `json:"annotations"`
	Labels      Labels      `json:"labels"`
}
type HighAvailability struct {
	Replicas int `json:"replicas"`
}
type NodeSelector struct {
}
type Affinity struct {
}
type Scheduling struct {
	PodManagementPolicy       string       `json:"podManagementPolicy"`
	NodeSelector              NodeSelector `json:"nodeSelector"`
	Affinity                  Affinity     `json:"affinity"`
	Tolerations               []string     `json:"tolerations"`
	TopologySpreadConstraints []string     `json:"topologySpreadConstraints"`
	PriorityClassName         string       `json:"priorityClassName"`
}
type PodSecurityContext struct {
}
type ContainerSecurityContext struct {
}
type Security struct {
	PodSecurityContext       PodSecurityContext       `json:"podSecurityContext"`
	ContainerSecurityContext ContainerSecurityContext `json:"containerSecurityContext"`
}
type VolumeClaim struct {
	Enabled         bool     `json:"enabled"`
	RetentionPolicy string   `json:"retentionPolicy"`
	Size            string   `json:"size"`
	StorageClass    string   `json:"storageClass"`
	AccessModes     []string `json:"accessModes"`
}
type Persistence struct {
	VolumeClaim          VolumeClaim `json:"volumeClaim"`
	VolumeClaimTemplates []string    `json:"volumeClaimTemplates"`
	AddVolumes           []string    `json:"addVolumes"`
	AddVolumeMounts      []string    `json:"addVolumeMounts"`
}
type StatefulSet struct {
	Enabled            bool             `json:"enabled"`
	EnableServiceLinks bool             `json:"enableServiceLinks"`
	Annotations        Annotations      `json:"annotations"`
	Labels             Labels           `json:"labels"`
	Image              Image            `json:"image"`
	ImagePullPolicy    string           `json:"imagePullPolicy"`
	ExtraArgs          []string         `json:"extraArgs"`
	Env                []string         `json:"env"`
	Resources          Resources        `json:"resources"`
	Pods               Pods             `json:"pods"`
	HighAvailability   HighAvailability `json:"highAvailability"`
	Scheduling         Scheduling       `json:"scheduling"`
	Security           Security         `json:"security"`
	Persistence        Persistence      `json:"persistence"`
}

type Etcd struct {
	Embedded ETCDEmbedded `json:"embedded"`
	Deploy   Deploy       `json:"deploy"`
}

type BackingStore struct {
	Database Database `json:"database"`
	Etcd     Etcd     `json:"etcd"`
}

type Deploy struct {
	Enabled         bool            `json:"enabled"`
	StatefulSet     ETCDStatefulSet `json:"statefulSet"`
	Service         ETCDService     `json:"service"`
	HeadlessService HeadlessService `json:"headlessService"`
}

type ETCDService struct {
	Enabled     bool        `json:"enabled"`
	Annotations Annotations `json:"annotations"`
}

type ETCDStatefulSet struct {
	Enabled            bool             `json:"enabled"`
	EnableServiceLinks bool             `json:"enableServiceLinks"`
	Annotations        Annotations      `json:"annotations"`
	Labels             Labels           `json:"labels"`
	Image              Image            `json:"image"`
	ImagePullPolicy    string           `json:"imagePullPolicy"`
	ExtraArgs          []string         `json:"extraArgs"`
	Env                []string         `json:"env"`
	Resources          Resources        `json:"resources"`
	Pods               Pods             `json:"pods"`
	HighAvailability   HighAvailability `json:"highAvailability"`
	Scheduling         Scheduling       `json:"scheduling"`
	Security           Security         `json:"security"`
	Persistence        Persistence      `json:"persistence"`
}

type Proxy struct {
	BindAddress string   `json:"bindAddress"`
	Port        int      `json:"port"`
	ExtraSANs   []string `json:"extraSANs"`
}
type CoreDNSServiceSpec struct {
	Type string `json:"type"`
}

type VclusterServiceSpec struct {
	Type string `json:"type"`
}

type CoreDNSService struct {
	Annotations []Annotations      `json:"annotations"`
	Labels      Labels             `json:"labels"`
	Spec        CoreDNSServiceSpec `json:"spec"`
}

type MatchLabels struct {
	K8SApp string `json:"k8s-app"`
}
type LabelSelector struct {
	MatchLabels MatchLabels `json:"matchLabels"`
}
type TopologySpreadConstraints struct {
	MaxSkew           int           `json:"maxSkew"`
	TopologyKey       string        `json:"topologyKey"`
	WhenUnsatisfiable string        `json:"whenUnsatisfiable"`
	LabelSelector     LabelSelector `json:"labelSelector"`
}
type Deployment struct {
	Annotations               Annotations                 `json:"annotations"`
	Labels                    Labels                      `json:"labels"`
	Image                     string                      `json:"image"`
	Replicas                  int                         `json:"replicas"`
	Pods                      CoreDNSDeploymentPods       `json:"pods"`
	NodeSelector              NodeSelector                `json:"nodeSelector"`
	Resources                 Resources                   `json:"resources"`
	TopologySpreadConstraints []TopologySpreadConstraints `json:"topologySpreadConstraints"`
}
type Coredns struct {
	Enabled            bool           `json:"enabled"`
	Embedded           bool           `json:"embedded"`
	OverwriteManifests string         `json:"overwriteManifests"`
	OverwriteConfig    string         `json:"overwriteConfig"`
	PriorityClassName  string         `json:"priorityClassName"`
	Service            CoreDNSService `json:"service"`
	Deployment         Deployment     `json:"deployment"`
}

type VclusterService struct {
	Enabled         bool                `json:"enabled"`
	Labels          Labels              `json:"labels"`
	Annotations     []string            `json:"annotations"`
	KubeletNodePort int                 `json:"kubeletNodePort"`
	HTTPSNodePort   int                 `json:"httpsNodePort"`
	Spec            VclusterServiceSpec `json:"spec"`
}

type IngressAnnotations struct {
	NginxIngressKubernetesIoBackendProtocol string `json:"nginx.ingress.kubernetes.io/backend-protocol"`
	NginxIngressKubernetesIoSslPassthrough  string `json:"nginx.ingress.kubernetes.io/ssl-passthrough"`
	NginxIngressKubernetesIoSslRedirect     string `json:"nginx.ingress.kubernetes.io/ssl-redirect"`
}

type IngressTLSSpec struct {
	TLS []networkingv1.IngressTLS `json:"tls"`
}

type Ingress struct {
	Enabled     bool               `json:"enabled"`
	Host        string             `json:"host"`
	PathType    string             `json:"pathType"`
	Labels      Labels             `json:"labels"`
	Annotations IngressAnnotations `json:"annotations"`
	Spec        IngressTLSSpec     `json:"spec"`
}
type Annotations struct {
}
type ControlPlaneStatefulsetHighAvailability struct {
	Replicas      int `json:"replicas"`
	LeaseDuration int `json:"leaseDuration"`
	RenewDeadline int `json:"renewDeadline"`
	RetryPeriod   int `json:"retryPeriod"`
}
type EmptyDir struct {
}
type BinariesVolume struct {
	Name     string   `json:"name"`
	EmptyDir EmptyDir `json:"emptyDir"`
}
type ControlPlaneStatefulsetPersistence struct {
	VolumeClaim          VolumeClaim      `json:"volumeClaim"`
	VolumeClaimTemplates []string         `json:"volumeClaimTemplates"`
	AddVolumeMounts      []string         `json:"addVolumeMounts"`
	AddVolumes           []string         `json:"addVolumes"`
	BinariesVolume       []BinariesVolume `json:"binariesVolume"`
}
type LivenessProbe struct {
	Enabled bool `json:"enabled"`
}
type ReadinessProbe struct {
	Enabled bool `json:"enabled"`
}
type StartupProbe struct {
	Enabled bool `json:"enabled"`
}
type Probes struct {
	LivenessProbe  LivenessProbe  `json:"livenessProbe"`
	ReadinessProbe ReadinessProbe `json:"readinessProbe"`
	StartupProbe   StartupProbe   `json:"startupProbe"`
}
type ControlPlaneStatefulSet struct {
	Labels             Labels                                  `json:"labels"`
	Annotations        Annotations                             `json:"annotations"`
	ImagePullPolicy    string                                  `json:"imagePullPolicy"`
	Image              Image                                   `json:"image"`
	WorkingDir         string                                  `json:"workingDir"`
	Command            []string                                `json:"command"`
	Args               []string                                `json:"args"`
	Env                []string                                `json:"env"`
	Resources          Resources                               `json:"resources"`
	Pods               Pods                                    `json:"pods"`
	HighAvailability   ControlPlaneStatefulsetHighAvailability `json:"highAvailability"`
	Security           Security                                `json:"security"`
	Persistence        ControlPlaneStatefulsetPersistence      `json:"persistence"`
	EnableServiceLinks bool                                    `json:"enableServiceLinks"`
	Scheduling         Scheduling                              `json:"scheduling"`
	Probes             Probes                                  `json:"probes"`
}
type ServiceMonitor struct {
	Enabled     bool        `json:"enabled"`
	Labels      Labels      `json:"labels"`
	Annotations Annotations `json:"annotations"`
}
type VirtualScheduler struct {
	Enabled bool `json:"enabled"`
}
type ServiceAccount struct {
	Enabled          bool        `json:"enabled"`
	Name             string      `json:"name"`
	ImagePullSecrets []string    `json:"imagePullSecrets"`
	Labels           Labels      `json:"labels"`
	Annotations      Annotations `json:"annotations"`
}
type WorkloadServiceAccount struct {
	Enabled          bool        `json:"enabled"`
	Name             string      `json:"name"`
	ImagePullSecrets []string    `json:"imagePullSecrets"`
	Annotations      Annotations `json:"annotations"`
	Labels           Labels      `json:"labels"`
}
type HeadlessService struct {
	Labels      Labels      `json:"labels"`
	Annotations Annotations `json:"annotations"`
}
type GlobalMetadata struct {
	Annotations Annotations `json:"annotations"`
}
type ControlPlaneStatefulsetAdvancedOptions struct {
	DefaultImageRegistry   string                 `json:"defaultImageRegistry"`
	VirtualScheduler       VirtualScheduler       `json:"virtualScheduler"`
	ServiceAccount         ServiceAccount         `json:"serviceAccount"`
	WorkloadServiceAccount WorkloadServiceAccount `json:"workloadServiceAccount"`
	HeadlessService        HeadlessService        `json:"headlessService"`
	GlobalMetadata         GlobalMetadata         `json:"globalMetadata"`
}
type ControlPlane struct {
	Distro         Distro                                 `json:"distro"`
	BackingStore   BackingStore                           `json:"backingStore"`
	Proxy          Proxy                                  `json:"proxy"`
	Coredns        Coredns                                `json:"coredns"`
	Service        VclusterService                        `json:"service"`
	Ingress        Ingress                                `json:"ingress"`
	StatefulSet    ControlPlaneStatefulSet                `json:"statefulSet"`
	ServiceMonitor ServiceMonitor                         `json:"serviceMonitor"`
	Advanced       ControlPlaneStatefulsetAdvancedOptions `json:"advanced"`
}
type MetricsServer struct {
	Enabled bool `json:"enabled"`
	Nodes   bool `json:"nodes"`
	Pods    bool `json:"pods"`
}
type Webhook struct {
	Enabled bool `json:"enabled"`
}
type DataVolumes struct {
	Enabled bool `json:"enabled"`
}

type Rbac struct {
	Role        Role        `json:"role"`
	ClusterRole ClusterRole `json:"clusterRole"`
}
type ReplicateServices struct {
	ToHost   []string `json:"toHost"`
	FromHost []string `json:"fromHost"`
}
type ProxyKubelets struct {
	ByHostname bool `json:"byHostname"`
	ByIP       bool `json:"byIP"`
}
type Advanced struct {
	ClusterDomain       string        `json:"clusterDomain"`
	FallbackHostCluster bool          `json:"fallbackHostCluster"`
	ProxyKubelets       ProxyKubelets `json:"proxyKubelets"`
}

type Quota struct {
	RequestsCPU                 int    `json:"requests.cpu"`
	RequestsMemory              string `json:"requests.memory"`
	RequestsStorage             string `json:"requests.storage"`
	RequestsEphemeralStorage    string `json:"requests.ephemeral-storage"`
	LimitsCPU                   int    `json:"limits.cpu"`
	LimitsMemory                string `json:"limits.memory"`
	LimitsEphemeralStorage      string `json:"limits.ephemeral-storage"`
	ServicesNodeports           int    `json:"services.nodeports"`
	ServicesLoadbalancers       int    `json:"services.loadbalancers"`
	CountEndpoints              int    `json:"count/endpoints"`
	CountPods                   int    `json:"count/pods"`
	CountServices               int    `json:"count/services"`
	CountSecrets                int    `json:"count/secrets"`
	CountConfigmaps             int    `json:"count/configmaps"`
	CountPersistentvolumeclaims int    `json:"count/persistentvolumeclaims"`
}
type ScopeSelector struct {
	MatchExpressions []string `json:"matchExpressions"`
}
type ResourceQuota struct {
	Enabled       bool          `json:"enabled"`
	Labels        Labels        `json:"labels"`
	Annotations   Annotations   `json:"annotations"`
	Quota         Quota         `json:"quota"`
	ScopeSelector ScopeSelector `json:"scopeSelector"`
	Scopes        []string      `json:"scopes"`
}
type Default struct {
	EphemeralStorage string `json:"ephemeral-storage"`
	Memory           string `json:"memory"`
	CPU              string `json:"cpu"`
}
type DefaultRequest struct {
	EphemeralStorage string `json:"ephemeral-storage"`
	Memory           string `json:"memory"`
	CPU              string `json:"cpu"`
}
type LimitRange struct {
	Enabled        bool           `json:"enabled"`
	Labels         Labels         `json:"labels"`
	Annotations    Annotations    `json:"annotations"`
	Default        Default        `json:"default"`
	DefaultRequest DefaultRequest `json:"defaultRequest"`
}
type IPBlock struct {
	Cidr   string   `json:"cidr"`
	Except []string `json:"except"`
}
type OutgoingConnections struct {
	IPBlock IPBlock `json:"ipBlock"`
}
type NetworkPolicy struct {
	Enabled             bool                `json:"enabled"`
	Labels              Labels              `json:"labels"`
	Annotations         Annotations         `json:"annotations"`
	FallbackDNS         string              `json:"fallbackDns"`
	OutgoingConnections OutgoingConnections `json:"outgoingConnections"`
}
type CentralAdmission struct {
	ValidatingWebhooks []string `json:"validatingWebhooks"`
	MutatingWebhooks   []string `json:"mutatingWebhooks"`
}
type Policies struct {
	ResourceQuota    ResourceQuota    `json:"resourceQuota"`
	LimitRange       LimitRange       `json:"limitRange"`
	NetworkPolicy    NetworkPolicy    `json:"networkPolicy"`
	CentralAdmission CentralAdmission `json:"centralAdmission"`
}
type Secret struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
type ExportKubeConfig struct {
	Context string `json:"context"`
	Server  string `json:"server"`
	Secret  Secret `json:"secret"`
}
type External struct {
}
type Plugins struct {
}
type MultiNamespaceMode struct {
	Enabled bool `json:"enabled"`
}
type SyncSettings struct {
	DisableSync              bool     `json:"disableSync"`
	SyncLabels               []string `json:"syncLabels"`
	RewriteKubernetesService bool     `json:"rewriteKubernetesService"`
	TargetNamespace          string   `json:"targetNamespace"`
	SetOwner                 bool     `json:"setOwner"`
}
type IsolatedControlPlane struct {
	Headless bool `json:"headless"`
}
type Host struct {
	Manifests         string `json:"manifests"`
	ManifestsTemplate string `json:"manifestsTemplate"`
}

type ClusterRole struct {
	ExtraRules []string `json:"extraRules"`
}
type Role struct {
	ExtraRules []string `json:"extraRules"`
}
type GenericSync struct {
	ClusterRole ClusterRole `json:"clusterRole"`
	Role        Role        `json:"role"`
}
type Experimental struct {
	MultiNamespaceMode   MultiNamespaceMode   `json:"multiNamespaceMode"`
	SyncSettings         SyncSettings         `json:"syncSettings"`
	IsolatedControlPlane IsolatedControlPlane `json:"isolatedControlPlane"`
	GenericSync          GenericSync          `json:"genericSync"`
}
type Telemetry struct {
	Enabled bool `json:"enabled"`
}
