package main

import (
	"log"
	"net/http"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mkch/sync2/example/dinning/dinning"
)

func StartServer(addr string) error {
	http.Handle("/", http.FileServer(http.Dir("res")))
	http.HandleFunc("/ws", handleWs)
	go wsLoop()

	return http.ListenAndServe(addr, nil)
}

var wsConns []*websocket.Conn
var wsConnLock sync.RWMutex

type StickChange struct {
	Index int
	Stick dinning.StickState
}

var incoming = make(chan string)
var stickChanged = make(chan StickChange)
var dinningChanged = make(chan bool)
var newConn = make(chan *websocket.Conn)

var upgrader = websocket.Upgrader{}

func handleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	wsConnLock.Lock()
	wsConns = append(wsConns, conn)
	wsConnLock.Unlock()

	defer func() {
		wsConnLock.Lock()
		wsConns = slices.DeleteFunc(wsConns, func(c *websocket.Conn) bool { return c == conn })
		wsConnLock.Unlock()
	}()

	for {
		newConn <- conn
		var msg string
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println(err)
			return
		}
		incoming <- msg
	}

}

type msgSticks struct {
	Cmd    string
	Sticks [dinning.N]dinning.StickState
}

func newSticks(state *[dinning.N]dinning.StickState) msgSticks {
	return msgSticks{"Sticks", *state}
}

type msgDinningState struct {
	Cmd   string
	State string
}

func newDinningState(dinning bool) msgDinningState {
	var msg = "stopped"
	if dinning {
		msg = "started"
	}
	return msgDinningState{"DinningState", msg}
}

func wsLoop() {
	var sticks [dinning.N]dinning.StickState
	var sendTo = func(conn *websocket.Conn, msg any) {
		if err := conn.WriteJSON(msg); err != nil {
			log.Println(err)
		}
	}
	var send = func(msg any) {
		wsConnLock.RLock()
		defer wsConnLock.RUnlock()
		for _, conn := range wsConns {
			sendTo(conn, msg)
		}
	}

	for {
		select {
		case conn := <-newConn:
			// Send current state to new connection.
			sendTo(conn, newDinningState(dinning.Dinning()))
			sendTo(conn, newSticks(&sticks))
		case msg := <-incoming:
			log.Println(msg)
			switch msg {
			case "start_mutex":
				dinning.Mutex()
			case "start_mutexgroup":
				dinning.MutexGroup()
			case "stop":
				dinning.Stop()
			}
		case state := <-dinningChanged:
			send(newDinningState(state))
		case changed := <-stickChanged:
			sticks[changed.Index] = changed.Stick
			send(newSticks(&sticks))
		}
	}
}

func init() {
	dinning.ChangeStick = func(i int, stick dinning.StickState) {
		stickChanged <- StickChange{i, stick}
	}
	dinning.OnDinningChanged = func(dinning bool) {
		dinningChanged <- dinning
	}
}
