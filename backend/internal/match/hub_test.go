package match_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/encertia/backend/internal/match"
	"github.com/google/uuid"
)

func TestHub_RoomLifecycle(t *testing.T) {
	hub := match.NewHub()
	pin := "123456"
	matchID := uuid.New()
	hostID := uuid.New()

	room := hub.GetOrCreateRoom(pin, matchID, hostID)
	if room == nil {
		t.Fatalf("esperava una room creada, ha retornat nil")
	}

	if room.PIN != pin || room.MatchID != matchID || room.HostID != hostID {
		t.Errorf("dades de room incorrectes: got %+v", room)
	}

	// Retrieve existing
	room2 := hub.GetRoom(pin)
	if room2 != room {
		t.Errorf("GetRoom no ha retornat la mateixa instància")
	}

	// Delete room
	hub.RemoveRoom(pin)
	if hub.GetRoom(pin) != nil {
		t.Errorf("la room no s'ha eliminat correctament")
	}
}

func TestRoom_ClientRegistrationAndMessaging(t *testing.T) {
	hub := match.NewHub()
	pin := "654321"
	matchID := uuid.New()
	hostID := uuid.New()
	playerID := uuid.New()
	userID := uuid.New()

	room := hub.GetOrCreateRoom(pin, matchID, hostID)

	// Host Client
	hostClient := &match.Client{
		Hub:      hub,
		Send:     make(chan []byte, 10),
		UserID:   hostID,
		Nickname: "Host",
		IsHost:   true,
	}
	room.RegisterClient(hostClient)

	// Player Client
	playerClient := &match.Client{
		Hub:      hub,
		Send:     make(chan []byte, 10),
		UserID:   userID,
		PlayerID: &playerID,
		Nickname: "Alumne1",
		IsHost:   false,
	}
	room.RegisterClient(playerClient)

	if count := room.ConnectedPlayerCount(); count != 1 {
		t.Errorf("esperava 1 jugador connectat, obtingut %d", count)
	}

	if found := room.GetPlayerClient(playerID); found != playerClient {
		t.Errorf("GetPlayerClient no ha trobat el client correcte")
	}

	if found := room.GetPlayerClientByUserID(userID); found != playerClient {
		t.Errorf("GetPlayerClientByUserID no ha trobat el client correcte")
	}

	// Broadcast message
	testMsg := match.OutgoingWSMessage{
		Event: "test:broadcast",
		Data:  map[string]string{"foo": "bar"},
	}
	room.Broadcast(testMsg)

	// Check host received
	select {
	case data := <-hostClient.Send:
		var parsed match.OutgoingWSMessage
		_ = json.Unmarshal(data, &parsed)
		if parsed.Event != "test:broadcast" {
			t.Errorf("Host no ha rebut l'esdeveniment esperat: %s", parsed.Event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout esperant missatge al Host")
	}

	// Check player received
	select {
	case data := <-playerClient.Send:
		var parsed match.OutgoingWSMessage
		_ = json.Unmarshal(data, &parsed)
		if parsed.Event != "test:broadcast" {
			t.Errorf("Player no ha rebut l'esdeveniment esperat: %s", parsed.Event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout esperant missatge al Player")
	}

	// Send to Host only
	room.SendToHost(match.OutgoingWSMessage{Event: "test:host_only"})
	select {
	case data := <-hostClient.Send:
		var parsed match.OutgoingWSMessage
		_ = json.Unmarshal(data, &parsed)
		if parsed.Event != "test:host_only" {
			t.Errorf("esperava test:host_only, got %s", parsed.Event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout esperant missatge al Host")
	}

	// Send to Player only
	room.SendToPlayer(playerID, match.OutgoingWSMessage{Event: "test:player_only"})
	select {
	case data := <-playerClient.Send:
		var parsed match.OutgoingWSMessage
		_ = json.Unmarshal(data, &parsed)
		if parsed.Event != "test:player_only" {
			t.Errorf("esperava test:player_only, got %s", parsed.Event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout esperant missatge al Player")
	}

	// Unregister Player
	room.UnregisterClient(playerClient)
	if count := room.ConnectedPlayerCount(); count != 0 {
		t.Errorf("esperava 0 jugadors connectats després de baixa, obtingut %d", count)
	}

	playerClient.Close()
	hostClient.Close()
}
