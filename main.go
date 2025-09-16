// MARK: Instructions
// Commande de lancement pour le serveur : ./TCPChat <PORT>
// Commande de connection du client au serveur : nc  localhost <PORT>

package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// MARK: Dessin d'accueil
// émis une fois le nom d'utilisateur choisit.
const asciiArt = `
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    ` + "`" + `.       | ` + "`" + `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     ` + "`" + `-'       ` + "`" + `--'
`

// MARK: Map et variables
// Pour stocker toutes les connexions actives (IP, Noms)

var clients = make(map[net.Conn]bool)
var userNames = make(map[net.Conn]string)

// Permet de bloquer toutes les goroutines pour exécuter l'opération désirée.
var clientsMutex sync.Mutex

// Origine et contenu du message.
type Message struct {
	ComeFrom string
	Content  string
	timer    string
}

// Canal global de circulation des messages.
var channels = make(chan Message)

// Permet d'éviter les appels/créations de variables entre fonctions.
var (
	ln   net.Listener
	port int
)

// Historique des conversations que chaque nouvel arrivant reçoit à la connexion.
var historique []Message

// MARK: Main
func main() {
	go messageHandler() // goroutines tourne en arrière fond pour gérer tous les messages.
	server()
}

// MARK: Erreurs
func gestionDesErreurs(err error) {
	if err != nil {
		panic(err)
	}
}

// MARK: Serveur
func server() {

	// Execution de la commande dans le terminal pour obtenir l'adresse IP locale (Seulement sur Linux et MacOS).
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.Output()
	gestionDesErreurs(err)
	before, _, _ := strings.Cut(string(output), " ")
	fmt.Printf("[SERVER]: L'adresse IP de l'ordinateur exécutant le programme est %s \n", before)
	ip := before

	// Condition d'initialisation du serveur à partir de la commande : ./TCPChat <PORT> localhost
	if len(os.Args) == 1 {
		port = 8989
		// ip = ip de l'ordinateur exécutant.
	} else if len(os.Args) == 2 && len(os.Args[1]) == 4 {
		port, err = strconv.Atoi(os.Args[1])
		gestionDesErreurs(err)
		// ip = ip de l'ordinateur exécutant.
		// Option si l'on veut pouvoir rajouter une adresse IP à la main.
		/* 	} else if len(os.Args) == 3 && len(os.Args[1]) == 4 && (len(os.Args[2]) >= 7 && len(os.Args[2]) <= 15) {
		port, err = strconv.Atoi(os.Args[1])
		gestionDesErreurs(err)
		ip = os.Args[2] */
	} else {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	fmt.Printf("[SERVER]: Le PORT de connexion est %d \n", port)

	ln, err = net.Listen("tcp", fmt.Sprintf("%s:%s", ip, strconv.Itoa(port)))
	gestionDesErreurs(err)
	fmt.Println("[SERVER]: Serveur lancé.")
	fmt.Println("[SERVER]: En attente de connexion des utilisateurs.")

	// Boucle pour gérer plusieurs connexions de différents clients.
	for {
		connexions, err := ln.Accept()
		gestionDesErreurs(err)

		fmt.Printf("[SERVER]: Nouvelle connexion de : %s\n", connexions.RemoteAddr().String())

		// Ajouter le client à la liste des clients connectés.
		// Sécurité pour éviter que deux utilisateurs n'agissent en même temps. (mutex)
		clientsMutex.Lock() // Bloque toutes les goroutines.
		if len(clients) >= 10 {
			clientsMutex.Unlock()
			connexions.Write([]byte("[SERVER]: Nous avons déjà 10 utilisateurs en ligne, veuillez patienter qu'une place se libère.\n"))
			connexions.Close()
			continue
		}
		clients[connexions] = true // Dans la MAP clients l'utilisateur est enregistré comme actif.
		clientsMutex.Unlock()      // Re-ouvre l'accès au goroutines.

		// Utilisation de go routines pour gérer plusieurs clients en même temps.
		go handleConnexion(connexions)
	}
}

// MARK: Gestion des connexions
// initialisation d'une fonction par utilisateur.
func handleConnexion(connexions net.Conn) {

	// Envoi au nouvel utilisateur du Pingouin en Ascii-Art.
	connexions.Write([]byte(asciiArt))

	// Demande du nom de l'utilisateur.
	userName := nameWithoutBlank(connexions)

	fmt.Printf("[SERVER]: Pour %s, acquisition du nom : %s \n", connexions.RemoteAddr().String(), userName)

	// Stockage des noms des utilisateurs.
	clientsMutex.Lock()
	userNames[connexions] = userName
	clientsMutex.Unlock()

	//collectiveMessageConnexion(userName)

	// Envoi de l'historique des conversations au nouvel arrivant.
	clientsMutex.Lock()
	for i := 0; i < len(historique); i++ {
		_, err := connexions.Write([]byte(fmt.Sprintf("%s[%s]: %s", historique[i].timer, historique[i].ComeFrom, historique[i].Content)))
		gestionDesErreurs(err)
	}
	clientsMutex.Unlock()

	collectiveMessageConnexion(userName)

	// Retire le client de la liste quand il se déconnecte.
	// C'est une fonction anonyme.
	// defer func va s'éxécuter une fois que la fonction où elle se trouve se termine.
	defer func() {
		fmt.Printf("[SERVER]: Client déconnecté : %s\n", userName)
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
			return
		}
		content := string(byteMessage[:n])
		if content == "\n" {
			continue
		}

		if strings.HasPrefix(content, ":/rename") {
			oldUserName := userNames[connexions]
			userName = Rename(connexions)
			clientsMutex.Lock()
			userNames[connexions] = userName
			clientsMutex.Unlock()
			collectiveMessageRename(userName, oldUserName)
			continue
		}
		// Envoyer le message dans le canal pour diffusion
		channels <- Message{ComeFrom: userNames[connexions], Content: content}
	}
}

// MARK: Nom utilisateur
// L'utilisateur doit définir un Nom, non vide.
func Rename(connexions net.Conn) string {
	var userName string
	for {
		nameBuffer := make([]byte, 1024)

		// Demande du nom de l'utilisateur.
		_, err := connexions.Write([]byte("Vous voulez changer de nom ? Pas de soucis, comment voulez vous vous rennomez ? \n"))
		gestionDesErreurs(err)
		// Lecture du nom de l'utilisateur.
		name, err := connexions.Read(nameBuffer)
		gestionDesErreurs(err)

		userName = strings.TrimSpace(string(nameBuffer[:name]))
		fmt.Println(userName)
		if userName != "" { // Si Nom, non vide, sortie de la boucle For.
			break
		}
		connexions.Write([]byte("Votre patronyme ne puis être sans caractère, veuillez retenter votre essais."))
	}
	return userName
}

func nameWithoutBlank(connexions net.Conn) string {

	var userName string

	for {
		nameBuffer := make([]byte, 1024)

		// Demande du nom de l'utilisateur.
		_, err := connexions.Write([]byte("[SERVER]: Bienvenue ! Veuillez saisir votre nom : \n"))
		gestionDesErreurs(err)

		// Lecture du nom de l'utilisateur.
		name, err := connexions.Read(nameBuffer)
		gestionDesErreurs(err)

		userName = strings.TrimSpace(string(nameBuffer[:name]))

		if userName == "" { // Si Nom, non vide, sortie de la boucle for.
			connexions.Write([]byte("[SERVER]: Votre patronyme ne puis être sans caractère, Keep Calm and Proceed. \n"))
			continue
		}
		if nameAlreadyPresent(userName) { // Si Nom en double.
			connexions.Write([]byte("[SERVER]: Nom déjà utilisé, veuillez en saisir un autre. \n"))
			continue
		}

		//Si le nom est valide et unique.
		break
	}

	return userName
}

// MARK: Nom déjà présent.
func nameAlreadyPresent(userName string) bool {
	for _, nameBis := range userNames {
		if userName == nameBis { // Si nom existe déjà, sortie de la boucle for.
			return true
		}
	}
	return false
}

// MARK: Transmission des messages
// Gestionnaire de messages qui diffuse le massage d'un client à tous les clients.
func messageHandler() {
	for {
		msg := <-channels
		timeLog := time.Now().Format("[2006-01-02 15:04:05]")
		fmt.Printf("%s[%s]: %s", timeLog, msg.ComeFrom, msg.Content)                                           // écriture partie terminal du serveur.
		historique = append(historique, Message{ComeFrom: msg.ComeFrom, Content: msg.Content, timer: timeLog}) // Pour archiver les messages.

		// Diffuser le message à tous les clients connectés
		clientsMutex.Lock()
		for client := range clients {
			// timeLog := time.Now().Format("[2006-01-02 15:04:05]")
			_, err := client.Write([]byte(fmt.Sprintf("%s[%s]: %s", timeLog, msg.ComeFrom, msg.Content)))
			if err != nil {
				// Si l'écriture échoue, supprimer le client de la struct.
				delete(clients, client)
				client.Close()
			}
		}
		clientsMutex.Unlock()
	}
}

// MARK: Message d'accueil
// Envoi du message collectif d'accueil.
func collectiveMessageConnexion(userName string) {
	clientsMutex.Lock()
	historique = append(historique, Message{ComeFrom: "Serveur", timer: fmt.Sprint(time.Now().Format("[2006-01-02 15:04:05]")), Content: fmt.Sprintf("Veuillez accueillir comme il se le doit : %s \n", userName)})
	defer clientsMutex.Unlock()

	for client := range clients {
		_, err := client.Write([]byte(fmt.Sprintf(time.Now().Format("[2006-01-02 15:04:05]")+"[SERVER] : Veuillez accueillir comme il se le doit : %s \n", userName)))
		gestionDesErreurs(err)
	}
	/* var message []Message
	historique = append(historique, msg)
	fmt.Printf(msg.ComeFrom)
	fmt.Printf(msg.Content) */
}

// MARK: Message de départ
func collectiveMessageRename(userName string, oldUserName string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		_, err := client.Write([]byte(fmt.Sprintf(time.Now().Format("[2006-01-02 15:04:05]")+"[Serveur] : Notre bien aimé %s se prénome maintenant %s \n", oldUserName, userName)))
		gestionDesErreurs(err)
	}
	historique = append(historique, Message{ComeFrom: "Serveur", timer: fmt.Sprint(time.Now().Format("[2006-01-02 15:04:05]")), Content: fmt.Sprintf("Notre bien aimé %s se prénome maintenant %s \n", oldUserName, userName)})
}

// Envoi du message collectif de départ.
func collectiveMessageDeconnexion(userName string) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		_, err := client.Write([]byte(fmt.Sprintf(time.Now().Format("[2006-01-02 15:04:05]")+"[SERVER] : Que nenni ?! Un folâtre prendre campagne ! Diable, en voilà un apache. Que son nom soit connu de tous pour sa vilenie : %s \n", userName)))
		gestionDesErreurs(err)
	}
	historique = append(historique, Message{ComeFrom: "Serveur", timer: fmt.Sprint(time.Now().Format("[2006-01-02 15:04:05]")), Content: fmt.Sprintf("Que nenni ?! Un folâtre osa partir ! Diable, en voilà un apache. Que son nom soit connu de tous pour sa vilenie : %s \n", userName)})
}
