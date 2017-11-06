package endpoint

import (
	"context"
	"reflect"

	"github.com/giantswarm/microerror"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) GetDeleteState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	deleteState := Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
	}
	for _, currentIP := range currentEndpoint.IPs {
		if !containsIP(deleteState.IPs, currentIP) {
			deleteState.IPs = append(deleteState.IPs, currentIP)
		}
	}
	for _, desiredIP := range desiredEndpoint.IPs {
		if containsIP(deleteState.IPs, desiredIP) {
			deleteState.IPs = removeIP(deleteState.IPs, desiredIP)
		}
	}

	return deleteState, nil
}

func (r *Resource) ProcessDeleteState(ctx context.Context, obj, deleteState interface{}) error {
	endpointToApply, err := toEndpoint(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if reflect.DeepEqual(endpointToApply, Endpoint{}) {
		return nil
	}

	k8sAddresses := []apiv1.EndpointAddress{}
	for _, endpointIP := range endpointToApply.IPs {
		k8sAddress := apiv1.EndpointAddress{
			IP: endpointIP,
		}
		k8sAddresses = append(k8sAddresses, k8sAddress)
	}

	if len(k8sAddresses) == 0 {
		err = r.k8sClient.CoreV1().Endpoints(endpointToApply.ServiceNamespace).Delete(endpointToApply.ServiceName, &metav1.DeleteOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		return nil
	}

	k8sEndpoint, err := r.k8sClient.CoreV1().Endpoints(endpointToApply.ServiceNamespace).Get(endpointToApply.ServiceName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}
	for i := range k8sEndpoint.Subsets {
		k8sEndpoint.Subsets[i].Addresses = k8sAddresses
	}

	_, err = r.k8sClient.CoreV1().Endpoints(endpointToApply.ServiceNamespace).Update(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
