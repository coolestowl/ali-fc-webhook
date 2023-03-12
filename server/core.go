package server

import (
	"encoding/json"
	"fmt"

	fc "github.com/alibabacloud-go/fc-open-20210406/v2/client"
)

func (cli *Client) CheckService(service string) (bool, error) {
	output, err := cli.sdk.ListServices(&fc.ListServicesRequest{})
	if err != nil {
		return false, err
	}

	if output.Body == nil || output.Body.Services == nil {
		return false, nil
	}

	for _, svc := range output.Body.Services {
		if svc.ServiceName != nil && *svc.ServiceName == service {
			return true, nil
		}
	}

	return false, nil
}

func (cli *Client) CheckServiceFunction(service, function string) (bool, error) {
	output, err := cli.sdk.ListFunctions(&service, &fc.ListFunctionsRequest{})
	if err != nil {
		return false, err
	}

	if output.Body == nil || output.Body.Functions == nil {
		return false, nil
	}

	for _, f := range output.Body.Functions {
		if f.FunctionName != nil && *f.FunctionName == function {
			return true, nil
		}
	}
	return false, nil
}

type FunctionReq struct {
	Service               *Service     `json:"service"`
	Custom                *CustomImage `json:"custom"`
	CreateWithHTTPTrigger bool         `json:"create_with_http_trigger"`
}

type Service struct {
	RoleARN string `json:"role"`
}

type CustomImage struct {
	Image        string `json:"image"`
	Acceleration string `json:"acceleration"`
}

func (cli *Client) CreateService(service string, req *FunctionReq) (interface{}, error) {
	in := &fc.CreateServiceRequest{}
	in.SetServiceName(service)
	if req.Service != nil {
		in.SetRole(req.Service.RoleARN)
	}

	return cli.sdk.CreateService(in)
}

func (cli *Client) CreateFunction(service, function string, req *FunctionReq) (interface{}, error) {
	in := &fc.CreateFunctionRequest{}
	in.SetFunctionName(function)
	in.SetInstanceConcurrency(100)
	in.SetInstanceSoftConcurrency(80)
	if req.Custom != nil {
		customImageConf := &fc.CustomContainerConfig{}
		customImageConf.SetImage(req.Custom.Image)

		in.SetRuntime("custom-container")
		in.SetHandler("index.handler")
		in.SetCustomContainerConfig(customImageConf)
	}

	resp, err := cli.sdk.CreateFunction(&service, in)
	if err != nil {
		return nil, err
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("status: %d", *resp.StatusCode)
	}

	if req.CreateWithHTTPTrigger {
		ctr := &fc.CreateTriggerRequest{}
		ctr.SetTriggerType("http")
		ctr.SetTriggerName("default-http")

		cfg := &fc.HTTPTriggerConfig{}
		cfg.SetAuthType("anonymous")
		cfg.SetMethods([]*string{toPtr("GET"), toPtr("POST")})
		cfg.SetDisableURLInternet(false)

		data, _ := json.Marshal(cfg)
		ctr.SetTriggerConfig(string(data))

		rsp, err := cli.sdk.CreateTrigger(&service, &function, ctr)
		if err != nil {
			return nil, err
		}
		if rsp.Body == nil {
			return nil, fmt.Errorf("status: %d", *rsp.StatusCode)
		}
	}

	protectSecret(resp.Body.EnvironmentVariables, "ENDPOINT", "ACCESS_KEY", "SECRET")
	return resp, nil
}

func (cli *Client) UpdateFunction(service, function string, req *FunctionReq) (interface{}, error) {
	customImageConf := &fc.CustomContainerConfig{}
	customImageConf.SetImage(req.Custom.Image)

	in := &fc.UpdateFunctionRequest{}
	in.SetCustomContainerConfig(customImageConf)

	resp, err := cli.sdk.UpdateFunction(&service, &function, in)
	if err != nil {
		return nil, err
	}

	if resp.Body == nil {
		return nil, fmt.Errorf("status: %d", *resp.StatusCode)
	}

	protectSecret(resp.Body.EnvironmentVariables, "ENDPOINT", "ACCESS_KEY", "SECRET")
	return resp, err
}

func toPtr[T any](x T) *T {
	return &x
}
