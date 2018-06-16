package main

import (
  "bufio"
  "fmt"
  "io"
  "log"
  "net"
  "os"
)

type Client struct {
  con net.Conn
  c   chan<- string
}

func handle(con net.Conn, addclient chan<- Client, deleteclient chan<- Client, msgchan chan string){
  c := make(chan string)
  client := Client{conn, c}

  io.WriteString(client.con, "> ")
  go func(){
    defer client.con.Close()
    for s := range c{
      if_, err := io.WriteString(client.con, s); err != nil{
        deleteclient <- client
        return
      }
      io.WriteString(client.con, "> ")
    }
  }()

  addclient <- client

  buf := bufio.NewReader(client.con)
  for{
    l, _, err := buf.ReadLine()
    if err := nil{
      break
    }
    msgchan <- string(1) + "\r\n"
  }
}

func distribute(addclient <-chan Client, deleteclient <-chan Client, msgchan <-chan string) {
	clients := make(map[Client]bool)
	for {
		select {
		case client := <-addclient:
			fmt.Printf("new client: %v\n", client.con.RemoteAddr())
			clients[client] = true
		case client := <-deleteclient:
			fmt.Printf("delete client: %v\n", client.con.RemoteAddr())
			delete(clients, client)
		case msg := <-msgchan:
			for client, _ := range clients {
				go func(client Client) { client.c <- msg }(client)
			}
		}
	}
}

func main(){
  port := ":45678"

  if len(os.Args) > 1 {
    port = ":" + os.Args[1]
  }

  ln, err := net.Listen("tcp", port)
  if err != nil{
    log.Fatal(err)
  }

  fmt.Println("Server Listen. port: ", port)

  addclient := make(chan Client)
  deleteclient := make(chan Client)
  msgchan := make(chan string)

  go distribute(addclient, deleteclient, msgchan)

  for{
    conn, err := ln.Accept()
    if err != nil{
      continue
    }

    go handle(conn, addclient, deleteclient, msgchan)
  }
}
