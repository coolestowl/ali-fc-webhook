package server

import (
	"fmt"
	"net/http"

	"github.com/aliyun/fc-go-sdk"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
)

type Client struct {
	sdk *fc.Client
}

func NewClient(endpoint, apiVersion, accessKeyID, accessKeySecret string, opts ...fc.ClientOption) (*Client, error) {
	sdk, err := fc.NewClient(endpoint, apiVersion, accessKeyID, accessKeySecret, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		sdk: sdk,
	}, nil
}

func (c *Client) GinServer(mountRoot string) *gin.Engine {
	e := gin.New()

	e.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"msg": "not found",
			"uri": c.Request.URL.Path,
		})
	})
	e.Use(cors.Default())

	rootGroup := e.Group(mountRoot)

	rootGroup.GET("/msg", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "auto update test",
		})
	})

	apiGroup := rootGroup.Group("/api")
	apiGroup.GET("/", ErrFuncWrapper(c.Services))
	apiGroup.GET("/:service", ErrFuncWrapper(c.Functions))
	apiGroup.GET("/:service/:function", ErrFuncWrapper(c.Get))
	apiGroup.POST("/:service/:function", ErrFuncWrapper(c.Apply))

	triggerGroup := rootGroup.Group("/alitrigger")
	triggerGroup.POST("/:service/:function", ErrFuncWrapper(c.AliTriggerApply))

	return e
}

func (cli *Client) Services(ctx *gin.Context) (interface{}, error) {
	out, err := cli.sdk.ListServices(fc.NewListServicesInput())
	if err != nil {
		return nil, err
	}

	type Service struct {
		ID   string
		Name string
	}

	services := make([]Service, 0, len(out.Services))
	for _, svc := range out.Services {
		services = append(services, Service{
			ID:   *svc.ServiceID,
			Name: *svc.ServiceName,
		})
	}

	return services, nil
}

func (cli *Client) Functions(ctx *gin.Context) (interface{}, error) {
	service := ctx.Param("service")

	out, err := cli.sdk.ListFunctions(fc.NewListFunctionsInput(service))
	if err != nil {
		return nil, err
	}

	type Function struct {
		ID     string
		Name   string
		Custom *CustomImage
	}

	functions := make([]Function, 0, len(out.Functions))
	for _, svc := range out.Functions {
		f := Function{
			ID:   *svc.FunctionID,
			Name: *svc.FunctionName,
		}

		if svc.CustomContainerConfig != nil {
			f.Custom = &CustomImage{
				Image:        *svc.CustomContainerConfig.Image,
				Acceleration: *svc.CustomContainerConfig.AccelerationType,
			}
		}

		functions = append(functions, f)
	}

	return functions, nil
}

func (cli *Client) Get(ctx *gin.Context) (interface{}, error) {
	var (
		service  = ctx.Param("service")
		function = ctx.Param("function")
	)

	ok, err := cli.CheckService(service)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("service: %v", service)
	}

	ok, err = cli.CheckServiceFunction(service, function)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, fmt.Errorf("function: %v", service)
	}

	data, err := cli.sdk.GetFunction(fc.NewGetFunctionInput(service, function))
	if err != nil {
		return nil, err
	}

	protectSecret(data.EnvironmentVariables, "ENDPOINT", "ACCESS_KEY", "SECRET")
	return data, nil
}

func (cli *Client) Apply(ctx *gin.Context) (interface{}, error) {
	var (
		service  = ctx.Param("service")
		function = ctx.Param("function")
		req      = FunctionReq{}
	)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	return cli.apply(service, function, &req)
}

func (cli *Client) apply(service, function string, req *FunctionReq) (interface{}, error) {
	ok, err := cli.CheckService(service)
	if err != nil {
		return nil, err
	}

	if !ok {
		in := fc.NewCreateServiceInput()
		in.WithServiceName(service)

		if _, err = cli.sdk.CreateService(in); err != nil {
			return nil, err
		}
	}

	ok, err = cli.CheckServiceFunction(service, function)
	if err != nil {
		return nil, err
	}

	f := cli.UpdateFunction
	if !ok {
		f = cli.CreateFunction
	}

	return f(service, function, req)
}

func ErrFuncWrapper(f func(*gin.Context) (interface{}, error)) func(*gin.Context) {
	return func(ctx *gin.Context) {
		resp, err := f(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    1,
				"message": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"code": 0,
			"data": resp,
		})
	}
}

func protectSecret(mp map[string]string, keys ...string) {
	for _, key := range keys {
		if _, ok := mp[key]; ok {
			mp[key] = "******"
		}
	}
}
