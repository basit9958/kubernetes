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

package types

import (
	"go.uber.org/fx"

	"k8s.io/kubernetes/api/v1alpha1"
)

type Controller string

type Object struct {
	Object v1alpha1.InnerObject
	Name   string
}

var ChaosObjects = fx.Supply(

	fx.Annotated{
		Group: "objs",
		Target: Object{
			Name:   "physicalmachinechaos",
			Object: &v1alpha1.PhysicalMachineChaos{},
		},
	},
)

type WebhookObject struct {
	Object v1alpha1.WebhookObject
	Name   string
}

var WebhookObjects = fx.Supply(
	fx.Annotated{
		Group: "webhookObjs",
		Target: WebhookObject{
			Name:   "physicalmachine",
			Object: &v1alpha1.PhysicalMachine{},
		},
	},
	//fx.Annotated{
	//	Group: "webhookObjs",
	//	Target: WebhookObject{
	//		Name:   "statuscheck",
	//		Object: &v1alpha1.StatusCheck{},
	//	},
	//},
)
