// package echoserver uses websockets and is currently under construction.
package echoserver

import (
	"encoding/json"
	"fmt"
	//"github.com/fractalbach/ninjaServer/gamestate"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

// HandleWs is called by ninjaServer.go and converts all received
// messages into uppercase, and sends it back to the original source.
func HandleWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		// Process the message and replace it with a response.
		message = handleMessage(message)

		// Send the response back.
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Println(err)
			break
		}
	}
	log.Println(err)
}

// ~~~~~~~~~~~~~~~~~~ Game ~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// TODO: move these out of this file, and make a seperate file for the
// state of the game.  Variables related to the game.
var (
	nextPlayerId = 136

	// playerlist maps usernames to players.
	playerlist = map[string]*Player{}
)

func getNextPlayerId() int {
	nextPlayerId++
	return nextPlayerId
}

type IncomingMessage struct {
	ID     int
	Method string
	Params []interface{}
}

type ResultMessage struct {
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Kind    string      `json:"kind,omitempty"`
	Comment string      `json:"comment,omitempty"`
}

func handleMessage(data []byte) []byte {

	// Incoming Bytes -> Incoming Json
	j := new(IncomingMessage)
	err := json.Unmarshal(data, j)

	// TODO: Handle badly formatted json by sending back an error.
	if err != nil {
		return []byte{}
	}

	// Pass the Incoming Json to the Command Handler.  That
	// command handler returns a Result Json.
	result := handleCommand(j.Method, j.Params)

	// Add the message id number to the Result Json.  This allows
	// the client to identify their Json again.
	result.ID = j.ID

	// Convert Result Json back into bytes so that it can be sent
	// back to the client.
	outbytes, err := json.Marshal(result)
	if err != nil {
		log.Println(err)
	}
	return outbytes
}

// Where all the magic happens.  Accepts a command, does something,
// then returns a string as a response.  The boolean indicates an
// error.
func handleCommand(cmd string, params []interface{}) ResultMessage {
	out := ResultMessage{}
	switch cmd {
	case "hello":
		out.Result = "well hello to you too!"

	case "params":
		out.Comment = "type info about the given parameters in the form of (type, value)."
		s := make([]string, len(params))
		for i, _ := range params {
			q := params[i]
			s[i] = fmt.Sprintf("(%T, %v)", q, q)
		}
		out.Result = s

	case "add":
		return doAddCmd(params)

	case "list":
		out.Kind = "playerlist"
		out.Result = playerlist

	case "move":
		return doMoveCmd(params)

	case "update":
		for _, p := range playerlist {
			p.UpdatePosition()
		}
		out.Result = true

	default:
		out.Error = "command not found."
	}
	return out
}

// Adds a player to the playerlist.
func doAddCmd(params []interface{}) ResultMessage {
	out := ResultMessage{}

	// throw error if parameters number is wrong.
	if len(params) != 1 {
		out.Error = "expected 1 parameter."
		return out
	}

	// convert the first param to  type string.
	name, ok := params[0].(string)
	if !ok {
		out.Error = "Parameter Type error"
		return out
	}

	// add the player to the playerlist.
	playerlist[name] = &Player{
		PlayerId:   getNextPlayerId(),
		CurrentPos: Loc{5, 5},
		TargetPos:  Loc{5, 5},
	}
	out.Result = true
	return out
}

func doMoveCmd(params []interface{}) ResultMessage {
	out := ResultMessage{}

	// param length check.
	if len(params) != 3 {
		out.Error = "expected 3 params: (username, x, y)"
		return out
	}

	// convert types.
	name, ok1 := params[0].(string)
	x, ok2 := params[1].(float64)
	y, ok3 := params[2].(float64)
	if !(ok1 && ok2 && ok3) {
		out.Error = "type error: expected (string, int, int)"
		return out
	}

	// retrieve player pointer.
	p, ok4 := playerlist[name]
	if !ok4 {
		out.Error = "player does not exist."
		return out
	}

	// update positions and return successfull.
	p.TargetPos.X = int(x)
	p.TargetPos.Y = int(y)
	out.Result = true
	return out
}

// converts the playerlist into an array of usernames, friendly for
// displaying the list of added players.
func sprintPlayerList() []string {
	s := make([]string, len(playerlist))
	i := 0
	for key, _ := range playerlist {
		s[i] = key
		i++
	}
	return s
}

type Loc struct {
	X int
	Y int
}

type Player struct {
	PlayerId   int
	CurrentPos Loc
	TargetPos  Loc
}

func (p *Player) UpdatePosition() {
	next := p.CurrentPos

	if p.CurrentPos.X == p.TargetPos.X {
		goto Skip1
	}
	if p.CurrentPos.X < p.TargetPos.X {
		next.X = p.CurrentPos.X + 1
	} else {
		next.Y = p.CurrentPos.X - 1
	}
Skip1:
	if p.CurrentPos.Y == p.TargetPos.Y {
		goto Skip2
	}
	if p.CurrentPos.Y < p.TargetPos.Y {
		next.Y = p.CurrentPos.Y + 1
	} else {
		next.Y = p.CurrentPos.Y - 1
	}
Skip2:
	// Check for collisions at new x-position
	if NoCollisionAt(next.X, p.CurrentPos.Y) {
		p.CurrentPos.X = next.X
	}

	// Check for collisions at new y-position
	if NoCollisionAt(p.CurrentPos.X, next.X) {
		p.CurrentPos.Y = next.Y
	}
}

// TODO: add collision checking; currently there are no collisions.
//
// NoCollisionAt checks the tile at (x,y) to see if there is something
// that might prevent movement to that tile.
func NoCollisionAt(x, y int) bool {
	return true
}
