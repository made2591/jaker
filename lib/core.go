package lib

import (
	"context"
	"net/http"
	"encoding/json"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types"
	"github.com/0xAX/notificator"
)

const (
	jakerName = "Jaker"
)

var (

	JCONFIGURATION = Jonfiguration{ Alerts: []Jalert{} }

)

func Notify() http.Handler {

	var notify *notificator.Notificator
	title   := "Docker image repository"
	content := "Limit reached: local repository size 9.8Gb"

	notify = notificator.New(notificator.Options{
		DefaultIcon: "icon/docker.png",
		AppName:     jakerName,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := notify.Push(title, content, "./img/docker.png", notificator.UR_NORMAL)
		if err != nil {
			w.WriteHeader(http.StatusOK)
		}
	})
}

// Serve configuration request - PUT
func Config() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}
		err := json.NewDecoder(r.Body).Decode(&JCONFIGURATION)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		b, err := json.Marshal(JCONFIGURATION)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(b)

	})

}

// Serve containers list - GET
func Listc() http.Handler {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	jontainers := []Jontainer{}
	for _, container := range containers {
		//fmt.Printf("%s %s\n", container.ID[:10], container.Image)
		jontainers = append(jontainers, Jontainer{Id: container.ID[:10], Name: container.Names[0], Image: container.Image, Status: container.Status})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(containers) == 0 {
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

func Listi() http.Handler {

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	images, err := cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	jmages := []Jmage{}
	for _, image := range images {
		jmages = append(jmages, Jmage{Id: image.ID[:10], Name: image.RepoDigests, Size: image.Size})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(images) == 0 {
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