// Copyright 2021 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package controllers

import (
	"go.uber.org/fx"

	"k8s.io/kubernetes/controllers/chaosimpl"
	"k8s.io/kubernetes/controllers/common"
	"k8s.io/kubernetes/controllers/schedule"
	"k8s.io/kubernetes/controllers/statuscheck"
	"k8s.io/kubernetes/controllers/utils/chaosdaemon"
	"k8s.io/kubernetes/controllers/utils/recorder"
	wfcontrollers "k8s.io/kubernetes/pkg/workflow/controllers"
)

var Module = fx.Options(
	fx.Provide(
		chaosdaemon.New,
		recorder.NewRecorderBuilder,
		common.AllSteps,
	),
	fx.Invoke(common.Bootstrap),
	fx.Invoke(wfcontrollers.BootstrapWorkflowControllers),
	fx.Invoke(statuscheck.Bootstrap),

	schedule.Module,
	chaosimpl.AllImpl)
