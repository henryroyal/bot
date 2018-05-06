package main

import (
	"os"
	"golang.org/x/crypto/ssh"
	"log"
	"fmt"
	"bufio"
	"crypto/rsa"
	"crypto/rand"
	"io/ioutil"
	"strings"
	"errors"
	"strconv"
	"time"
)

var (
	// Environment Variables
	BOTNAME string = os.Getenv("BOT_NAME")
	VERSION string = os.Getenv("BOT_VERSION")
	HOST    string = os.Getenv("BOT_HOST")
	PORT    string = os.Getenv("BOT_PORT")

	BOT_PRIVATE_KEY string = os.Getenv("BOT_PRIVATE_KEY")
	HOST_PUBLIC_KEY string = os.Getenv("HOST_PUBLIC_KEY")

	ALLOW_INSECURE_HOSTKEY string = os.Getenv("ALLOW_INSECURE_HOSTKEY")
	HISTORY_PLAYBACK_LEN   string = os.Getenv("HISTORY_PLAYBACK_LEN")

	// Errors
	InsecureHostkeyError error = errors.New("Set ALLOW_INSECURE_HOSTKEY=true to allow ssh.InsecureIgnoreHostKey callback")
)

func getHostPubkeyCallback(hostPublicKeyPath string, allowInsecureHostkey bool) (ssh.HostKeyCallback, error) {

	if hostPublicKeyPath == "" && allowInsecureHostkey == false {
		return nil, InsecureHostkeyError
	}

	if hostPublicKeyPath == "" && allowInsecureHostkey == true {
		return ssh.InsecureIgnoreHostKey(), nil
	}

	keyBytes, err := ioutil.ReadFile(HOST_PUBLIC_KEY)
	if err != nil {
		return nil, err
	}

	hostkey, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes)
	if err != nil {
		return nil, err
	}

	return ssh.FixedHostKey(hostkey), nil
}

func getKeySigner(botPrivateKeyPath string) (ssh.Signer, error) {

	var signer ssh.Signer

	if botPrivateKeyPath == "" {
		key, err := rsa.GenerateKey(rand.Reader, 2014)
		if err != nil {
			return nil, err
		}

		signer, err = ssh.NewSignerFromKey(key)
		if err != nil {
			return nil, err
		}

	} else {
		key, err := ioutil.ReadFile(botPrivateKeyPath)
		if err != nil {
			return nil, err
		}

		signer, err = ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
	}

	return signer, nil
}

func getClientConfig(botName string, botKeyPath string, hostPubkeyPath string, allowInsecure bool) *ssh.ClientConfig {

	hostkeyCallback, err := getHostPubkeyCallback(hostPubkeyPath, allowInsecure)
	if err != nil {
		log.Fatalf("FATAL: getHostPubkeyCallback %v", err)
	}

	if hostPubkeyPath == "" {
		log.Println("warning: using ssh.InsecureIgnoreHostKey")
	}

	keySigner, err := getKeySigner(botKeyPath)
	if err != nil {
		log.Fatalf("FATAL: getKeySigner %v", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: botName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(keySigner),
		},
		HostKeyCallback: hostkeyCallback,
	}

	return sshConfig
}

func main() {

	allowInsecureHostkey, err := strconv.ParseBool(ALLOW_INSECURE_HOSTKEY)
	if err != nil {
		log.Fatalf("failed to parse ALLOW_INSECURE_HOSTKEY into bool: %v", err)
	}

	historyRead := 0 // on connect, server sends last 20 lines to client - ignore these
	historyLines, err := strconv.Atoi(HISTORY_PLAYBACK_LEN)
	if err != nil {
		log.Fatalf("failed to cast HISTORY_PLAYBACK_LEN to integer: %v", err)
	}

	sshConfig := getClientConfig(BOTNAME, BOT_PRIVATE_KEY, HOST_PUBLIC_KEY, allowInsecureHostkey)

	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", HOST, PORT), sshConfig)
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
	defer in.Close()

	greeting := fmt.Sprintf("/theme mono\r\n%v v%v: https://www.youtube.com/watch?v=SFLSOIufuhM\r\n", BOTNAME, VERSION)
	in.Write([]byte(greeting))

	out, err := session.StdoutPipe()
	if err != nil {
		log.Fatalln("unable to create stdout pipe to SSH session")
	}

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		if historyRead < historyLines {
			historyRead++
			continue
		}
		line := scanner.Text()
		log.Print(line)
		time.Sleep(1 * time.Second)

		if strings.Contains(line, fmt.Sprintf("hi %v", BOTNAME)) {
			in.Write([]byte("hello\r\n"))
		}
	}

	fmt.Println("connection closed")
	os.Exit(0)
}
