package service

import (
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	apiv1 "k8s.io/client-go/pkg/api/v1"

	"github.com/giantswarm/microendpoint/service/version"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/client/k8sclient"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/framework/resource/logresource"
	"github.com/giantswarm/operatorkit/framework/resource/metricsresource"
	"github.com/giantswarm/operatorkit/framework/resource/retryresource"
	"github.com/giantswarm/operatorkit/informer"

	"github.com/giantswarm/endpoint-operator/flag"
	"github.com/giantswarm/endpoint-operator/service/healthz"
	"github.com/giantswarm/endpoint-operator/service/operator"
	endpointresource "github.com/giantswarm/endpoint-operator/service/resource/endpoint"
)

const (
	ResourceRetries uint64 = 3
)

type Config struct {
	Flag   *flag.Flag
	Logger micrologger.Logger
	Viper  *viper.Viper

	Description string
	GitCommit   string
	Name        string
	Source      string
}

func DefaultConfig() Config {
	return Config{
		Flag:   nil,
		Logger: nil,
		Viper:  nil,

		Description: "",
		GitCommit:   "",
		Name:        "",
		Source:      "",
	}
}

type Service struct {
	Healthz  *healthz.Service
	Operator *operator.Operator
	Version  *version.Service

	bootOnce sync.Once
}

func New(config Config) (*Service, error) {
	if config.Flag == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Flag must not be empty")
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Viper must not be empty")
	}

	var err error

	var newK8sClient kubernetes.Interface
	{
		k8sConfig := k8sclient.DefaultConfig()

		k8sConfig.Logger = config.Logger

		k8sConfig.Address = config.Viper.GetString(config.Flag.Service.Kubernetes.Address)
		k8sConfig.InCluster = config.Viper.GetBool(config.Flag.Service.Kubernetes.InCluster)
		k8sConfig.TLS.CAFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CAFile)
		k8sConfig.TLS.CrtFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.CrtFile)
		k8sConfig.TLS.KeyFile = config.Viper.GetString(config.Flag.Service.Kubernetes.TLS.KeyFile)

		newK8sClient, err = k8sclient.New(k8sConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newHealthzService *healthz.Service
	{
		healthzConfig := healthz.DefaultConfig()

		healthzConfig.K8sClient = newK8sClient
		healthzConfig.Logger = config.Logger

		newHealthzService, err = healthz.New(healthzConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newOperatorBackOff *backoff.ExponentialBackOff
	{
		newOperatorBackOff = backoff.NewExponentialBackOff()
		newOperatorBackOff.MaxElapsedTime = 5 * time.Minute
	}

	var newEndpointResource framework.Resource
	{
		endpointConfig := endpointresource.DefaultConfig()

		endpointConfig.K8sClient = newK8sClient
		endpointConfig.Logger = config.Logger

		newEndpointResource, err = endpointresource.New(endpointConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var resources []framework.Resource
	{
		resources = []framework.Resource{
			newEndpointResource,
		}

		logWrapConfig := logresource.DefaultWrapConfig()
		logWrapConfig.Logger = config.Logger
		resources, err = logresource.Wrap(resources, logWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		retryWrapConfig := retryresource.DefaultWrapConfig()
		retryWrapConfig.BackOffFactory = func() backoff.BackOff {
			return backoff.WithMaxTries(backoff.NewExponentialBackOff(), ResourceRetries)
		}
		retryWrapConfig.Logger = config.Logger
		resources, err = retryresource.Wrap(resources, retryWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		metricsWrapConfig := metricsresource.DefaultWrapConfig()
		metricsWrapConfig.Name = config.Name
		resources, err = metricsresource.Wrap(resources, metricsWrapConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}
	var newWatcherFactory informer.WatcherFactory
	{

		zeroObjectFactory := &informer.ZeroObjectFactoryFuncs{
			NewObjectFunc: func() runtime.Object {
				var pod apiv1.Pod
				return &pod
			},
			NewObjectListFunc: func() runtime.Object {
				var podList apiv1.PodList
				return &podList
			},
		}
		newWatcherFactory = informer.NewWatcherFactory(newK8sClient.Discovery().RESTClient(), "/api/v1/watch/pods/", zeroObjectFactory)
	}

	var newInformer *informer.Informer
	{
		informerConfig := informer.DefaultConfig()

		informerConfig.BackOff = backoff.NewExponentialBackOff()
		informerConfig.WatcherFactory = newWatcherFactory

		informerConfig.ResyncPeriod = time.Minute * 5

		newInformer, err = informer.New(informerConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newOperatorFramework *framework.Framework
	{
		frameworkConfig := framework.DefaultConfig()

		newFrameworkBackOff := backoff.NewExponentialBackOff()
		newFrameworkBackOff.MaxElapsedTime = 5 * time.Minute

		frameworkConfig.BackOff = newFrameworkBackOff
		frameworkConfig.Logger = config.Logger
		frameworkConfig.Resources = resources

		newOperatorFramework, err = framework.New(frameworkConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newOperator *operator.Operator
	{
		operatorConfig := operator.DefaultConfig()

		operatorConfig.BackOff = newOperatorBackOff
		operatorConfig.Framework = newOperatorFramework
		operatorConfig.Informer = newInformer
		operatorConfig.K8sClient = newK8sClient
		operatorConfig.Logger = config.Logger

		operatorConfig.ResyncPeriod = time.Minute * 5

		newOperator, err = operator.New(operatorConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	var newVersionService *version.Service
	{
		versionConfig := version.DefaultConfig()

		versionConfig.Description = config.Description
		versionConfig.GitCommit = config.GitCommit
		versionConfig.Name = config.Name
		versionConfig.Source = config.Source

		newVersionService, err = version.New(versionConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		Healthz:  newHealthzService,
		Operator: newOperator,
		Version:  newVersionService,

		bootOnce: sync.Once{},
	}

	return newService, nil
}

func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		s.Operator.Boot()
	})
}
