package cmd

import (
	"os"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	credential "github.com/aliyun/credentials-go/credentials"

	"github.com/coolestowl/ali-fc-webhook/server"
)

func Excute() {
	var (
		endpoint  = os.Getenv("ENDPOINT")
		region    = os.Getenv("REGION")
		accessKey = os.Getenv("ACCESS_KEY")
		secret    = os.Getenv("SECRET")
		jwtSecret = os.Getenv("JWT_SECRET")
	)

	for _, each := range []string{endpoint, region, accessKey, secret} {
		if len(each) == 0 {
			panic("not enough parameters !")
		}
	}

	cfg := new(openapi.Config)
	cfg.SetAccessKeyId(accessKey)
	cfg.SetAccessKeySecret(secret)
	cfg.SetRegionId(region)
	cfg.SetEndpoint(endpoint)
	cfg.SetCredential(func() credential.Credential {
		in := &credential.Config{}
		in.SetType("access_key")
		in.SetAccessKeyId(*cfg.AccessKeyId)
		in.SetAccessKeySecret(*cfg.AccessKeySecret)

		cred, err := credential.NewCredential(in)
		if err != nil {
			panic(err)
		}

		return cred
	}())

	if len(jwtSecret) > 0 {
		server.InitJwtSecret([]byte(jwtSecret))
	}

	cli, err := server.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	if err := cli.GinServer().Run(":8000"); err != nil {
		panic(err)
	}
}
