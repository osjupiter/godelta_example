package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"log"
)

var (
	upgrader = websocket.Upgrader{}
	newbie   = make(chan *websocket.Conn, 10)
	clients  = make([]*websocket.Conn, 10)
	receives = make(chan data, 10)
)

type data struct {
	message string
	id      int
}

func wsWatch(id int, receiveChan chan data, ws *websocket.Conn) {
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("%s", err)
		}
		fmt.Printf("%s\n", msg)
		receiveChan <- data{string(msg), id}
	}
}
func wsSend(receive data, ws *websocket.Conn) {
	err := ws.WriteMessage(websocket.TextMessage, []byte(receive.message))
	if err != nil {
		log.Printf("%s", err)
	}
}

func daemon() {
	for {
		//入室をリストに
		select {
		case v, ok := <-newbie:
			if ok {

				clients = append(clients, v)
				go wsWatch(len(clients)-1, receives, v)
			} else {
				log.Print("something happen")
			}
		case v, _ := <-receives:
			for i, ws := range clients {
				if i == v.id {
					continue
				}
				if ws==nil{
					continue
				}
				wsSend(v, ws)
			}
		}

	}
}

func hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	newbie <- ws
	return nil

}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/ws", hello)
	e.Static("/", "static")

	go daemon()

	e.Logger.Fatal(e.Start(":1323"))
}
