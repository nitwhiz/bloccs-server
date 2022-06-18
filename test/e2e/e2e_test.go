package e2e

import (
	"bloccs-server/internal/server"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// todo: finish this

var serverHost = "server:7000"

func createRoom(t *testing.T) string {
	roomsUrl := url.URL{
		Scheme: "http",
		Host:   serverHost,
		Path:   "/rooms",
	}

	resp, err := http.Post(roomsUrl.String(), "text/plain", nil)

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /rooms responded with status code %d, expected %d", resp.StatusCode, http.StatusCreated)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	var postRoomResponse struct {
		RoomId string `json:"roomId"`
	}

	err = json.Unmarshal(body, &postRoomResponse)

	if err != nil {
		t.Fatal(err)
	}

	roomId := postRoomResponse.RoomId

	if roomId == "" {
		t.Fatal("missing room id")
	}

	return roomId
}

func createSocket(t *testing.T, roomId string) *websocket.Conn {
	socketUrl := url.URL{
		Scheme: "ws",
		Host:   serverHost,
		Path:   fmt.Sprintf("/rooms/%s/socket", roomId),
	}

	client, _, err := websocket.DefaultDialer.Dial(socketUrl.String(), nil)

	if err != nil {
		t.Fatal(err)
	}

	return client
}

func handshakeRetrieveHello(t *testing.T, client *websocket.Conn) {
	msgType, msg, err := client.ReadMessage()

	if err != nil {
		t.Fatal(err)
	}

	if msgType != websocket.TextMessage {
		t.Fatal("message is not a text message")
	}

	var helloResponse EventResponse[interface{}, interface{}]

	err = json.Unmarshal(msg, &helloResponse)

	if err != nil {
		t.Fatal(err)
	}

	if helloResponse.Type != server.EventHello {
		t.Fatal("hello event malformed, wrong type")
	}

	if helloResponse.Source != nil {
		t.Fatal("hello event malformed, source is not nil")
	}

	if helloResponse.Payload != nil {
		t.Fatal("hello event malformed, payload is not nil")
	}

	t.Log("hello recieved")
}

func handshakeSendName(t *testing.T, client *websocket.Conn, name string) {
	err := client.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{ "name": "%s" }`, name)))

	if err != nil {
		t.Fatal(err)
	}

	t.Log("name sent")
}

func handshakeRetrieveHelloAckFirstClient(t *testing.T, client *websocket.Conn, expectedName string) (EventResponse[Room, HelloAckPayload], EventResponse[Room, PlayerJoinLeavePayload]) {
	msgType, msg, err := client.ReadMessage()

	if err != nil {
		t.Fatal(err)
	}

	if msgType != websocket.TextMessage {
		t.Fatal("message is not a text message")
	}

	var helloAckResponse EventResponse[Room, HelloAckPayload]

	err = json.Unmarshal(msg, &helloAckResponse)

	if err != nil {
		t.Fatal(err)
	}

	if helloAckResponse.Type != server.EventHelloAck {
		t.Fatal("hello event malformed, wrong type")
	}

	t.Log("hello ack recieved")

	if helloAckResponse.Payload.Player.Name != strings.ToUpper(expectedName) {
		t.Fatal("player name must be uppercased")
	}

	if len(helloAckResponse.Source.Players) != 0 {
		t.Fatal("not enough players in hello ack response source")
	}

	if len(helloAckResponse.Source.Games) != 0 {
		t.Fatal("not enough games in hello ack response source")
	}

	msgType, msg, err = client.ReadMessage()

	if err != nil {
		t.Fatal("unable to read message")
	}

	var playerJoinEvent EventResponse[Room, PlayerJoinLeavePayload]

	err = json.Unmarshal(msg, &playerJoinEvent)

	if err != nil {
		t.Fatal(err)
	}

	if playerJoinEvent.Type != server.EventPlayerJoin {
		t.Fatal("join event malformed, wrong type")
	}

	if playerJoinEvent.Source.Id != helloAckResponse.Source.Id {
		t.Fatal("join event malformed, wrong room id")
	}

	if playerJoinEvent.Payload.Player.Name != strings.ToUpper(expectedName) {
		t.Fatal("player name must be uppercased")
	}

	if len(playerJoinEvent.Source.Players) < 1 {
		t.Fatal("not enough players in hello ack response source")
	}

	if len(playerJoinEvent.Source.Games) < 1 {
		t.Fatal("not enough games in hello ack response source")
	}

	t.Log("join retrieved")

	return helloAckResponse, playerJoinEvent
}

func handshakeRetrieveHelloAckSecondClient(t *testing.T, client1 *websocket.Conn, client2 *websocket.Conn, expectedName1 string, expectedName2 string) (EventResponse[Room, HelloAckPayload], EventResponse[Room, PlayerJoinLeavePayload]) {
	msgType, msg, err := client2.ReadMessage()

	if err != nil {
		t.Fatal(err)
	}

	if msgType != websocket.TextMessage {
		t.Fatal("message is not a text message")
	}

	var helloAckResponse EventResponse[Room, HelloAckPayload]

	err = json.Unmarshal(msg, &helloAckResponse)

	if err != nil {
		t.Fatal(err)
	}

	if helloAckResponse.Type != server.EventHelloAck {
		t.Fatal("hello event malformed, wrong type")
	}

	if helloAckResponse.Payload.Player.Name != strings.ToUpper(expectedName2) {
		t.Fatal("player name must be uppercased")
	}

	if len(helloAckResponse.Source.Players) < 1 {
		t.Fatal("not enough players in hello ack response source")
	}

	if len(helloAckResponse.Source.Games) < 1 {
		t.Fatal("not enough games in hello ack response source")
	}

	t.Log("hello ack recieved")

	msgType, msg, err = client1.ReadMessage()

	if err != nil {
		t.Fatal("unable to read message")
	}

	var playerJoinEvent EventResponse[Room, PlayerJoinLeavePayload]

	err = json.Unmarshal(msg, &playerJoinEvent)

	if err != nil {
		t.Fatal(err)
	}

	if playerJoinEvent.Type != server.EventPlayerJoin {
		t.Fatal("join event malformed, wrong type")
	}

	if playerJoinEvent.Source.Id != helloAckResponse.Source.Id {
		t.Fatal("join event malformed, wrong room id")
	}

	if playerJoinEvent.Payload.Player.Name != strings.ToUpper(expectedName2) {
		t.Fatal("player name must be uppercased")
	}

	if len(playerJoinEvent.Source.Players) < 2 {
		t.Fatal("not enough players in hello ack response source")
	}

	player1Found := false
	player2Found := false

	for _, p := range playerJoinEvent.Source.Players {
		if p.Name == strings.ToUpper(expectedName1) {
			player1Found = true
		} else if p.Name == strings.ToUpper(expectedName2) {
			player2Found = true
		}
	}

	if player1Found == false {
		t.Fatalf("%s not found in event source", expectedName1)
	}

	if player2Found == false {
		t.Fatalf("%s not found in event source", expectedName2)
	}

	if len(playerJoinEvent.Source.Games) < 2 {
		t.Fatal("not enough games in hello ack response source")
	}

	t.Log("join retrieved")

	return helloAckResponse, playerJoinEvent
}

func Test_E2E(t *testing.T) {
	roomId := createRoom(t)

	client1 := createSocket(t, roomId)

	defer client1.Close()

	handshakeRetrieveHello(t, client1)
	handshakeSendName(t, client1, "andy")
	_, _ = handshakeRetrieveHelloAckFirstClient(t, client1, "andy")

	client2 := createSocket(t, roomId)

	defer client2.Close()

	handshakeRetrieveHello(t, client2)
	handshakeSendName(t, client2, "lisa")
	_, _ = handshakeRetrieveHelloAckSecondClient(t, client1, client2, "andy", "lisa")
}
