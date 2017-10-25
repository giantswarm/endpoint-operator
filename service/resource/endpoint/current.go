package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {

	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}
	podAnnotations := pod.GetAnnotations()
	podIP, ipOk := podAnnotations["endpoint.kvm.giantswarm.io/ip"]
	podServiceName, serviceNameOk := pod.Annotations["endpoint.kvm.giantswarm.io/service"]

	if !(ipOk && serviceNameOk) {
		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.Log("pod", pod.GetName(), "debug", "canceling reconciliation for pod")

			return nil, nil
		}
	}

	endpoint := &Endpoint{
		ServiceName:      podServiceName,
		ServiceNamespace: pod.GetNamespace(),
		IP:               podIP,
	}

	return endpoint, nil
}
