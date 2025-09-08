/* Lignes de commandes pour NC :

OpenBSD netcat (Debian patchlevel 1.226-1ubuntu2)
usage: nc [-46CDdFhklNnrStUuvZz] [-I length] [-i interval] [-M ttl]
          [-m minttl] [-O length] [-P proxy_username] [-p source_port]
          [-q seconds] [-s sourceaddr] [-T keyword] [-V rtable] [-W recvlimit]
          [-w timeout] [-X proxy_protocol] [-x proxy_address[:port]]
          [destination] [port]
        Command Summary:
                -4              Use IPv4
                -6              Use IPv6
                -b              Allow broadcast
                -C              Send CRLF as line-ending
                -D              Enable the debug socket option
                -d              Detach from stdin
                -F              Pass socket fd
                -h              This help text
                -I length       TCP receive buffer length
                -i interval     Delay interval for lines sent, ports scanned
                -k              Keep inbound sockets open for multiple connects
                -l              Listen mode, for inbound connects
                -M ttl          Outgoing TTL / Hop Limit
                -m minttl       Minimum incoming TTL / Hop Limit
                -N              Shutdown the network socket after EOF on stdin
                -n              Suppress name/port resolutions
                -O length       TCP send buffer length
                -P proxyuser    Username for proxy authentication
                -p port         Specify local port for remote connects
                -q secs         quit after EOF on stdin and delay of secs
                -r              Randomize remote ports
                -S              Enable the TCP MD5 signature option
                -s sourceaddr   Local source address
                -T keyword      TOS value
                -t              Answer TELNET negotiation
                -U              Use UNIX domain socket
                -u              UDP mode
                -V rtable       Specify alternate routing table
                -v              Verbose
                -W recvlimit    Terminate after receiving a number of packets
                -w timeout      Timeout for connects and final net reads
                -X proto        Proxy protocol: "4", "5" (SOCKS) or "connect"
                -x addr[:port]  Specify proxy address and port
                -Z              DCCP mode
                -z              Zero-I/O mode [used for scanning]
        Port numbers can be individual or ranges: lo-hi [inclusive]*/

/*
Instructions :
-Pour lancer le serveur : go run main.go 8989
-Pour se connecter au serveur : nc 127.0.0.1 8989
*/

package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

const (
	IP = "127.0.0.1"
)

// Map pour stocker toutes les connexions actives
var clients = make(map[net.Conn]bool)
var clientsMutex sync.Mutex

// Origine et contenu du message.
type Message struct {
	ComeFrom string
	Content  string
}

// Canal global de circulation des messages.
var channels = make(chan Message)

var ( // Permet d'éviter les appels de variables entre fonctions.
	ln   net.Listener
	port int
	err  error
)

func main() {
	go messageHandler()
	server()
}

func gestionDesErreurs(err error) {
	if err != nil {
		panic(err)
	}
}

func server() {

	// Condition pour le PORT.
	if len(os.Args) != 2 {
		fmt.Println("Usage attendu : go run main.go <8989>")
		return
	} else if len(os.Args[1]) == 0 {
		port = 8989
	} else {
		port, err = strconv.Atoi(os.Args[1])
		gestionDesErreurs(err)
	}

	ln, err = net.Listen("tcp", fmt.Sprintf("%s:%s", IP, strconv.Itoa(port)))
	gestionDesErreurs(err)
	fmt.Println("Serveur lancé.")
	fmt.Println("En attente de connexion des utilisateurs.")

	// Boucle pour gérer plusieurs connexions de différents clients.
	for {
		connexions, err := ln.Accept()
		gestionDesErreurs(err)
		// go handleConnexion(connexions) // Doublons non nécessaire ?

		fmt.Printf("Nouvelle connexion de : %s\n", connexions.RemoteAddr().String())

		// Ajouter le client à la liste des clients connectés.
		// Sécurité pour éviter que deux utilisateurs n'agissent en même temps. (mutex)
		clientsMutex.Lock()        // Bloque l'accès de la MAP clients aux goroutines
		clients[connexions] = true // Dans la MAP clients l'utilisateur est enregistré comme actif.
		clientsMutex.Unlock()      // Re-ouvre l'accès au goroutines de Clients.

		// Utilisation de go routines pour gérer plusieurs clients en même temps.
		go handleConnexion(connexions)
	}
}

func handleConnexion(connexions net.Conn) {
	defer func() {
		// Retirer le client de la liste quand il se déconnecte
		clientsMutex.Lock()
		delete(clients, connexions)
		clientsMutex.Unlock()
		connexions.Close()
		fmt.Printf("Client déconnecté : %s\n", connexions.RemoteAddr().String())
	}()

	byteMessage := make([]byte, 1024)

	for {
		n, err := connexions.Read(byteMessage)
		if err != nil {
			return // Le client s'est déconnecté
		}

		content := string(byteMessage[:n])
		// Envoyer le message dans le canal pour diffusion
		channels <- Message{ComeFrom: connexions.RemoteAddr().String(), Content: content}
	}
}

// Gestionnaire de messages qui diffuse les messages à tous les clients
func messageHandler() {
	for {
		msg := <-channels
		fmt.Printf("[%s] a envoyé : %s", msg.ComeFrom, msg.Content)

		// Diffuser le message à tous les clients connectés
		clientsMutex.Lock()
		for client := range clients {
			// Ne pas renvoyer le message à l'expéditeur
			if client.RemoteAddr().String() != msg.ComeFrom {
				_, err := client.Write([]byte(fmt.Sprintf("[%s]: %s", msg.ComeFrom, msg.Content)))
				if err != nil {
					// Si l'écriture échoue, supprimer le client
					delete(clients, client)
					client.Close()
				}
			}
		}
		clientsMutex.Unlock()
	}
}
