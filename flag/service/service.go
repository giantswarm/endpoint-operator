package service

import (
	"github.com/giantswarm/endpoint-operator/flag/service/kubernetes"
	"github.com/giantswarm/endpoint-operator/flag/service/operator"
	"github.com/giantswarm/endpoint-operator/flag/service/resource"
)

type Service struct {
	Kubernetes kubernetes.Kubernetes
	Operator   operator.Operator
	Resource   resource.Resource
}
