package operator

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cenk/backoff"
	apismetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
)

type Config struct {
	BackOff   backoff.BackOff
	Framework *framework.Framework
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ResyncPeriod time.Duration
}

func DefaultConfig() Config {
	return Config{
		BackOff:   nil,
		Framework: nil,
		K8sClient: nil,
		Logger:    nil,

		ResyncPeriod: 0,
	}
}

type Operator struct {
	// Dependencies.
	backOff   backoff.BackOff
	framework *framework.Framework
	k8sClient kubernetes.Interface
	logger    micrologger.Logger

	// Internals.
	bootOnce     sync.Once
	mutex        sync.Mutex
	resyncPeriod time.Duration
}

func New(config Config) (*Operator, error) {
	if config.BackOff == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.BackOff must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.Framework == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Framework must not be empty")
	}

	if config.ResyncPeriod == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.ResyncPeriod must not be zero")
	}

	newOperator := &Operator{
		// Dependencies.
		backOff:   config.BackOff,
		framework: config.Framework,
		k8sClient: config.K8sClient,
		logger:    config.Logger,

		// Internals.
		bootOnce:     sync.Once{},
		resyncPeriod: config.ResyncPeriod,
	}

	return newOperator, nil
}

func (o *Operator) Boot() {
	o.bootOnce.Do(func() {
		operation := func() error {
			err := o.bootWithError()
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		notifier := func(err error, d time.Duration) {
			o.logger.Log("warning", fmt.Sprintf("retrying operator boot due to error: %#v", microerror.Mask(err)))
		}

		err := backoff.RetryNotify(operation, o.backOff, notifier)
		if err != nil {
			o.logger.Log("error", fmt.Sprintf("stop operator boot retries due to too many errors: %#v", microerror.Mask(err)))
			os.Exit(1)
		}
	})
}

func (o *Operator) bootWithError() error {
	o.logger.Log("debug", "starting list/watch")

	newResourceEventHandler := o.framework.NewCacheResourceEventHandler()

	listWatch := &cache.ListWatch{
		ListFunc: func(options apismetav1.ListOptions) (runtime.Object, error) {
			o.logger.Log("debug", "listing all pods", "event", "list")
			return o.k8sClient.CoreV1().Pods("").List(options)
		},
		WatchFunc: func(options apismetav1.ListOptions) (watch.Interface, error) {
			o.logger.Log("debug", "watching all pods", "event", "watch")
			return o.k8sClient.CoreV1().Pods("").Watch(options)
		},
	}

	_, informer := cache.NewInformer(
		listWatch,
		&v1.Pod{},
		o.resyncPeriod,
		newResourceEventHandler,
	)
	informer.Run(nil)

	return nil
}
