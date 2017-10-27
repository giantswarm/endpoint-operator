package endpoint

import (
	"context"

	"github.com/giantswarm/microerror"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func (r *Resource) GetCreateState(ctx context.Context, obj, currentState, desiredState interface{}) (interface{}, error) {
	currentEndpoints, err := toEndpointSlice(currentState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	desiredEndpoints, err := toEndpointSlice(desiredState)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	var createState = []Endpoint{}
	for _, desiredEndpoint := range desiredEndpoints {
		if !containsEndpoint(currentEndpoints, desiredEndpoint) {
			createState = append(createState, desiredEndpoint)
		}
	}

	if len(createState) == 0 {
		return nil, nil
	}

	return createState, nil
}

func (r *Resource) ProcessCreateState(ctx context.Context, obj, createState interface{}) error {
	endpointsToCreate, err := toEndpointSlice(createState)
	if err != nil {
		return microerror.Mask(err)
	}
	for _, endpoint := range endpointsToCreate {
		k8sService, err := r.k8sClient.CoreV1().Services(endpoint.ServiceNamespace).Get(endpoint.ServiceName, metav1.GetOptions{})
		if err != nil {
			return microerror.Mask(err)
		}
		k8sEndpoint := &apiv1.Endpoints{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: endpoint.ServiceName,
			},
			Subsets: []apiv1.EndpointSubset{
				{
					Addresses: []apiv1.EndpointAddress{
						{
							IP: endpoint.IP,
						},
					},
					Ports: serviceToPorts(k8sService),
				},
			},
		}
		_, err = r.k8sClient.CoreV1().Endpoints(endpoint.ServiceNamespace).Create(k8sEndpoint)
		if errors.IsAlreadyExists(err) {
			// In case the endpoint we tried to create does already exist, we only
			// need to append the IPs we got in podInfos.
			err = appendIPs(r, endpoint.ServiceNamespace, endpoint.ServiceName, endpoint.IP)
			if err != nil {
				return microerror.Mask(err)
			}
		} else if err != nil {
			return microerror.Mask(err)
		}

	}

	return nil
}

func appendIPs(r *Resource, serviceNamespace string, serviceName string, IP string) error {
	k8sEndpoint, err := r.k8sClient.CoreV1().Endpoints(serviceNamespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return microerror.Mask(err)
	}

	for i, k8sSubset := range k8sEndpoint.Subsets {
		addresses := k8sSubset.Addresses
		if ipInAddresses(addresses, IP) {
			continue
		}

		address := apiv1.EndpointAddress{
			IP: IP,
		}

		addresses = append(addresses, address)
		k8sEndpoint.Subsets[i].Addresses = addresses

	}

	_, err = r.k8sClient.CoreV1().Endpoints(serviceNamespace).Update(k8sEndpoint)
	if err != nil {
		return microerror.Mask(err)
	}
	return nil
}

func ipInAddresses(addresses []apiv1.EndpointAddress, IP string) bool {
	for _, a := range addresses {
		if a.IP == IP {
			return true
		}
	}

	return false
}

func serviceToPorts(s *apiv1.Service) []apiv1.EndpointPort {
	var ports []apiv1.EndpointPort

	for _, p := range s.Spec.Ports {
		port := apiv1.EndpointPort{
			Name: p.Name,
			Port: p.Port,
		}

		ports = append(ports, port)
	}

	return ports
}
