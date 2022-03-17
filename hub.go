// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

// boardSize is the width and height of the board, matching the table
// size in home.html
const boardSize = 100

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// board is an in-memory representation of the board
	// where each entry is a javascript color
	board [][]string
}

func newHub() *Hub {
	hub := &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		board:      make([][]string, boardSize),
	}

	// initialize board
	for i := 0; i < len(hub.board); i++ {
		hub.board[i] = make([]string, boardSize)
		for j := 0; j < boardSize; j++ {
			hub.board[i][j] = "cornflowerblue"
		}
	}

	return hub
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			// parse and set color in memory
			if err := h.parseAndSave(message); err != nil {
				log.Println(err)
			}
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// parseAndSave parses a message into x, y, and color and saves it to
// the board
func (h *Hub) parseAndSave(message []byte) error {
	s := string(message)
	parts := strings.Split(s, " ")

	if len(parts) < 3 {
		// do nothing if we don't have enough information from the message
		return errors.New("malformed message")
	}

	x, y, color := parts[0], parts[1], parts[2]
	var xPos, yPos int
	var err error

	// convert to integers
	if xPos, err = strconv.Atoi(x); err != nil {
		return err
	}
	if yPos, err = strconv.Atoi(y); err != nil {
		return err
	}

	// check bounds
	if yPos < 0 || xPos < 0 || yPos >= len(h.board) || xPos >= len(h.board[yPos]) {
		return errors.New("out of bounds")
	}

	h.board[yPos][xPos] = color
	return nil
}
