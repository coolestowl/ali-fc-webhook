package server

import (
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

	apiGroup := rootGroup.Group("/api/function")
	apiGroup.POST("/update", c.UpdateFunction)

	return e
}

type CreateFunctionReq struct {
	Service  string       `json:"service"`
	Function string       `json:"function"`
	Custom   *CustomImage `json:"custom"`
}

func (cli *Client) CreateFunction(ctx *gin.Context) {
	ErrFuncHandler(ctx, func(c *gin.Context) (interface{}, error) {
		req := &CreateFunctionReq{}

		if err := c.ShouldBindJSON(req); err != nil {
			return nil, err
		}

		in := fc.NewCreateFunctionInput(req.Function)
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
		if err != nil {
			return nil, err
		}

		return resp, nil
	})
}

type UpdateFunctionReq struct {
	Service  string       `json:"service"`
	Function string       `json:"function"`
	Custom   *CustomImage `json:"custom"`
}

type CustomImage struct {
	Image        string `json:"image"`
	Acceleration string `json:"acceleration"`
}

func (cli *Client) UpdateFunction(ctx *gin.Context) {
	ErrFuncHandler(ctx, func(c *gin.Context) (interface{}, error) {
		req := &UpdateFunctionReq{}

		if err := ctx.ShouldBindJSON(req); err != nil {
			return nil, err
		}

		in := fc.NewUpdateFunctionInput(req.Service, req.Function)
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
		if err != nil {
			return nil, err
		}

		return resp, nil
	})
}
