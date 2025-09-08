package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	IP = "127.0.0.1"
)

// Map pour stocker toutes les connexions actives (IP, Noms)
var clients = make(map[net.Conn]bool)
var userNames = make(map[net.Conn]string)
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

// Historique des conversations que chaque nouvel arrivant reçoit à la connexion.
var historique []Message

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

// Gestion des connexions.
func handleConnexion(connexions net.Conn) {

	// Demande du nom de l'utilisateur.
	_, err := connexions.Write([]byte("Bienvenue ! Veuillez saisir votre nom : \n"))
	gestionDesErreurs(err)

	// Lecture du nom de l'utilisateur.
	nameBuffer := make([]byte, 1024)
	name, err := connexions.Read(nameBuffer)
	gestionDesErreurs(err)
	userName := strings.TrimSpace(string(nameBuffer[:name]))
	fmt.Printf("Pour %s, acquisition du nom : %s \n", connexions.RemoteAddr().String(), userName)

	// Stockage des noms des utilisateurs.
	clientsMutex.Lock()
	userNames[connexions] = userName
	clientsMutex.Unlock()

	collectiveMessageConnexion(userName)

	// Envoi de l'historique des conversations au nouvel arrivant.
	clientsMutex.Lock()
	for _, msg := range historique {
		_, err := connexions.Write([]byte(fmt.Sprintf("[%s]: %s\n", msg.ComeFrom, msg.Content)))
		gestionDesErreurs(err)
	}
	clientsMutex.Unlock()

	// Retire le client de la liste quand il se déconnecte.
	defer func() {
		fmt.Printf("Client déconnecté : %s\n", userName)
		clientsMutex.Lock()
		delete(clients, connexions)
		delete(userNames, connexions)
		clientsMutex.Unlock()
		connexions.Close()
		collectiveMessageDeconnexion(userName)
	}()

	// Boucle de réception des messages.
	byteMessage := make([]byte, 1024)
	for {
		n, err := connexions.Read(byteMessage)
		if err != nil {
			return // Le client s'est déconnecté
		}
		content := string(byteMessage[:n])
		// Envoyer le message dans le canal pour diffusion
		channels <- Message{ComeFrom: userNames[connexions], Content: content}
	}
}

// Gestionnaire de messages qui diffuse les messages à tous les clients
func messageHandler() {
	for {
		msg := <-channels
		fmt.Printf("[%s] a envoyé : %s", msg.ComeFrom, msg.Content)
		historique = append(historique, msg) // Pour archiver les conversations.

		// Diffuser le message à tous les clients connectés
		clientsMutex.Lock()
		for client := range clients {
			_, err := client.Write([]byte(fmt.Sprintf("[%s]: %s", msg.ComeFrom, msg.Content)))
			if err != nil {
				// Si l'écriture échoue, supprimer le client de la struct.
				delete(clients, client)
				client.Close()
			}
		}
		clientsMutex.Unlock()
	}
}

// Envoi du message collectif d'accueil.
func collectiveMessageConnexion(userName string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		_, err := client.Write([]byte(fmt.Sprintf("[Serveur] : Veuillez accueillir comme il se le doit : %s \n", userName)))
		gestionDesErreurs(err)
	}
}

// Envoi du message collectif de départ.
func collectiveMessageDeconnexion(userName string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		_, err := client.Write([]byte(fmt.Sprintf("[Serveur] : Que nenni ?! Un folâtre osa partir ! Diable, en voilà un apache. Que son nom soit connu de tous pour sa vilenie : %s \n", userName)))
		gestionDesErreurs(err)
	}
}
