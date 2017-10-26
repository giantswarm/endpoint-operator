package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetDesiredState(ctx context.Context, obj interface{}) (interface{}, error) {
	endpoint, err := toEndpoint(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredEndpoints := []Endpoint{
		*endpoint,
	}

	k8sEndpoints, err := r.k8sClient.CoreV1().Endpoints(endpoint.ServiceNamespace).Get(endpoint.ServiceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return desiredEndpoints, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, endpointSubset := range k8sEndpoints.Subsets {
		for _, endpointAddress := range endpointSubset.Addresses {
			foundEndpoint := Endpoint{
				IP:               endpointAddress.IP,
				ServiceName:      endpoint.ServiceName,
				ServiceNamespace: endpoint.ServiceNamespace,
			}
			if !isEndpointInEndpoints(foundEndpoint, desiredEndpoints) {
				desiredEndpoints = append(desiredEndpoints, foundEndpoint)
			}
		}
	}
	return desiredEndpoints, nil
}
