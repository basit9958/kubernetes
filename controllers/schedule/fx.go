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

package schedule

import (
	"go.uber.org/fx"

	"k8s.io/kubernetes/controllers/schedule/active"
	"k8s.io/kubernetes/controllers/schedule/cron"
	"k8s.io/kubernetes/controllers/schedule/gc"
	"k8s.io/kubernetes/controllers/schedule/pause"
	"k8s.io/kubernetes/controllers/schedule/utils"
)

var Module = fx.Options(
	fx.Invoke(cron.Bootstrap),
	fx.Invoke(active.Bootstrap),
	fx.Invoke(gc.Bootstrap),
	fx.Invoke(pause.Bootstrap),

	fx.Provide(utils.NewActiveLister),
)
