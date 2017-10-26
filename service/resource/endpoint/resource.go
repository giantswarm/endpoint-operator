package endpoint

import (
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

const (
	Name = "endpoint"
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

func getAnnotations(pod apiv1.Pod, ipAnnotationName string, serviceAnnotationName string) (ipAnnotationValue string, serviceAnnotationValue string, err error) {
	ipAnnotationValue, ok := pod.GetAnnotations()[ipAnnotationName]
	if !ok {
		return "", "", microerror.Maskf(missingAnnotationError, "expected annotation '%s' to be set", ipAnnotationName)
	}
	serviceAnnotationValue, ok = pod.GetAnnotations()[serviceAnnotationName]
	if !ok {
		return "", "", microerror.Maskf(missingAnnotationError, "expected annotation '%s' to be set", serviceAnnotationName)
	}
	return ipAnnotationValue, serviceAnnotationValue, nil
}
