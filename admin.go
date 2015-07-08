package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

func adminStart() {
	go func() {
		l, err := net.Listen("tcp", ":2000")
		if err != nil {
			log.Fatal(err)
		}
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleCommand(conn)
		}
	}()
}

func handleCommand(c net.Conn) {
	defer c.Close()
	for {
		inBuf := make([]byte, 128)
		n, err := c.Read(inBuf)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(c, "read command error: %s\n", err)
			}
			continue
		}
		cmd := string(inBuf[:n-2])
		switch cmd {
		case "shutdown":
			_, err = shutdown()
			if err != nil {
				fmt.Fprintln(c, "shutdown server failed")
			}
			break
		case "stats":
			outBuf := stats()
			fmt.Fprintln(c, outBuf)
		default:
			fmt.Fprintln(c, "no known command")
		}
	}
}

func shutdown() (string, error) {
	return "shutdown successfully", nil
}

func stats() string {
	return "print stats"
}
