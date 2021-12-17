/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	ccmv1alpha1 "indeed.com/compute-platform/cluster-config-map/apis/clusterconfigmap/v1alpha1"
	"indeed.com/compute-platform/cluster-config-map/pkg/ccm"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(ccmv1alpha1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var endpoint string

	opts := zap.Options{}
	opts.BindFlags(flag.CommandLine)
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&endpoint, "endpoint", "unix:///var/lib/kubelet/plugins/clusterconfigmaps.indeed.com/csi.sock", "CSI endpoint")
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseDevMode(true), zap.UseFlagOptions(&opts)))

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(ccm.Metrics, promhttp.HandlerOpts{}))
		err := http.ListenAndServe(metricsAddr, mux)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "failed to serve metrics: "+err.Error())
			os.Exit(2)
		}
	}()

	drv, err := ccm.NewDriver(endpoint)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to create csi driver: "+err.Error())
		return
	}
	if err := drv.Run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "failed to run csi driver: "+err.Error())
		return
	}
}
