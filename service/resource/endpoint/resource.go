package endpoint

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	IPAnnotation      = "endpoint.kvm.giantswarm.io/ip"
	Name              = "endpoint"
	ServiceAnnotation = "endpoint.kvm.giantswarm.io/service"
)

type Config struct {
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger
}

func DefaultConfig() Config {
	return Config{
		K8sClient: nil,
		Logger:    nil,
	}
}

type Resource struct {
	k8sClient kubernetes.Interface
	logger    micrologger.Logger
}

func New(config Config) (*Resource, error) {
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}

	resource := &Resource{
		k8sClient: config.K8sClient,
		logger:    config.Logger,
	}
	return resource, nil
}

func (r *Resource) Name() string {
	return Name
}

func (r *Resource) Underlying() framework.Resource {
	return r
}

func (r *Resource) newK8sEndpoint(endpoint *Endpoint) (*apiv1.Endpoints, error) {
	k8sAddresses := []apiv1.EndpointAddress{}
	for _, endpointIP := range endpoint.IPs {
		k8sAddress := apiv1.EndpointAddress{
			IP: endpointIP,
		}
		k8sAddresses = append(k8sAddresses, k8sAddress)
	}

	k8sService, err := r.k8sClient.CoreV1().Services(endpoint.ServiceNamespace).Get(endpoint.ServiceName, metav1.GetOptions{})
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sEndpoint := &apiv1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      endpoint.ServiceName,
			Namespace: endpoint.ServiceNamespace,
		},
		Subsets: []apiv1.EndpointSubset{
			{
				Ports: serviceToPorts(k8sService),
			},
		},
	}

	for i := range k8sEndpoint.Subsets {
		k8sEndpoint.Subsets[i].Addresses = k8sAddresses
	}

	return k8sEndpoint, nil
}

func cutIPs(base []string, cutset []string) []string {
	resultIPs := []string{}
	// Deduplicate entries from base.
	for _, baseIP := range base {
		if !containsIP(resultIPs, baseIP) {
			resultIPs = append(resultIPs, baseIP)
		}
	}
	// Cut the cutset out of base.
	for _, cutsetIP := range cutset {
		if containsIP(resultIPs, cutsetIP) {
			resultIPs = removeIP(resultIPs, cutsetIP)
		}
	}
	return resultIPs
}

func containsIP(ips []string, ip string) bool {
	for _, foundIP := range ips {
		if foundIP == ip {
			return true
		}
	}
	return false
}

func getAnnotations(pod apiv1.Pod, ipAnnotationName string, serviceAnnotationName string) (ipAnnotationValue string, serviceAnnotationValue string, err error) {
	ipAnnotationValue, ok := pod.GetAnnotations()[ipAnnotationName]
	if !ok {
		return "", "", microerror.Maskf(missingAnnotationError, "expected annotation '%s' to be set", ipAnnotationName)
	}
	if ipAnnotationValue == "" {
		return "", "", microerror.Maskf(missingAnnotationError, "empty annotation '%s'", ipAnnotationName)
	}
	serviceAnnotationValue, ok = pod.GetAnnotations()[serviceAnnotationName]
	if !ok {
		return "", "", microerror.Maskf(missingAnnotationError, "expected annotation '%s' to be set", serviceAnnotationName)
	}
	return ipAnnotationValue, serviceAnnotationValue, nil
}

func isEmptyEndpoint(endpoint apiv1.Endpoints) bool {
	for _, subset := range endpoint.Subsets {
		if len(subset.Addresses) > 0 {
			return false
		}
	}
	return true
}

func removeIP(ips []string, ip string) []string {
	for index, foundIP := range ips {
		if foundIP == ip {
			return append(ips[:index], ips[index+1:]...)
		}
	}
	return ips
}

func serviceToPorts(s *apiv1.Service) []apiv1.EndpointPort {
	var ports []apiv1.EndpointPort

	for _, p := range s.Spec.Ports {
		port := apiv1.EndpointPort{
			Name: p.Name,
			Port: p.Port,
		}

		ports = append(ports, port)
	}

	return ports
}

func toEndpoint(v interface{}) (*Endpoint, error) {
	if v == nil {
		return nil, nil
	}
	endpoint, ok := v.(*Endpoint)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &Endpoint{}, v)
	}
	return endpoint, nil
}

func toK8sEndpoint(v interface{}) (*apiv1.Endpoints, error) {
	if v == nil {
		return nil, nil
	}
	k8sEndpoint, ok := v.(*apiv1.Endpoints)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.Endpoints{}, v)
	}
	return k8sEndpoint, nil
}

func toPod(v interface{}) (*apiv1.Pod, error) {
	if v == nil {
		return nil, nil
	}

	pod, ok := v.(*apiv1.Pod)
	if !ok {
		return nil, microerror.Maskf(wrongTypeError, "expected '%T', got '%T'", &apiv1.Pod{}, v)
	}

	return pod, nil
}
