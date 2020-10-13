package server

import (
	"OnlineXO/config"
	"OnlineXO/internal/app"
	"OnlineXO/internal/template"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var server = app.NewServer()

func WebSocket(context echo.Context) error {
	token := context.Param("token")
	conn, err := upgrader.Upgrade(context.Response(), context.Request(), nil)
	if err != nil {
		log.Print(err)
	} else {
		c := server.GetClientByToken(token)
		if c == nil {
			c = app.NewClient(conn, server)
			server.AddClient(c)
		} else {
			c.SetConnection(conn)
			c.Reconnect()
		}
	}
	return context.HTML(http.StatusOK, "done")
}

func Home(context echo.Context) error {
	return context.Render(http.StatusOK, "base.html", template.TemplateData(nil))
}

func Join(context echo.Context) error {
	room_id := context.Param("uid")
	room := server.GetRoom(room_id)
	if room == nil {
		return context.Redirect(http.StatusFound, config.UrlFor("/xo"))
	}
	return context.Render(http.StatusOK, "base.html", template.TemplateData(echo.Map{
		"room_to_join": room_id,
	}))
}
