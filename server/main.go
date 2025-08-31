package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
)

var id = 0

type Client struct {
	ID          int
	Username    string
	Connection  net.Conn
	CurrentRoom string
}

var clients = make(map[int]*Client)
var roomIDs = make(map[string]bool)

func (c *Client) NotInRoom() bool {
	return c.CurrentRoom == ""
}

func (c *Client) SetUsername(command string) {
	username := strings.TrimPrefix(command, "set username ")
	if username == "" {
		c.Connection.Write([]byte("Failed to change username!\n"))
		return
	}
	c.Username = username
	c.Connection.Write([]byte("Username changed!\n"))
}

func (c *Client) CreateRoom(command string) {
	roomId := strings.TrimPrefix(command, "create-room ")
	if roomId == "" {
		c.Connection.Write([]byte("Failed to create room!\n"))
		return
	}
	if roomIDs[roomId] {
		c.Connection.Write([]byte("This Room already exists!\n"))
		return
	}
	roomIDs[roomId] = true
	c.CurrentRoom = roomId
	c.Connection.Write([]byte("Room created!\n"))
}

func (c *Client) JoinRoom(command string) {
	roomId := strings.TrimPrefix(command, "join-room ")
	if !roomIDs[roomId] {
		c.Connection.Write([]byte("Room not found!\n"))
		return
	}
	c.CurrentRoom = roomId
	roomIDs[roomId] = true
	c.Connection.Write([]byte("Room joined!\n"))
}

func (c *Client) LeaveRoom() {
	c.CurrentRoom = ""
	c.Connection.Write([]byte("You left the room!\n"))
}

func getNextID() int {
	id++
	return id
}

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	log.Println("TCP server listening on 8080")

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection ", err.Error())
			continue
		}
		clientID := getNextID()
		newClient := &Client{
			ID:          clientID,
			Connection:  connection,
			CurrentRoom: "",
			Username:    strconv.Itoa(clientID),
		}
		clients[newClient.ID] = newClient
		go chat(newClient)
	}
}

func chat(client *Client) {
	defer client.Connection.Close()
	scanner := bufio.NewScanner(client.Connection)
	for scanner.Scan() {
		msg := scanner.Text()
		if client.NotInRoom() {
			if msg == "quit" {
				client.Connection.Write([]byte("Bye!\n"))
				break
			} else if strings.HasPrefix(msg, "set username ") {
				client.SetUsername(msg)
			} else if strings.HasPrefix(msg, "create-room ") {
				client.CreateRoom(msg)
			} else if strings.HasPrefix(msg, "join-room ") {
				client.JoinRoom(msg)
			}
		} else {
			switch msg {
			case "!leave-room":
				client.LeaveRoom()
			default:
				//send message to all connected clients
				for _, c := range clients {
					if c.ID != client.ID && c.CurrentRoom == client.CurrentRoom {
						c.Connection.Write([]byte(msg + "\n"))
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Println("Read Error: ", err.Error())
	}
}
