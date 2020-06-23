package main

import (
	"fmt"
	COAP "github.com/dustin/go-coap"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var maxInt = 100

func handleA(l *net.UDPConn, a *net.UDPAddr, m *COAP.Message) *COAP.Message{
	log.Printf("Got message in handleA: path=%q: %#v from %v", m.Path(), m, a)

	if ! m.IsConfirmable() {
		s := m.Payload
		num := string(s)
		k, _ := strconv.Atoi(num)
		if k >= 1 && k <= 32767{
			maxInt = k
		}
	}

	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(maxInt)
	t := strconv.Itoa(i)

	if m.IsConfirmable(){

		res := &COAP.Message{
			Type:      COAP.Acknowledgement,
			Code:      COAP.Content,
			MessageID: m.MessageID,
			Token:     m.Token,
			Payload:   []byte(t),
		}

		res.SetOption(COAP.ContentFormat, COAP.TextPlain)
		log.Printf("Transmitting from A %#v", res)

		return res
	}

	return nil
}

func handleB(l *net.UDPConn, a *net.UDPAddr, m *COAP.Message) *COAP.Message{
	log.Printf("Got message in handleB: path=%q: %#v from %v", m.Path(), m, a)
	if m.IsConfirmable() {
		res := &COAP.Message{
			Type:      COAP.Acknowledgement,
			Code:      COAP.Content,
			MessageID: m.MessageID,
			Token:     m.Token,
			Payload:   []byte("pong"),
		}
		res.SetOption(COAP.ContentFormat, COAP.TextPlain)

		log.Printf("Transmitting from B %#v", res)
		return res
	}
	return nil
}

func main(){

	fmt.Println("Coap Service start ...")

	mux := COAP.NewServeMux()
	mux.Handle("/rand",COAP.FuncHandler(handleA))
	mux.Handle("/ping",COAP.FuncHandler(handleB))

	log.Fatal(COAP.ListenAndServe("udp","0.0.0.0:5683",mux))
}