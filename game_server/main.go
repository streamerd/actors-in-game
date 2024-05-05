package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/streamerd/actors-in-game/types"

	"github.com/anthdm/hollywood/actor"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// actor producer is a function that produce a receiver.
const ONE_KB int = 1024

type GameServer struct {
	ctx      *actor.Context
	sessions map[*actor.PID]struct{}
}

func newGameServer() actor.Receiver {
	return &GameServer{
		sessions: make(map[*actor.PID]struct{}),
	}
}

type PlayerState struct {
	Position types.Position
	Health   int
}

type PlayerSession struct {
	clientID  string
	sessionID string
	username  string
	conn      *websocket.Conn
}

func newPlayerSession(sid string, conn *websocket.Conn) actor.Producer {
	return func() actor.Receiver {
		return &PlayerSession{
			sessionID: sid,
			conn:      conn,
		}
	}
}

func (s *PlayerSession) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		s.readloop()
		_ = msg
		// statePid := c.SpawnChild(newPlayerState, "PlayerState")
	}
}

func (s *PlayerSession) readloop() {
	var msg types.WSMessage

	for {
		if err := s.conn.ReadJSON(&msg); err != nil {
			fmt.Println("read error ", err)
			return
		}
		go s.handleMessage(msg)
	}
}

func (s *PlayerSession) handleMessage(msg types.WSMessage) {

	switch msg.Type {
	case "Login":
		var loginMsg types.Login
		if err := json.Unmarshal(msg.Data, &loginMsg); err != nil {
			panic(err)
		}
		s.clientID = loginMsg.ClientID
		s.username = loginMsg.Username
		fmt.Printf("%s logged in with sessionID: %s \n", loginMsg.Username, s.sessionID)

	case "PlayerState":
		var ps types.PlayerState
		if err := json.Unmarshal(msg.Data, &ps); err != nil {
			panic(err)
		}
		fmt.Printf("SessionID : %s - Position updated: %v\n", s.sessionID, ps.Position)

	}
}

func (s *GameServer) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case actor.Started:
		s.startHTTP()
		s.ctx = c
		_ = msg
	}
}

func (s *GameServer) startHTTP() {
	log.Println("Starting http server on port 40000")
	go func() {
		http.HandleFunc("/ws", s.handleWS)
		http.ListenAndServe(":40000", nil)
	}()
}

// hndles the upgrate of the websocket
func (s *GameServer) handleWS(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		ReadBufferSize:  ONE_KB,
		WriteBufferSize: ONE_KB,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("ws upgrade error: ", err)
		return
	}
	sid := uuid.NewString()
	fmt.Printf("session:%s just connected \n", sid)
	pid := s.ctx.SpawnChild(newPlayerSession(sid, conn), fmt.Sprintf("session:%s", sid))
	s.sessions[pid] = struct{}{}
	fmt.Printf("%s just exited \n", sid)

}

func main() {
	e, _ := actor.NewEngine(actor.EngineConfig{})
	e.Spawn(newGameServer, "server")
	select {}
}
