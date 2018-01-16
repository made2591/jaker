package lib

import (
	"github.com/0xAX/notificator"
	"net/http"
)

const (
	jakerName = "Jaker"
)

func Notify(title string, content string, ) http.Handler {

	var notify *notificator.Notificator

	notify = notificator.New(notificator.Options{
		DefaultIcon: "./img/docker.png",
		AppName:     jakerName,
	})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := notify.Push(title, content, "./img/docker.png", notificator.UR_NORMAL)
		if err != nil {
			w.WriteHeader(http.StatusOK)
		}
	})
}