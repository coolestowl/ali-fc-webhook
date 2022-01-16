package server

import "github.com/gin-gonic/gin"

type PushData struct {
	Digest   string `json:"digest"`
	PushedAt string `json:"pushed_at"`
	Tag      string `json:"tag"`
}

type Repo struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	CreatedAt string `json:"date_created"`

	Region         string `json:"region"`
	RepoAuthType   string `json:"repo_authentication_type"`
	RepoFullName   string `json:"repo_full_name"`
	RepoOriginType string `json:"repo_origin_type"`
	RepoTyoe       string `json:"repo_type"`
}

func (cli *Client) AliTriggerApply(ctx *gin.Context) (interface{}, error) {
	var (
		service  = ctx.Param("service")
		function = ctx.Param("function")
		req      = struct {
			PushData PushData `json:"push_data"`
			Repo     Repo     `json:"repository"`
		}{}
	)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return nil, err
	}

	return cli.apply(service, function, &FunctionReq{
		Custom: &CustomImage{
			Image:        req.Repo.Name + req.PushData.Tag,
			Acceleration: "Default",
		},
	})
}
