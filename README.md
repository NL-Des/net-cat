```
                      /^--^\     /^--^\     /^--^\
                      \____/     \____/     \____/
                     /      \   /      \   /      \
                    |        | |        | |        |
                     \__  __/   \__  __/   \__  __/
|^|^|^|^|^|^|^|^|^|^|^|^\ \^|^|^|^/ /^|^|^|^|^\ \^|^|^|^|^|^|^|^|^|^|^|^|
| | | | | | | | | | | | |\ \| | |/ /| | | | | | \ \ | | | | | | | | | | |
########################/ /######\ \###########/ /#######################
| | | | | | | | | | | | \/| | | | \/| | | | | |\/ | | | | | | | | | | | |
|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|_|
```
## 🧵 Présentation du projet NetCat
Ce projet recrée le comportement de **NetCat (`nc`)** en Go pour établir un **chat de groupe TCP**.  
Le serveur écoute sur un port spécifié et gère jusqu’à **10 clients** qui peuvent :
- Échanger des messages en temps réel
- Recevoir l’historique complet à leur arrivée
- Être informés des connexions/déconnexions des autres participants

```
  ,-.       _,---._ __  / \
 /  )    .-'       `./ /   \
(  (   ,'            `/    /|
 \  `-"             \'\   / |
  `.              ,  \ \ /  |
   /`.          ,'-`----Y   |
  (            ;        |   '
  |  ,-.    ,-'         |  /
  |  | (   |            | /
  )  |  \  `.___________|/
  `--'   `--'
```
---
## ⚙️ Fonctionnalités attendues

### Serveur :
-Si aucun port n'est spécifié au lancement du serveur, il faut utiliser le port 8989.

Sinon, afficher l’usage correct :

    ```[USAGE]: ./TCPChat $port```

### Le chat :
- Maximum 10 connexions simultanées.
- Refuser toute connexion supplémentaire avec un message d’erreur.

### Les utilisateurs :
-Quand un nouveau participant arrive, le serveur informe toutes les personnes connectées de qui arrive sur le chat.
-Quand un participant part, le serveur informe toutes les personnes connectées de qui quitte le chat.
-À la connexion, chaque participant doit fournir un nom, non vide.
-Lorsqu’un nouveau participant rejoint le chat, il reçoit tous les messages précédemment échangés.

### Communication sur le chat :
-Tout le monde doit voir les messages de tout le monde.
-Tout le monde doit pouvoir poster des messages.
-Les messages vides ne doivent pas êtres diffusés.

### Format des messages :
```-[YYYY-MM-DD HH:MM:SS][NomClient]:[Message]```
---
       _                        
       \`*-.                    
        )  _`-.                 
       .  : `. .                
       : _   '  \               
       ; *` _.   `*-._          
       `-.-'          `-.       
         ;       `       `.     
         :.       .        \    
         . \  .   :   .-'   .   
         '  `+.;  ;  '      :   
         :  '  |    ;       ;-. 
         ; '   : :`-:     _.`* ;
      .*' /  .*' ; .*`- +'  `*' 
      `*-*   `*-*  `*-*'
## 🛠️ Instructions de développement

-Langage autorisé : Go

-Techniques attendues :

    *Goroutines

    *Channels

    *Mutexes

    *Gestion des erreurs autant du côté serveur que du côté client.

### 📦 Packages autorisés :
io, log, os, fmt, net, sync, time, bufio, errors, strings, reflect

## 🚀 Usages attendus :

### Séquence de lancement du serveur :

```bash
Listening on the port :8989
$ go run . 2525
Listening on the port :2525
$ go run . 2525 localhost
[USAGE]: ./TCPChat $port
```

### Séquence d'ouverture du chat :
```
$ nc $IP $port
Welcome to TCP-Chat!
         _nnnn_
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
 |    `.       | `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     `-'       `--'
[ENTER YOUR NAME]:
```
## 🎁 Bonus :
-Changement de nom en cours de session avec notification.

-Multiples salons de chat.

-Options NetCat supplémentaires (flags).

-Interface terminal graphique autorisée :  https://github.com/jroimartin/gocui

-Sauvegarde des logs dans un fichier.