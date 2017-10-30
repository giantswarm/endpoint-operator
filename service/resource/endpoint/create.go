package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoints, err := toEndpointSlice(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredEndpoints, err := toEndpointSlice(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	createState := []Endpoint{}
	for _, currentEndpoint := range currentEndpoints {
		if !containsEndpoint(createState, currentEndpoint) {
			createState = append(createState, currentEndpoint)
		}
	}
	for _, desiredEndpoint := range desiredEndpoints {
		if !containsEndpoint(createState, desiredEndpoint) {
			createState = append(createState, desiredEndpoint)
		}
	}

	return createState, nil
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	endpointsToCreate, err := toEndpointSlice(createState)
	if err != nil {
		return microerror.Mask(err)
	}
	var serviceName string
	var serviceNamespace string
	if len(endpointsToCreate) > 0 {
		serviceName = endpointsToCreate[0].ServiceName
		serviceNamespace = endpointsToCreate[0].ServiceNamespace
	} else {
		return nil
	}

	k8sAddresses := []apiv1.EndpointAddress{}
	for _, endpoint := range endpointsToCreate {
		k8sAddress := apiv1.EndpointAddress{
			IP: endpoint.IP,
		}
		k8sAddresses = append(k8sAddresses, k8sAddress)
	}

	k8sEndpoint, err := r.k8sClient.CoreV1().Endpoints(serviceNamespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}
	for i := range k8sEndpoint.Subsets {
		k8sEndpoint.Subsets[i].Addresses = k8sAddresses
	}

	_, err = r.k8sClient.CoreV1().Endpoints(serviceNamespace).Update(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
