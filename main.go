package main

import (
	"os"
	"golang.org/x/crypto/ssh"
	"log"
	"fmt"
	"bufio"
	"net"
	"crypto/rsa"
	"crypto/rand"
)

const (
	NAME   string = "chatty"
	SERVER string = "www.hwr.io"
	PORT   int    = 2022
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		log.Fatalln("failed to generate RSA key")
	}

	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		log.Fatalf("unable to parse private key: %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: NAME,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", SERVER, PORT), sshConfig)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Fatalln("unable to create SSH session")
	}

	in, err := session.StdinPipe()
	if err != nil {
		log.Fatalln("unable to create stdin pipe to SSH session")
	}

	in.Write([]byte("/theme mono\r\n"))

	out, err := session.StdoutPipe()
	if err != nil {
		log.Fatalln("unable to create stdout pipe to SSH session")
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)
	}

	fmt.Println("connection closed")
	os.Exit(0)
}
