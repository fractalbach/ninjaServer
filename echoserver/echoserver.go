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

	// todo: use gamestate instead.
	// currently, this is not concurrency safe.
	playerlist = map[string]bool{}
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

type IncomingMessage struct {
	ID     int
	Method string
	Params []interface{}
}

type ResultMessage struct {
	ID     int         `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
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
	case "myparams":
		out.Result = fmt.Sprint(params)
	case "add":
		return doAddCmd(params)		
	case "list":
		out.Result = sprintPlayerList()
	default:
		out.Error = "command not found."
	}
	return out
}

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
	playerlist[name] = true
	out.Result = true
	return out
}

func sprintPlayerList() []string {
	s := make([]string, len(playerlist))
	i := 0
	for key, _ := range(playerlist) {
		s[i] = key
		i++
	}
	return s
}
