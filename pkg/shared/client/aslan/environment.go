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

package aslan

import (
	"fmt"

	"github.com/koderover/zadig/pkg/tool/httpclient"
	"github.com/koderover/zadig/pkg/types"
)

func (c *Client) ListEnvironments(projectName string) ([]*Environment, error) {
	url := "/environment/environments"

	res := make([]*Environment, 0)
	_, err := c.Get(url, httpclient.SetQueryParam("projectName", projectName), httpclient.SetResult(&res))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) GetEnvironment(envName, projectName string) (*Environment, error) {
	url := fmt.Sprintf("/environment/environments/%s", envName)

	res := &Environment{}
	_, err := c.Get(url, httpclient.SetQueryParam("projectName", projectName), httpclient.SetResult(res))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) GetServiceDetail(projectName, serviceName, envName string) (*ServiceDetail, error) {
	url := fmt.Sprintf("/environment/environments/%s/services/%s", envName, serviceName)

	res := &ServiceDetail{}
	req := map[string]string{
		"projectName": projectName,
	}
	_, err := c.Get(url, httpclient.SetQueryParams(req), httpclient.SetResult(res))
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PatchWorkload(projectName, envName, serviceName, devImage string) (*types.WorkloadInfo, error) {
	res := &types.WorkloadInfo{}

	body := types.StartDevmodeInfo{
		DevImage: devImage,
	}
	url := fmt.Sprintf("/environment/environments/%s/services/%s/devmode/patch?projectName=%s", envName, serviceName, projectName)
	_, err := c.Post(url, httpclient.SetBody(body), httpclient.SetResult(res))

	return res, err
}

func (c *Client) RecoverWorkload(projectName, envName, serviceName string) error {
	url := fmt.Sprintf("/environment/environments/%s/services/%s/devmode/recover?projectName=%s", envName, serviceName, projectName)
	_, err := c.Post(url)

	return err
}
