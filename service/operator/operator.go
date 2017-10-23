package operator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"k8s.io/client-go/kubernetes"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/informer"
)

type Config struct {
	BackOff   backoff.BackOff
	Framework *framework.Framework
	Informer  *informer.Informer
	K8sClient kubernetes.Interface
	Logger    micrologger.Logger

	ResyncPeriod time.Duration
}

func DefaultConfig() Config {
	return Config{
		BackOff:   nil,
		Framework: nil,
		Informer:  nil,
		K8sClient: nil,
		Logger:    nil,

		ResyncPeriod: 0,
	}
}

type Operator struct {
	// Dependencies.
	backOff   backoff.BackOff
	framework *framework.Framework
	informer  *informer.Informer
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
	if config.Framework == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Framework must not be empty")
	}
	if config.Informer == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Informer must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.ResyncPeriod == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.ResyncPeriod must not be zero")
	}

	newOperator := &Operator{
		// Dependencies.
		backOff:   config.BackOff,
		framework: config.Framework,
		informer:  config.Informer,
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

	deleteChan, updateChan, errChan := o.informer.Watch(context.TODO())
	o.framework.ProcessEvents(context.TODO(), deleteChan, updateChan, errChan)

	return nil
}
