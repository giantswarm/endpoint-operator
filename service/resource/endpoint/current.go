package endpoint

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/operatorkit/framework/context/canceledcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	podWatcherLabel = "kvm-operator.giantswarm.io/pod-watcher"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	pod, err := toPod(obj)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	r.logger.Log("pod", pod.GetName(), "debug", "looking for annotations on pod")
	serviceNamespace := pod.GetNamespace()
	_, serviceName, err := getAnnotations(*pod, IPAnnotation, ServiceAnnotation)
	if IsMissingAnnotationError(err) {
		canceledcontext.SetCanceled(ctx)
		if canceledcontext.IsCanceled(ctx) {
			r.logger.Log("pod", pod.GetName(), "debug", fmt.Sprintf("canceling reconciliation for pod,%#v", microerror.Mask(err)))
			return nil, nil
		}
	} else if err != nil {
		return nil, microerror.Maskf(err, "an error occurred while fetching the annotations of the pod")
	}

	{
		_, ok := pod.GetLabels()[podWatcherLabel]
		if ok {
			canceledcontext.SetCanceled(ctx)
			if canceledcontext.IsCanceled(ctx) {
				r.logger.Log("pod", pod.GetName(), "debug", "canceling reconciliation for pod due to pod watcher annotation")
				return nil, nil
			}
		}
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
	return &currentEndpoint, nil
}
