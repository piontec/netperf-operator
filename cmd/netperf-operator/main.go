package main

import (
	"context"
	"runtime"

	"github.com/piontec/netperf-operator/pkg/netperf-operator"

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	k8sutil "github.com/operator-framework/operator-sdk/pkg/util/k8sutil"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/piontec/netperf-operator/pkg/apis/app/realkube"
	stub "github.com/piontec/netperf-operator/pkg/stub"

	"github.com/sirupsen/logrus"
)

const version = "0.1.3-dev"

func printVersion() {
	logrus.Infof("Go Version: %s", runtime.Version())
	logrus.Infof("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logrus.Infof("operator-sdk Version: %v", sdkVersion.Version)
	logrus.Infof("Netperf-operator Version: %v", version)
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	printVersion()

	resource := "app.example.com/v1alpha1"
	kind := "Netperf"
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		logrus.Fatalf("Failed to get watch namespace: %v", err)
	}
	resyncPeriod := 5
	logrus.Infof("Watching %s, %s, %s, %d", resource, kind, namespace, resyncPeriod)
	sdk.Watch(resource, kind, namespace, resyncPeriod)
	sdk.Watch("v1", "Pod", namespace, resyncPeriod)
	sdk.Handle(stub.NewHandler(operator.NewNetperf(realkube.NewRealProvider())))
	sdk.Run(context.TODO())
}
