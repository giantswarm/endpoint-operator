package main

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/microkit/command"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/microkit/transaction"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/microstorage"
	"github.com/giantswarm/microstorage/memory"

	"github.com/giantswarm/endpoint-operator/flag"
	"github.com/giantswarm/endpoint-operator/server"
	"github.com/giantswarm/endpoint-operator/service"
)

var (
	f *flag.Flag = flag.New()

	description string = "The endpoint-operator handles IPs inside of endpoints based on pod annotations."
	gitCommit   string = "n/a"
	name        string = "endpoint-operator"
	source      string = "https://github.com/giantswarm/endpoint-operator"
)

func panicOnErr(err error) {
	if err != nil {
		panic(fmt.Sprintf("%#v", err))
	}
}

func main() {
	err := mainWithError()
	if err != nil {
		panic(fmt.Sprintf("%#v\n", microerror.Mask(err)))
	}
}

func mainWithError() error {
	var err error

	var newLogger micrologger.Logger
	{
		loggerConfig := micrologger.DefaultConfig()

		loggerConfig.IOWriter = os.Stdout

		newLogger, err = micrologger.New(loggerConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	newServerFactory := func(v *viper.Viper) microserver.Server {
		var newService *service.Service
		{
			serviceConfig := service.DefaultConfig()

			serviceConfig.Flag = f
			serviceConfig.Logger = newLogger
			serviceConfig.Viper = v

			serviceConfig.Description = description
			serviceConfig.GitCommit = gitCommit
			serviceConfig.Name = name
			serviceConfig.Source = source

			newService, err = service.New(serviceConfig)
			panicOnErr(err)

			go newService.Boot()
		}

		var newStorage microstorage.Storage
		{
			storageConfig := memory.DefaultConfig()

			newStorage, err = memory.New(storageConfig)
			panicOnErr(err)
		}

		var newTransactionResponder transaction.Responder
		{
			transactionResponderConfig := transaction.DefaultResponderConfig()

			transactionResponderConfig.Logger = newLogger
			transactionResponderConfig.Storage = newStorage

			newTransactionResponder, err = transaction.NewResponder(transactionResponderConfig)
			panicOnErr(err)
		}

		var newServer microserver.Server
		{
			serverConfig := server.DefaultConfig()

			serverConfig.MicroServerConfig.Logger = newLogger
			serverConfig.MicroServerConfig.TransactionResponder = newTransactionResponder
			serverConfig.MicroServerConfig.Viper = v
			serverConfig.Service = newService

			serverConfig.MicroServerConfig.ServiceName = name

			newServer, err = server.New(serverConfig)
			panicOnErr(err)
		}

		return newServer
	}

	var newCommand command.Command
	{
		commandConfig := command.DefaultConfig()

		commandConfig.Logger = newLogger
		commandConfig.ServerFactory = newServerFactory

		commandConfig.Description = description
		commandConfig.GitCommit = gitCommit
		commandConfig.Name = name
		commandConfig.Source = source

		newCommand, err = command.New(commandConfig)
		if err != nil {
			return microerror.Mask(err)
		}
	}

	daemonCommand := newCommand.DaemonCommand().CobraCommand()

	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.Address, "http://127.0.0.1:6443", "Address used to connect to Kubernetes. When empty in-cluster config is created.")
	daemonCommand.PersistentFlags().Bool(f.Service.Kubernetes.InCluster, false, "Whether to use the in-cluster config to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CAFile, "", "Certificate authority file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.CrtFile, "", "Certificate file path to use to authenticate with Kubernetes.")
	daemonCommand.PersistentFlags().String(f.Service.Kubernetes.TLS.KeyFile, "", "Key file path to use to authenticate with Kubernetes.")

	newCommand.CobraCommand().Execute()

	return nil
}
