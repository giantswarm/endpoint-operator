package endpoint

import (
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

type Endpoint struct {
	EndpointAddress  apiv1.EndpointAddress
	ServiceName      string
	ServiceNamespace string
}
