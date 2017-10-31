package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	createState := Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
	}
	for _, currentIP := range currentEndpoint.IPs {
		if !containsIP(createState.IPs, currentIP) {
			createState.IPs = append(createState.IPs, currentIP)
		}
	}
	for _, desiredIP := range desiredEndpoint.IPs {
		if !containsIP(createState.IPs, desiredIP) {
			createState.IPs = append(createState.IPs, desiredIP)
		}
	}

	return createState, nil
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	endpointToCreate, err := toEndpoint(createState)
	if err != nil {
		return microerror.Mask(err)
	}

	if len(endpointToCreate.IPs) == 0 {
		// Nothing to do.
		return nil
	}

	k8sAddresses := []apiv1.EndpointAddress{}
	for _, endpointIP := range endpointToCreate.IPs {
		k8sAddress := apiv1.EndpointAddress{
			IP: endpointIP,
		}
		k8sAddresses = append(k8sAddresses, k8sAddress)
	}

	k8sEndpoint, err := r.k8sClient.CoreV1().Endpoints(endpointToCreate.ServiceNamespace).Get(endpointToCreate.ServiceName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}
	for i := range k8sEndpoint.Subsets {
		k8sEndpoint.Subsets[i].Addresses = k8sAddresses
	}

	_, err = r.k8sClient.CoreV1().Endpoints(endpointToCreate.ServiceNamespace).Update(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
