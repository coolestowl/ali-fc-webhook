package main

import (
	"os"

	"github.com/coolestowl/ali-fc-webhook/server"
)

func main() {
	var (
		endpoint  = os.Getenv("ENDPOINT")
		version   = os.Getenv("VERSION")
		accessKey = os.Getenv("ACCESS_KEY")
		secret    = os.Getenv("SECRET")
	)

	for _, each := range []string{endpoint, version, accessKey, secret} {
		if len(each) == 0 {
			panic("not enough parameters !")
		}
	}

	cli, err := server.NewClient(endpoint, version, accessKey, secret)
	if err != nil {
		panic(err)
	}

	if err := cli.GinServer().Run(":8000"); err != nil {
		panic(err)
	}
}
