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

	ipAnnotationValue, serviceAnnotationValue, err := getAnnotations(*pod, ipAnnotation, serviceAnnotation)
	if IsMissingAnnotationError(err) {
		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.Log("pod", pod.GetName(), "debug", "canceling reconciliation for pod")
			return nil, nil
		}
	} else if err != nil {
		return nil, microerror.Maskf(err, "an error occurred while fetching the annotations of the pod")
	}

	endpoint := &Endpoint{
		IP:               ipAnnotationValue,
		ServiceName:      serviceAnnotationValue,
		ServiceNamespace: pod.GetNamespace(),
	}

	return endpoint, nil
}
