package server

import "github.com/aliyun/fc-go-sdk"

func (cli *Client) CheckService(service string) (bool, error) {
	output, err := cli.sdk.ListServices(fc.NewListServicesInput())
	if err != nil {
		return false, err
	}

	for _, svc := range output.Services {
		if svc.ServiceName != nil && *svc.ServiceName == service {
			return true, nil
		}
	}
	return false, nil
}

func (cli *Client) CheckServiceFunction(service, function string) (bool, error) {
	output, err := cli.sdk.ListFunctions(fc.NewListFunctionsInput(service))
	if err != nil {
		return false, err
	}

	for _, f := range output.Functions {
		if f.FunctionName != nil && *f.FunctionName == function {
			return true, nil
		}
	}
	return false, nil
}

type FunctionReq struct {
	Custom *CustomImage `json:"custom"`
}

type CustomImage struct {
	Image        string `json:"image"`
	Acceleration string `json:"acceleration"`
}

func (cli *Client) CreateFunction(service, function string, req *FunctionReq) (interface{}, error) {
	in := fc.NewCreateFunctionInput(service)
	in.WithFunctionName(function)

	if req.Custom != nil {
		customImageConf := fc.NewCustomContainerConfig().
			WithImage(req.Custom.Image).
			WithAccelerationType("None")

		if req.Custom.Acceleration == "Default" {
			customImageConf.WithAccelerationType("Default")
		}

		in.WithCustomContainerConfig(customImageConf)
	}

	resp, err := cli.sdk.CreateFunction(in)

	protectSecret(resp.EnvironmentVariables, "ENDPOINT", "ACCESS_KEY", "SECRET")
	return resp, err
}

func (cli *Client) UpdateFunction(service, function string, req *FunctionReq) (interface{}, error) {
	in := fc.NewUpdateFunctionInput(service, function)

	if req.Custom != nil {
		customImageConf := fc.NewCustomContainerConfig().
			WithImage(req.Custom.Image).
			WithAccelerationType("None")

		if req.Custom.Acceleration == "Default" {
			customImageConf.WithAccelerationType("Default")
		}

		in.WithCustomContainerConfig(customImageConf)
	}

	resp, err := cli.sdk.UpdateFunction(in)

	protectSecret(resp.EnvironmentVariables, "ENDPOINT", "ACCESS_KEY", "SECRET")
	return resp, err
}
