package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"golang.org/x/net/context"
	"os"
)

func testBuildImage() {
	log.SetLevel(log.DebugLevel)
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.12", nil, defaultHeaders)
	if err != nil {
		panic(err)
	}

	buildContext, err := os.Open("Dockerfile.tar")
	defer buildContext.Close()
	response, err := cli.ImageBuild(context.Background(), buildContext, types.ImageBuildOptions{
		Tags:       []string{"lol"},
		Dockerfile: "Dockerfile",
	})
	if err != nil {
		log.Errorf("Error, got %v", err)
	}
	log.Debug(response.OSType)

}

func main() {
	testBuildImage()
}
