package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pion/webrtc/v4"
	"golang.org/x/net/websocket"
)

const room_name = "chompnet"

type Announce_request struct {
	Numwant    int     `json:"numwant"`
	Uploaded   int64   `json:"uploaded"`
	Downloaded int64   `json:"downloaded"`
	Left       int64   `json:"left"`
	Event      string  `json:"event,omitempty"`
	Action     string  `json:"action"`
	Info_hash  string  `json:"info_hash"`
	Peer_id    string  `json:"peer_id"`
	Offers     []Offer `json:"offers"`
}

type Offer struct {
	Offer_id string                    `json:"offer_id"`
	Offer    webrtc.SessionDescription `json:"offer"`
}

type Announce_response struct {
	Info_hash  string                     `json:"info_hash"`
	Action     string                     `json:"action"`
	Interval   int                        `json:"interval,omitempty"`
	Complete   int                        `json:"complete,omitempty"`
	Incomplete int                        `json:"incomplete,omitempty"`
	Peer_id    string                     `json:"peer_id,omitempty"`
	To_peer_id string                     `json:"to_peer_id,omitempty"`
	Answer     *webrtc.SessionDescription `json:"answer,omitempty"`
	Offer      *webrtc.SessionDescription `json:"offer,omitempty"`
	Offer_id   string                     `json:"offer_id,omitempty"`
}

func main() {
	fmt.Printf("WebTorrent Signaling: Server\n")
	fmt.Printf("Room name: %s\n", room_name)

	var wt_id [20]byte
	_, err := rand.Read(wt_id[:20])
	if err != nil {
		log.Fatalf("Unable to generate WebTorrent ID: %v", err)
	}
	fmt.Printf("Peer ID: %s\n", wt_id)

	ws_origin := "http://localhost"
	ws_dest := "wss://tracker.openwebtorrent.com"
	ws, err := websocket.Dial(ws_dest, "", ws_origin)
	if err != nil {
		log.Fatal(err)
	}

	peer_connection, err := webrtc.NewAPI().NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{{
			URLs: []string{
				"stun:stun.l.google.com:19305",
				"stun:stun.l.google.com:19302",
			},
		},
		},
	})
	if err != nil {
		log.Fatalf("Unable to create PeerConnection: %v\n", err)
	}

	pc_offer, err := peer_connection.CreateOffer(nil)
	if err != nil {
		log.Fatalf("Unable to create WebRTC offer: %v\n", err)
	}
	offers := []Offer{{Offer: pc_offer, Offer_id: "alskdjlawid"}}

	var bytes_read int
	// torrent req
	torrent_req := &Announce_request{
		Numwant:    50,
		Uploaded:   0,
		Downloaded: 0,
		Left:       1,
		Action:     "announce",
		Info_hash:  room_name,
		Peer_id:    string(wt_id[:]),
		Event:      "started",
		Offers:     offers,
	}
	torrent_req_json, err := json.Marshal(torrent_req)
	if err != nil {
		log.Fatalf("Unable to send torrent_req: %v", err)
	}
	fmt.Printf("torrent_req: \n%s\n", &torrent_req_json)
	ws.Write(torrent_req_json)

	// tracker res
	var tracker_res_json [1500]byte
	var tracker_res Announce_response
	bytes_read, err = ws.Read(tracker_res_json[:])
	if err != nil {
		log.Fatalf("Unable to read from Websocket: %v", err)
	}

	fmt.Printf("tracker_res: \n%s\n", tracker_res_json)
	err = json.Unmarshal(tracker_res_json[:bytes_read], &tracker_res)
	if err != nil {
		log.Fatalf("Unable to unmarshal tracker_res: %v", err)
	}

	// server_answer
	var server_answer_json [1500]byte
	var server_answer Announce_response
	bytes_read, err = ws.Read(server_answer_json[:])
	if err != nil {
		log.Fatalf("Unable to read from Websocket: %v", err)
	}

	fmt.Printf("server_answer: \n%s\n", server_answer_json)
	err = json.Unmarshal(server_answer_json[:bytes_read], &server_answer)
	if err != nil {
		log.Fatalf("Unable to unmarshal server_answer: %v", err)
	}

	peer_connection.SetRemoteDescription(*server_answer.Answer)
}
