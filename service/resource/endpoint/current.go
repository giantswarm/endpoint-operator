package endpoint

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	serviceNamespace := pod.GetNamespace()
	_, serviceName, err := getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
	if IsMissingAnnotationError(err) {
		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.Log("pod", pod.GetName(), "debug", "canceling reconciliation for pod")
			return nil, nil
		}
	} else if err != nil {
		return nil, microerror.Maskf(err, "an error occurred while fetching the annotations of the pod")
	}

	currentEndpoint := Endpoint{
		IPs:              []string{},
		ServiceName:      serviceName,
		ServiceNamespace: serviceNamespace,
	}
	k8sEndpoints, err := r.k8sClient.CoreV1().Endpoints(serviceNamespace).Get(serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, microerror.Mask(err)
	}

	for _, endpointSubset := range k8sEndpoints.Subsets {
		for _, endpointAddress := range endpointSubset.Addresses {
			foundIP := endpointAddress.IP

			if !containsIP(currentEndpoint.IPs, foundIP) {
				currentEndpoint.IPs = append(currentEndpoint.IPs, foundIP)
			}
		}
	}
	return currentEndpoint, nil
}
