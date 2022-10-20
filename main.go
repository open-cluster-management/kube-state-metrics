/*
Copyright 2015 The Kubernetes Authors All rights reserved.

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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/prometheus/common/version"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"

	"k8s.io/kube-state-metrics/v2/pkg/customresource"
	"k8s.io/kube-state-metrics/v2/pkg/customresourcestate"

	"k8s.io/kube-state-metrics/v2/pkg/app"
	"k8s.io/kube-state-metrics/v2/pkg/options"
)

func main() {
	opts := options.NewOptions()
	opts.AddFlags()

	if err := opts.Parse(); err != nil {
		klog.ErrorS(err, "Parsing flag definitions error")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	if opts.Version {
		fmt.Printf("%s\n", version.Print("kube-state-metrics"))
		os.Exit(0)
	}

	if opts.Help {
		opts.Usage()
		os.Exit(0)
	}

	var factories []customresource.RegistryFactory
	if config, set := resolveCustomResourceConfig(opts); set {
		crf, err := customresourcestate.FromConfig(config)
		if err != nil {
			klog.ErrorS(err, "Parsing from Custom Resource State Metrics file failed")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		factories = append(factories, crf...)
	}

	ctx := context.Background()
	if err := app.RunKubeStateMetrics(ctx, opts, factories...); err != nil {
		klog.ErrorS(err, "Failed to run kube-state-metrics")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}

func resolveCustomResourceConfig(opts *options.Options) (customresourcestate.ConfigDecoder, bool) {
	if s := opts.CustomResourceConfig; s != "" {
		return yaml.NewDecoder(strings.NewReader(s)), true
	}
	if file := opts.CustomResourceConfigFile; file != "" {
		f, err := os.Open(file)
		if err != nil {
			klog.ErrorS(err, "Custom Resource State Metrics file could not be opened")
			klog.FlushAndExit(klog.ExitFlushTimeout, 1)
		}
		return yaml.NewDecoder(f), true
	}
	return nil, false
}
