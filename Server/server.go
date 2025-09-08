package server

import (
	"bufio"
	"fmt"
	"net"
)

// func main() {
// utiliser goroutines
// Explication : https://devopssec.fr/article/goroutines-golang
// utiliser channels
// Explication : https://devopssec.fr/article/channels-golang
// utiliser Mutexes
// Explications : cela sert à verrouiller l'accès à une ressource commune.
// tuto ? https://devopssec.fr/article/tp-creer-application-de-chat-golang#begin-article-section
// Ressources autres : https://www.commandlinux.com/man-page/man1/nc.1.html
// }

// Définition des données de connexions.
const (
	PORT = "8989"
	IP   = "127.0.0.1"
)

// Pour répondre à toutes les erreurs et réduire les if présents dans le code.
func gestionDesErreurs(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	server()
}

func gestionDePlusieursClients(conn net.Conn) {
	// On écoute tous les messages émis par le client, on ajoute un saut à la ligne.
	//"message" c'est la transmission du serveur.
	//--------------bufio.NewReader(conn) va créer un tampon de lecture pour lire les données reçues.
	//-----------------------------------"ReadString('\n)" va lire les données jusqu'au saut de ligne.
	message, err := bufio.NewReader(conn).ReadString('\n')
	gestionDesErreurs(err)

	// Affichage du message reçu du client.
	fmt.Print("Client:", string(message))
}

// Lancement et gestion du serveur
func server() {
	fmt.Println("Lancement du serveur en cours...")
	fmt.Println("En attente de la connexion d'un client")

	// On écoute sur le port choisit.
	//"ln" est une contraction de ln.Accept(), qui attend une connexion avant de s'exécuter.
	//---------net.Listen() écoute les connexions entrantes sur l'adresse indiquée.
	//-------------------("tcp est le nom du protocole de communication")
	//--------------------------------------(assemblage des données de "const")
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%s", IP, PORT))
	gestionDesErreurs(err)

	var clients []net.Conn // tableau des clients

	// Boucle pour que le serveur soit toujours à l'écoute de nouvelles connexions entrantes.
	for {

		// On accepte les connexions sur le port choisit.
		// "connexions" est maintenant un objet de type "net.Conn" qui représente la connexion.
		connexions, err := ln.Accept()
		if err == nil {
			clients = append(clients, connexions) // Quand un client se connecte on le rajoute à notre tableau.
		}
		gestionDesErreurs(err)

		// Ecriture des informations sur les clients qui se connectent.
		//----------------------------------------- "connexions.RemoteAddr()" cela permet d'obtenir l'adresse IP et le port du client qui s'est connecté au serveur grâce à l'objet "net.Conn"
		fmt.Println("Un client est connecté depuis", connexions.RemoteAddr())

		// On écoute les messages émis par les clients.
		buffer := make([]byte, 4096) // Pour éviter une surcharge, on limite la capacité d'envoi dans un message du client à 4096 bytes.
		// "length" c'est le nombre réel d'octets lus.
		//-----------------------.Read(buffer) il va lire le message qui est une slice de bytes.
		length, err := connexions.Read(buffer)
		message := string(buffer[:length]) // Supprime les bits (0 ou 1) inutiles et converti les bytes en string.

		if err != nil {
			fmt.Println("Le client s'est déconnecté")
		}

		// Affichage test du message.
		fmt.Println("Client :", message)

		// Transmission du message au client.
		// Les transmissions de messages se font toujours Bytes, il faut donc toujours convertir avant de transmettre.
		connexions.Write([]byte(message + "\n"))
	}

}
