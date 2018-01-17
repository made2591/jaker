package main

import (

	"fmt"
	"context"
	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"

)

func size() (result int64) {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		result += image.Size
	}

}

func main() {

	switch os.Args[1] {
		case "size":
			fmt.Println(size())
		case "clean":
			fmt.Println(clean())
		default:
			fmt.Println("No args passed")
	}

}