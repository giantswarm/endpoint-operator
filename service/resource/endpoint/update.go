package endpoint

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework"
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
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return microerror.Mask(err)
	}

	return nil
}

func (r *Resource) NewUpdatePatch(ctx context.Context, obj, currentState, desiredState interface{}) (*framework.Patch, error) {
	k8sEndpoint, err := r.newCreateChange(ctx, obj, currentState, desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	patch := framework.NewPatch()

	patch.SetCreateChange(k8sEndpoint)
	patch.SetUpdateChange(k8sEndpoint)

	return patch, nil
}
