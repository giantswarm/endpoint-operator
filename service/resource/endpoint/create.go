package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) ApplyCreateChange(ctx context.Context, obj, createState interface{}) error {
	k8sEndpoint, err := toK8sEndpoint(createState)
	if err != nil {
		return microerror.Mask(err)
	}
	if k8sEndpoint == nil {
		return nil // Nothing to do.
	}

	_, err = r.k8sClient.CoreV1().Endpoints(k8sEndpoint.Namespace).Create(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) newCreateChange(ctx context.Context, obj, currentState, desiredState interface{}) (*apiv1.Endpoints, error) {
	currentEndpoint, err := toEndpoint(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if currentEndpoint != nil {
		return nil, nil // An endpoint exists already, we should update instead of create.
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
	for _, desiredIP := range desiredEndpoint.IPs {
		if !containsIP(createState.IPs, desiredIP) {
			createState.IPs = append(createState.IPs, desiredIP)
		}
	}

	if len(createState.IPs) == 0 {
		return nil, nil // Nothing to do.
	}

	k8sEndpoint, err := r.newK8sEndpoint(createState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sEndpoint, nil
}
