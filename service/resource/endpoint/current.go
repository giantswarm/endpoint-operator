package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"

	apiv1 "k8s.io/client-go/pkg/api/v1"
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

	endpoint := Endpoint{
		ServiceName:      podServiceName,
		ServiceNamespace: pod.GetNamespace(),
		EndpointAddress: apiv1.EndpointAddress{
			IP: podIP,

			// NodeName and TargetRef are optional but included here, to keep the EndpointAddress complete.
			NodeName: &pod.Spec.NodeName,
			TargetRef: &apiv1.ObjectReference{
				Kind:            "Pod",
				Namespace:       pod.ObjectMeta.Namespace,
				Name:            pod.ObjectMeta.Name,
				UID:             pod.ObjectMeta.UID,
				ResourceVersion: pod.ObjectMeta.ResourceVersion,
			},
		},
	}

	return endpoint, nil
}
