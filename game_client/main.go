package main

import (
	"encoding/json"
	"log"

	"github.com/streamerd/actors-in-game/types"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const wsServerEndpoint = "ws://localhost:40000/ws"
const ONE_KB = 1024

func (c *GameClient) login() error {
	b, err := json.Marshal(types.Login{
		ClientID: c.clientID,
		Username: c.username,
	})

	if err != nil {
		return err
	}
	msg := types.WSMessage{
		Type: "Login",
		Data: b,
	}
	return c.conn.WriteJSON(msg)
}

func (ps *PlayerState) updatePosition(x int, y int) error {
	b, err := json.Marshal(types.PlayerState{
		Health:   ps.Health,
		Position: types.Position{X: x, Y: y},
	})

	if err != nil {
		return err
	}
	msg := types.WSMessage{
		Type: "PlayerState",
		Data: b,
	}
	return ps.conn.WriteJSON(msg)
}

type GameClient struct {
	clientID string
	username string
	conn     *websocket.Conn
}

func newGameClient(conn *websocket.Conn, username string) *GameClient {
	return &GameClient{
		clientID: uuid.New().String(),
		username: username,
		conn:     conn,
	}
}

func main() {
	dialer := websocket.Dialer{
		ReadBufferSize:  ONE_KB,
		WriteBufferSize: ONE_KB,
	}

	conn, _, err := dialer.Dial(wsServerEndpoint, nil)
	if err != nil {
		log.Fatal(err)
	}

	c := newGameClient(conn, "UserX")
	if err := c.login(); err != nil {
		log.Fatal(err)
	}

	// Send messages of type PlayerState to GameServer with random Position values.

	// for {
	// 	x := rand.Intn(1000)
	// 	y := rand.Intn(1000)
	// 	state := types.PlayerState{
	// 		Health:   100,
	// 		Position: types.Position{X: x, Y: y},
	// 	}
	// 	b, err := json.Marshal(state)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	msg := types.WSMessage{
	// 		Type: "PlayerState",
	// 		Data: b,
	// 	}
	// 	if err := conn.WriteJSON(msg); err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	time.Sleep(time.Millisecond * 100)
	// }

}
