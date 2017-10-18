package main

import (
	"cheng/agent/utils"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/websocket"
)

func handleClient(conn *websocket.Conn, readDeadline time.Duration) {
	defer utils.PrintPanicStack()

	host, port, err := net.SplitHostPort(conn.Request().RemoteAddr)
	if err != nil {
		log.Error("cannot get remote address:", err)
		return
	}
	log.Infof("new ws connection from:%v port: %v", host, port)

	out := make(chan []byte, 64)
	exit := make(chan struct{})

	go func(conn *websocket.Conn) {
		for {
			select {
			case <-exit:
				return
			case msg := <-out:
				conn.Write(msg)
			}
		}
	}(conn)

	for {
		// solve dead link problem:
		// physical disconnection without any communcation between client and server
		// will cause the read to block FOREVER, so a timeout is a rescue.
		conn.SetReadDeadline(time.Now().Add(readDeadline))

		buf := make([]byte, 256)

		len, err := conn.Read(buf)
		if err != nil {
			log.Error(err)
			close(exit)
			return
		}
		log.Println(string(buf[0:len]))
		out <- buf[0:len]
	}
}

func main() {
	pool := x509.NewCertPool()
	caCertPath := "./keys/ca/ca.crt"

	caCrt, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		fmt.Println("ReadFile err:", err)
		return
	}

	pool.AppendCertsFromPEM(caCrt)

	mux := http.NewServeMux()
	mux.Handle("/", websocket.Handler(func(conn *websocket.Conn) {
		conn.PayloadType = websocket.BinaryFrame
		handleClient(conn, 3*time.Second)
	}))

	server := &http.Server{
		Addr:    ":8787",
		Handler: mux,
		TLSConfig: &tls.Config{
			ClientCAs:  pool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		},
	}

	if err := server.ListenAndServeTLS("./keys/ca/server.crt", "./keys/ca/server.key"); err != nil {
		log.Fatal(err)
	}

}
