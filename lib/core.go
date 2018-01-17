package lib

import (
	"context"
	"net/http"
	"encoding/json"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/0xAX/notificator"
	"github.com/dustin/go-humanize"
	"fmt"
)

const (
	JAKER_NOTIFICATION_NAME     = "Jaker"
	JAKER_LIMIT_REACHED_TITLE   = "Docker image repository"
	JAKER_LIMIT_REACHED_CONTENT = "Limit reached: "
)

var (

	JONFIGURATION = Jonfiguration{ Alerts: []Jalert{} }

)

func NotifyLocalRepositorySize() http.Handler {

	var notify *notificator.Notificator

	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/docker.png",
		AppName:     JAKER_NOTIFICATION_NAME,
	})

	jmages := listImages()

	totalSize := int64(0)
	for _, jmage := range jmages {
		totalSize += jmage.Size
	}

	if totalSize > JONFIGURATION.LocalRepositorySize.Threshold {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := notify.Push(
				JAKER_LIMIT_REACHED_TITLE,
				JAKER_LIMIT_REACHED_CONTENT+
					humanize.Bytes(uint64(totalSize))+
					" with limit at "+
					humanize.Bytes(uint64(JONFIGURATION.LocalRepositorySize.Threshold)),
				"./img/docker.png",
				notificator.UR_NORMAL)
			if err != nil {
				w.WriteHeader(http.StatusOK)
			}
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		return
	})

}

// Serve configuration request - PUT
func Configuration() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&JONFIGURATION)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Print(err)
			return
		}

		b, err := json.Marshal(JONFIGURATION)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)

	})

}

// Get containers list from Docker API
func listContainers() (jontainers []Jontainer){

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		//fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		jontainers = append(jontainers, Jontainer{Id: container.ID[:10],
		Name: container.Names[0], Image: container.Image, Status: container.Status})
	}

	return jontainers

}

// Serve containers list - GET
func ListContainers() http.Handler {

	jontainers := listContainers()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(jontainers) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			b, err := json.Marshal(jontainers)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})

}

// Get images list from Docker API
func listImages() (jmages []Jmage) {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		jmages = append(jmages, Jmage{Id: image.ID[:10], Name: image.RepoDigests, Size: image.Size})
	}

	return jmages

}

// Serve images list - GET
func ListImages() http.Handler {

	jmages := listImages()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(jmages) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			b, err := json.Marshal(jmages)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(b)
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	})

}

// Get images list from Docker API
func getImagesSize() (Value) {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	result := int64(0)
	for _, image := range images {
		result += image.Size
	}

	return Value{ Value: result }

}

// Get images list from Docker API
func GetImagesSize() http.Handler {

	totalSize := getImagesSize()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(totalSize)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)
	})

}
