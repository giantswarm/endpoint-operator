package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) ApplyUpdateChange(ctx context.Context, obj, updateState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(updateState)
	if err != nil {
		return microerror.Mask(err)
	}
	if k8sEndpoint == nil {
		return nil // Nothing to do.
	}

	if isEmptyEndpoint(*k8sEndpoint) {
		return nil
	}

	_, err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Update(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	createState, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	updateState, err := r.newUpdateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()

	patch.SetCreateChange(createState)
	patch.SetUpdateChange(updateState)

	return patch, nil
}

func (r *Resource) newUpdateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*apiv1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if currentEndpoint == nil {
		return nil, nil // The endpoint does not exist, we should create it instead.
	}

	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if desiredEndpoint == nil {
		return nil, nil // Nothing to do.
	}

	createState := &Endpoint{
		ServiceName:      desiredEndpoint.ServiceName,
		ServiceNamespace: desiredEndpoint.ServiceNamespace,
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

	if len(createState.IPs) == 0 {
		// Nothing to do.
		return nil, nil
	}

	k8sEndpoint, err := r.newK8sEndpoint(createState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sEndpoint, nil
}
