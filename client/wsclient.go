package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/pescaria/lib/trace"
	"golang.org/x/net/websocket"
)

type WsClinet struct {
	config  *websocket.Config
	conn    *websocket.Conn
	send    chan []byte
	receive chan []byte
}

func NewWsClient(config *websocket.Config) *WsClinet {
	return &WsClinet{
		config:  config,
		send:    make(chan []byte, 10),
		receive: make(chan []byte, 10),
	}
}

func (client *WsClinet) SendMsg(msg []byte) {
	client.send <- msg
}

func (client *WsClinet) ReceiveChan() chan []byte {
	return client.receive
}

func (client *WsClinet) runOnce() {
	trace.Trace.Debug("runOnce")
	defer client.conn.Close()
	defer close(client.receive)
	die := make(chan struct{})
	go func() {
		for {
			buf := make([]byte, 256)
			n, err := client.conn.Read(buf)
			if err != nil {
				trace.Trace.Warningf("read payload failed, ip:%v reason:%v size:%v", client.conn.RemoteAddr().String(), err, n)
				continue
			}
			client.receive <- buf[0:n]
		}
	}()
	for {
		select {
		case data := <-client.send:
			if data == nil {
				trace.Trace.Error("nil msg")
				break
			}
			/*
				err := websocket.Message.Send(client.conn, data)
				if err != nil {
					trace.Trace.Warningf("Error send  data, bytes: %v reason: %v", len(data), err)
				}
			*/

			n, err := client.conn.Write(data)
			if err != nil {
				trace.Trace.Warningf("Error send reply data, bytes: %v reason: %v", n, err)
			}

		case <-die:
			return
		}
	}

}

func (client *WsClinet) run() {
	nsleep := 100
	for {
		trace.Trace.Debugf("connecting to %s", client.config.Location.String())
		c, err := websocket.DialConfig(client.config)
		if err != nil {
			trace.Trace.Errorf("connet login error: %s", err)
			nsleep *= 2
			if nsleep > 60*1000 {
				nsleep = 60 * 1000
			}
			trace.Trace.Infof("connect again after %d Millisecond ", nsleep)
			time.Sleep(time.Duration(nsleep) * time.Millisecond)
			continue
		}

		c.PayloadType = websocket.BinaryFrame
		client.conn = c
		nsleep = 100
		client.runOnce()
	}
}

func (client *WsClinet) Start() {
	go client.run()
}

func initWebsocketWithTLS(url string) (*WsClinet, error) {
	config, err := websocket.NewConfig("wss://"+url, "https://"+url)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	cacrt, err := ioutil.ReadFile("../keys/ca/ca.crt")
	if err != nil {
		log.Fatal(err)
	}
	pool.AppendCertsFromPEM(cacrt)

	cliCrt, err := tls.LoadX509KeyPair("../keys/ca/client.crt", "../keys/ca/client.key")
	if err != nil {
		fmt.Println("Loadx509keypair err:", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		//InsecureSkipVerify: true,
		RootCAs:      pool,
		ServerName:   "localhost",
		Certificates: []tls.Certificate{cliCrt},
	}

	config.TlsConfig = tlsConfig

	return NewWsClient(config), nil
}

func initWebsocket(url string) (*WsClinet, error) {
	config, err := websocket.NewConfig("wss://"+url, "http://"+url)
	if err != nil {
		return nil, err
	}
	return NewWsClient(config), nil
}

func Init(ssl bool, url string) (*WsClinet, error) {
	if ssl {
		return initWebsocketWithTLS(url)
	}
	return initWebsocket(url)
}
