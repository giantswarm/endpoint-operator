package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *Resource) GetCurrentState(ctx context.Context, obj interface{}) (interface{}, error) {
	r.logger.Log("debug", "fetching a list of all current endpoints")

	endpointList, err := r.k8sClient.CoreV1().Endpoints("").List(metav1.ListOptions{})
	if err != nil {
		return nil, microerror.Maskf(err, "an error occurred fetching the list endpoins")
	}

	r.logger.Log("debug", "found a list of all current endpoints")
	return endpointList, nil
}
