package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Delete(k8sEndpoint.Name, &metav1.DeleteOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	endpoint, err := r.newDeleteChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	k8sEndpoint, err := constructK8sEndpoint(endpoint, r.k8sClient)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()
	if len(endpoint.IPs) == 0 {
		patch.SetDeleteChange(k8sEndpoint)
	} else {
		patch.SetUpdateChange(k8sEndpoint)
	}

	return patch, nil
}

func (r *Resource) newDeleteChange(ctx context.Context, obj, currentState, desiredState interface{}) (Endpoint, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return Endpoint{}, microerror.Mask(err)
	}

	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return Endpoint{}, microerror.Mask(err)
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
