package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) ApplyDeleteChange(ctx context.Context, obj, deleteState interface{}) error {
	if deleteState == nil {
		return nil // Nothing to do.
	}

	k8sEndpoint, err := toK8sEndpoint(deleteState)
	if err != nil {
		return microerror.Mask(err)
	}

	if !isEmptyEndpoint(*k8sEndpoint) {
		return nil
	}

	err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Delete(k8sEndpoint.Name, &metav1.DeleteOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) NewDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	deleteState, err := r.newDeleteChangeForDeletePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	updateState, err := r.newDeleteChangeForUpdatePatch(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	patch := framework.NewPatch()
	patch.SetDeleteChange(deleteState)
	patch.SetUpdateChange(updateState)

	return patch, nil
}

func (r *Resource) newDeleteChangeForDeletePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*apiv1.Endpoints, error) {
	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	deleteState := Endpoint{
		ServiceName:      currentEndpoint.ServiceName,
		ServiceNamespace: currentEndpoint.ServiceNamespace,
		IPs:              cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs),
	}
	if len(deleteState.IPs) > 0 {
		return nil, nil
	}

	k8sEndpoint, err := newK8sEndpoint(deleteState, r.k8sClient)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sEndpoint, nil
}

func (r *Resource) newDeleteChangeForUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*apiv1.Endpoints, error) {
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
		IPs:              cutIPs(currentEndpoint.IPs, desiredEndpoint.IPs),
	}

	if len(deleteState.IPs) == 0 {
		return nil, nil
	}

	k8sEndpoint, err := newK8sEndpoint(deleteState, r.k8sClient)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sEndpoint, nil
}
