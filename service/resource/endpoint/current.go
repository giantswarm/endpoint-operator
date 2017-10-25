package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
)

const (
	ipAnnotation      = "endpoint.kvm.giantswarm.io/ip"
	serviceAnnotation = "endpoint.kvm.giantswarm.io/service"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {

	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	annotations, err := getAnnotations(*pod, []string{ipAnnotation, serviceAnnotation})
	if err != nil {
		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.Log("pod", pod.GetName(), "debug", "canceling reconciliation for pod")
			return nil, nil
		}
	}

	endpoint := &Endpoint{
		IP:               annotations[ipAnnotation],
		ServiceName:      annotations[serviceAnnotation],
		ServiceNamespace: pod.GetNamespace(),
	}

	return endpoint, nil
}
