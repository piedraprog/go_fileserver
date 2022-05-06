package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"io"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)



type Message struct {
	Message string `json:"message"`

}

var (
	// Clientes conectados
	clientes = make(map[*websocket.Conn]bool)
	// Broadcast channel
	broadcast = make(chan Message)

	wsUpgrader = websocket.Upgrader{}

	wsConn *websocket.Conn
)

func WsEndpoint(w http.ResponseWriter, r *http.Request) {

	wsUpgrader.CheckOrigin = func(r *http.Request) bool {
		// check the http.Request
		// make sure it's OK to access
		return true
	}
	var err error
	wsConn, err = wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("could not upgrade: %s\n", err.Error())
		return
	}

	defer wsConn.Close()

	clientes[wsConn] = true
	// event loop
	for {
		var msg Message
		err := wsConn.ReadJSON(&msg)
		if err != nil {
			fmt.Printf("error reading JSON: %s\n", err.Error())
			delete(clientes, wsConn)
			break
		}

		broadcast <- msg

		fmt.Printf("Message Received: %s\n", msg.Message)
		// SendMessage("Hello, Client!")
	}
}

func handleMensajes() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		fmt.Println("hay esta cantidad de clientes:",len(clientes))
			// Send it out to every client that is currently connected
		for client := range clientes {

			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clientes, client)
			}
		}
	}
}

func FileHandler(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")

	if err != nil {
		fmt.Println("Error al recibir el archivo",err)
		return
	}

	defer file.Close()

	out, pathError := os.Create("./files/" + header.Filename)

	if pathError != nil {
		log.Println("error al crear un archivo para escribir: ", pathError)
		return
	}
	defer out.Close()

	_,copyFileerror := io.Copy(out, file)
	if copyFileerror != nil {
		log.Println("error al copiar el archivo: ", copyFileerror)
		return
	}

	w.Write([]byte("Archivo subido con exito %s\n"))
	fmt.Printf("Archivo subido con exito: %s\n", header.Filename)
	fmt.Printf("La ruta es: %s\n", out.Name())
}


func Fileservice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Servicio de archivos")
	fmt.Fprintf(w, "Servicio de archivos")
}

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/socket", WsEndpoint)
	go handleMensajes()

	router.HandleFunc("/upload", FileHandler)
	router.HandleFunc("/files",Fileservice)
	
	
	fmt.Println("Server is running on port 9100")
	log.Fatal(http.ListenAndServe(":9100", router))

}
