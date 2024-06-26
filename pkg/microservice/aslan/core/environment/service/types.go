/*
Copyright 2021 The KodeRover Authors.

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

package service

import (
	"fmt"

	"github.com/koderover/zadig/pkg/microservice/aslan/config"
	commonmodels "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	commonservice "github.com/koderover/zadig/pkg/microservice/aslan/core/common/service"
	commontypes "github.com/koderover/zadig/pkg/microservice/aslan/core/common/types"
	"github.com/koderover/zadig/pkg/setting"
	internalresource "github.com/koderover/zadig/pkg/shared/kube/resource"
)

type ProductRevision struct {
	ID          string `json:"id,omitempty"`
	EnvName     string `json:"env_name"`
	ProductName string `json:"product_name"`
	// 表示该产品更新前版本
	CurrentRevision int64 `json:"current_revision"`
	// 表示该产品更新后版本
	NextRevision int64 `json:"next_revision"`
	// true: 表示该产品的服务发生变化, 需要更新
	// false: 表示该产品的服务未发生变化, 无需更新
	Updatable bool `json:"updatable"`
	// 可以自动更新产品, 展示用户更新前和更新后的服务组以及服务详细对比
	ServiceRevisions []*SvcRevision `json:"services"`
	IsPublic         bool           `json:"isPublic"`
}

type SvcRevision struct {
	ServiceName       string                          `json:"service_name"`
	Type              string                          `json:"type"`
	CurrentRevision   int64                           `json:"current_revision"`
	NextRevision      int64                           `json:"next_revision"`
	Updatable         bool                            `json:"updatable"`
	DeployStrategy    string                          `json:"deploy_strategy"`
	Error             string                          `json:"error"`
	Deleted           bool                            `json:"deleted"`
	New               bool                            `json:"new"`
	Containers        []*commonmodels.Container       `json:"containers,omitempty"`
	UpdateServiceTmpl bool                            `json:"update_service_tmpl"`
	VariableYaml      string                          `json:"variable_yaml"`
	VariableKVs       []*commontypes.RenderVariableKV `json:"variable_kvs"`
}

type ProductIngressInfo struct {
	IngressInfos []*commonservice.IngressInfo `json:"ingress_infos"`
	EnvName      string                       `json:"env_name"`
}

type SvcOptArgs struct {
	EnvName           string
	ProductName       string
	ServiceName       string
	ServiceType       string
	ServiceRev        *SvcRevision
	UpdateBy          string
	UpdateServiceTmpl bool
}

type PreviewServiceArgs struct {
	ProductName           string                          `json:"product_name"`
	EnvName               string                          `json:"env_name"`
	ServiceName           string                          `json:"service_name"`
	UpdateServiceRevision bool                            `json:"update_service_revision"`
	ServiceModules        []*commonmodels.Container       `json:"service_modules"`
	VariableKVs           []*commontypes.RenderVariableKV `json:"variable_kvs"`
}

type RestartScaleArgs struct {
	Type        string `json:"type"`
	ProductName string `json:"product_name"`
	EnvName     string `json:"env_name"`
	Name        string `json:"name"`
	// deprecated, since it is not used
	ServiceName string `json:"service_name"`
}

type ScaleArgs struct {
	Type        string `json:"type"`
	ProductName string `json:"product_name"`
	EnvName     string `json:"env_name"`
	ServiceName string `json:"service_name"`
	Name        string `json:"name"`
	Number      int    `json:"number"`
}

// SvcResp struct 产品-服务详情页面Response
type SvcResp struct {
	ServiceName string                       `json:"service_name"`
	Scales      []*internalresource.Workload `json:"scales"`
	Ingress     []*internalresource.Ingress  `json:"ingress"`
	Services    []*internalresource.Service  `json:"service_endpoints"`
	CronJobs    []*internalresource.CronJob  `json:"cron_jobs"`
	Namespace   string                       `json:"namespace"`
	EnvName     string                       `json:"env_name"`
	ProductName string                       `json:"product_name"`
	GroupName   string                       `json:"group_name"`
	Workloads   []*commonservice.Workload    `json:"-"`
}

func (pr *ProductRevision) GroupsUpdated() bool {
	if pr.ServiceRevisions == nil || len(pr.ServiceRevisions) == 0 {
		return false
	}
	for _, serviceRev := range pr.ServiceRevisions {
		if serviceRev.Updatable {
			return true
		}
	}
	return pr.Updatable
}

type ContainerNotFound struct {
	ServiceName string
	Container   string
	EnvName     string
	ProductName string
}

func (c *ContainerNotFound) Error() string {
	return fmt.Sprintf("serviceName:%s,container:%s", c.ServiceName, c.Container)
}

type NodeResp struct {
	Nodes  []*internalresource.Node `json:"data"`
	Labels []string                 `json:"labels"`
}

type ShareEnvReady struct {
	IsReady bool                `json:"is_ready"`
	Checks  ShareEnvReadyChecks `json:"checks"`
}

type ShareEnvReadyChecks struct {
	NamespaceHasIstioLabel  bool `json:"namespace_has_istio_label"`
	VirtualServicesDeployed bool `json:"virtualservice_deployed"`
	PodsHaveIstioProxy      bool `json:"pods_have_istio_proxy"`
	WorkloadsReady          bool `json:"workloads_ready"`
	WorkloadsHaveK8sService bool `json:"workloads_have_k8s_service"`
}

// Note: `WorkloadsHaveK8sService` is an optional condition.
func (s *ShareEnvReady) CheckAndSetReady(state ShareEnvOp) {
	if !s.Checks.WorkloadsReady {
		s.IsReady = false
		return
	}

	switch state {
	case ShareEnvEnable:
		if !s.Checks.NamespaceHasIstioLabel || !s.Checks.VirtualServicesDeployed || !s.Checks.PodsHaveIstioProxy {
			s.IsReady = false
		} else {
			s.IsReady = true
		}
	default:
		if !s.Checks.NamespaceHasIstioLabel && !s.Checks.VirtualServicesDeployed && !s.Checks.PodsHaveIstioProxy {
			s.IsReady = true
		} else {
			s.IsReady = false
		}
	}
}

type EnvoyClusterConfigLoadAssignment struct {
	ClusterName string             `json:"cluster_name"`
	Endpoints   []EnvoyLBEndpoints `json:"endpoints"`
}

type EnvoyLBEndpoints struct {
	LBEndpoints []EnvoyEndpoints `json:"lb_endpoints"`
}

type EnvoyEndpoints struct {
	Endpoint EnvoyEndpoint `json:"endpoint"`
}

type EnvoyEndpoint struct {
	Address EnvoyAddress `json:"address"`
}

type EnvoyAddress struct {
	SocketAddress EnvoySocketAddress `json:"socket_address"`
}

type EnvoySocketAddress struct {
	Protocol  string `json:"protocol"`
	Address   string `json:"address"`
	PortValue int    `json:"port_value"`
}

type ShareEnvOp string

const (
	ShareEnvEnable  ShareEnvOp = "enable"
	ShareEnvDisable ShareEnvOp = "disable"
)

type MatchedEnv struct {
	EnvName   string
	Namespace string
}

type OpenAPIScaleServiceReq struct {
	ProjectKey     string `json:"project_key"`
	EnvName        string `json:"env_name"`
	WorkloadName   string `json:"workload_name"`
	WorkloadType   string `json:"workload_type"`
	TargetReplicas int    `json:"target_replicas"`
}

func (req *OpenAPIScaleServiceReq) Validate() error {
	if req.ProjectKey == "" {
		return fmt.Errorf("project_key is required")
	}
	if req.EnvName == "" {
		return fmt.Errorf("env_name is required")
	}
	if req.WorkloadName == "" {
		return fmt.Errorf("workload_name is required")
	}
	if req.WorkloadType == "" {
		return fmt.Errorf("workload_type is required")
	}

	switch req.WorkloadType {
	case setting.Deployment, setting.StatefulSet:
	default:
		return fmt.Errorf("unsupported workload type: %s", req.WorkloadType)
	}

	if req.TargetReplicas < 0 {
		return fmt.Errorf("target_replicas must be greater than or equal to 0")
	}

	return nil
}

type OpenAPIApplyYamlServiceReq struct {
	EnvName     string               `json:"env_name"`
	ServiceList []*YamlServiceWithKV `json:"service_list"`
}

type YamlServiceWithKV struct {
	ServiceName string `json:"service_name"`
}

func (req *OpenAPIApplyYamlServiceReq) Validate() error {
	if req.EnvName == "" {
		return fmt.Errorf("env_name is required")
	}

	for _, serviceDef := range req.ServiceList {
		if serviceDef.ServiceName == "" {
			return fmt.Errorf("service_name is required for all services")
		}
	}
	return nil
}

type OpenAPIDeleteYamlServiceFromEnvReq struct {
	EnvName      string   `json:"env_name"`
	ServiceNames []string `json:"service_names"`
}

func (req *OpenAPIDeleteYamlServiceFromEnvReq) Validate() error {
	if req.EnvName == "" {
		return fmt.Errorf("env_name is required")
	}

	return nil
}

type OpenAPIEnvCfgArgs struct {
	Name             string                  `json:"name"`
	EnvName          string                  `json:"env_name"`
	ProductName      string                  `json:"product_name"`
	ServiceName      string                  `json:"service_name"`
	YamlData         string                  `json:"yaml_data"`
	CommonEnvCfgType config.CommonEnvCfgType `json:"common_env_cfg_type"`
	AutoSync         bool                    `json:"auto_sync"`
}

func (req *OpenAPIEnvCfgArgs) Validate() error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.EnvName == "" {
		return fmt.Errorf("env_name is required")
	}
	if req.ProductName == "" {
		return fmt.Errorf("project_name is required")
	}
	if req.CommonEnvCfgType == "" {
		return fmt.Errorf("common_env_cfg_type is required")
	}
	if req.YamlData == "" {
		return fmt.Errorf("yaml_data is required")
	}
	return nil
}
