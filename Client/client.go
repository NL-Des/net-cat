package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const (
	PORT = "8989"
	IP   = "127.0.0.1"
)

func gestionErreurs(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	//----------"net.Dial" cherche une adresse précise et tente de connecter pour envoyer et recevoir des données.
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", IP, PORT))
	gestionErreurs(err)

	for {
		// entrée utilisateur, tout ce qui est tapé sur le clavier est stocké en mémoire tampon pour être envoyé quant il appuie sur la touche entrée.
		//--------"bufio" permet de lire les données d'une source (clavier, connexion réseau, fichier,...)
		//-------------"NewReader" créé un tampon qui stocke les données lues.
		//-----------------------"os.Stdin" lit ce qui est tapé sur le clavier.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Client: ")
		//"text" est le message en string écrit depuis le clavier.
		//----------"reader" est la contraction de "bufio.Reader", ce qui permet de lire la variable "reader" précédente.
		//-----------------"ReadString()" va tout lire jusqu'à rencontrer la limite indiquée, ici c'est  '\n'.
		text, err := reader.ReadString('\n')
		gestionErreurs(err)

		// Transmission du message au serveur.
		// Les transmissions de messages se font toujours Bytes, il faut donc toujours convertir avant de transmettre.
		conn.Write([]byte(text))

		// On écoute tous les messages émis par le serveur, on ajoute un saut à la ligne.
		//"message" c'est la transmission du serveur.
		//--------------bufio.NewReader(conn) va créer un tampon de lecture pour lire les données reçues.
		//-----------------------------------"ReadString('\n)" va lire les données jusqu'au saut de ligne.
		message, err := bufio.NewReader(conn).ReadString('\n')
		gestionErreurs(err)

		// On affiche le message reçu du serveur.
		fmt.Print("serveur : " + message)
	}
}
