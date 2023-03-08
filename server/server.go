package server

import (
	"errors"
	"fmt"
	"net/http"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	fc "github.com/alibabacloud-go/fc-open-20210406/v2/client"
	"github.com/coolestowl/ali-fc-webhook/build"
	"github.com/gin-gonic/contrib/cors"
	"github.com/gin-gonic/gin"
)

type Client struct {
	sdk *fc.Client
}

func NewClient(cfg *openapi.Config) (*Client, error) {
	sdk, err := fc.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		sdk: sdk,
	}, nil
}

func (c *Client) GinServer(mountRoot string) *gin.Engine {
	e := gin.Default()

	e.Use(cors.Default())
	e.NoRoute(ErrFuncWrapper(func(c *gin.Context) (interface{}, error) {
		return c.Request.URL.Path, errors.New("not found")
	}))

	rootGroup := e.Group(mountRoot)
	rootGroup.GET("/version", ErrFuncWrapper(func(c *gin.Context) (interface{}, error) {
		return build.InfoMap, nil
	}))

	apiGroup := rootGroup.Group("/api", JWTAuth)
	{
		apiGroup.GET("/domains", ErrFuncWrapper(c.Domains))
		apiGroup.GET("/services", ErrFuncWrapper(c.Services))
		apiGroup.GET("/service/:service/functions", ErrFuncWrapper(c.Functions))
		apiGroup.GET("/service/:service/function/:function", ErrFuncWrapper(c.Get))
		apiGroup.POST("/service/:service/function/:function", ErrFuncWrapper(c.Apply))
	}

	triggerGroup := rootGroup.Group("/alitrigger", JWTAuth)
	{
		triggerGroup.POST("/service/:service/function/:function", ErrFuncWrapper(c.AliTriggerApply))
	}

	gin.SetMode(gin.ReleaseMode)
	return e
}

func (cli *Client) Domains(ctx *gin.Context) (interface{}, error) {
	out, err := cli.sdk.ListCustomDomains(&fc.ListCustomDomainsRequest{})
	if err != nil {
		return nil, err
	}

	if out.Body == nil || out.Body.CustomDomains == nil {
		return nil, fmt.Errorf("status: %d", *out.StatusCode)
	}

	type Route struct {
		Path   string
		Target string
	}

	type Domain struct {
		Name   string
		Routes []Route
	}

	domains := make([]Domain, 0, len(out.Body.CustomDomains))
	for _, dom := range out.Body.CustomDomains {
		d := Domain{
			Name: *dom.DomainName,
		}

		if dom.RouteConfig != nil {
			routes := make([]Route, 0, len(dom.RouteConfig.Routes))
			for _, r := range dom.RouteConfig.Routes {
				routes = append(routes, Route{
					Path:   *r.Path,
					Target: fmt.Sprintf("%s/%s:%s", *r.ServiceName, *r.FunctionName, *r.Qualifier),
				})
			}
			d.Routes = routes
		}

		domains = append(domains, d)
	}

	return domains, nil
}

func (cli *Client) Services(ctx *gin.Context) (interface{}, error) {
	out, err := cli.sdk.ListServices(&fc.ListServicesRequest{})
	if err != nil {
		return nil, err
	}

	if out.Body == nil || out.Body.Services == nil {
		return nil, fmt.Errorf("status: %d", *out.StatusCode)
	}

	type Service struct {
		ID   string
		Name string
	}

	services := make([]Service, 0, len(out.Body.Services))
	for _, svc := range out.Body.Services {
		services = append(services, Service{
			ID:   *svc.ServiceId,
			Name: *svc.ServiceName,
		})
	}

	return services, nil
}

func (cli *Client) Functions(ctx *gin.Context) (interface{}, error) {
	service := ctx.Param("service")

	out, err := cli.sdk.ListFunctions(&service, &fc.ListFunctionsRequest{})
	if err != nil {
		return nil, err
	}

	if out.Body == nil || out.Body.Functions == nil {
		return nil, fmt.Errorf("status: %d", *out.StatusCode)
	}

	type Function struct {
		ID     string
		Name   string
		Custom *CustomImage
	}

	functions := make([]Function, 0, len(out.Body.Functions))
	for _, svc := range out.Body.Functions {
		f := Function{
			ID:   *svc.FunctionId,
			Name: *svc.FunctionName,
		}

		if svc.CustomContainerConfig != nil && svc.CustomContainerConfig.Image != nil && svc.CustomContainerConfig.AccelerationType != nil {
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

	data, err := cli.sdk.GetFunction(&service, &function, &fc.GetFunctionRequest{})
	if err != nil {
		return nil, err
	}

	if data.Body == nil {
		return nil, fmt.Errorf("status: %d", *data.StatusCode)
	}

	protectSecret(data.Body.EnvironmentVariables, "ENDPOINT", "ACCESS_KEY", "SECRET")
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
		if _, err = cli.CreateService(service, req); err != nil {
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
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"code": 1,
				"msg":  err.Error(),
				"data": resp,
			})
			return
		}

		ctx.AbortWithStatusJSON(http.StatusOK, gin.H{
			"code": 0,
			"msg":  "success",
			"data": resp,
		})
	}
}

func protectSecret(mp map[string]*string, keys ...string) {
	if mp == nil {
		return
	}

	secret := "******"
	for _, key := range keys {
		if _, ok := mp[key]; ok {
			mp[key] = &secret
		}
	}
}
