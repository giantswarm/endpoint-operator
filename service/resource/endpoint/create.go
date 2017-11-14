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

	desiredEndpoint, err := toEndpoint(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	createState := Endpoint{
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

	k8sEndpoint, err := newK8sEndpoint(createState, r.k8sClient)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return k8sEndpoint, nil
}
