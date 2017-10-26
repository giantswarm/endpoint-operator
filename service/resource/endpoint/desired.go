package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	endpointIP, serviceName, err := getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
	if err != nil {
		return nil, microerror.Maskf(err, "an error occurred while fetching the annotations of the pod")
	}

	endpoint := Endpoint{
		IP:               endpointIP,
		ServiceName:      serviceName,
		ServiceNamespace: pod.GetNamespace(),
	}

	desiredEndpoints := []Endpoint{
		endpoint,
	}
	return desiredEndpoints, nil
}
