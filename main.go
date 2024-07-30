package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	g "xabbo.b7c.io/goearth"
	"xabbo.b7c.io/goearth/shockwave/in"
	"xabbo.b7c.io/goearth/shockwave/out"
)

var ext = g.NewExt(g.ExtInfo{
	Title:       "Tradesound",
	Description: "An extension that plays a sound when you trade others, or others trade you",
	Author:      "Loke",
	Version:     "1.0.0",
})

var (
	username  string
	selftrade bool
)

func main() {
	ext.Activated(func() {
		log.Printf("Extension activated")
		if ext.IsConnected() {
			go getUserInfo()
		} else {
			log.Printf("Game is not connected")
		}
	})
	ext.Intercept(in.TRADE_ITEMS).With(handleTrade)
	ext.Intercept(out.CHAT, out.SHOUT, out.WHISPER).With(onChatMessage)
	ext.Run()
}

func onChatMessage(e *g.Intercept) {
	msg := e.Packet.ReadString()
	if strings.HasPrefix(msg, ":") {
		command := strings.TrimPrefix(msg, ":")
		switch {
		case strings.HasSuffix(command, "selftrade"):
			e.Block()
			selftrade = !selftrade
			msg = ""
			if selftrade {
				msg = "Playing sounds when other trade you, or you trade others"
			} else {
				msg = "Playing sound only when others trade you"
			}
			ext.Send(in.CHAT, 0, msg, 0, 34, 0, 0)
		}
	}
}

func extractName(input string) string {
	str := input[1:]
	re := regexp.MustCompile(`^[^\\]*`)
	match := re.FindStringSubmatch(str)
	return match[0]
}

func getUserInfo() {
	log.Printf("Retrieving user info...")
	ext.Send(out.INFORETRIEVE)
	if pkt := ext.Recv(in.USER_OBJ).Block().Wait(); pkt != nil {
		var id g.Id
		var name string
		pkt.Read(&id, &name)
		msg1 := fmt.Sprintf("%q", name)
		username = extractName(msg1)
	} else {
		log.Printf("Timed out.")
	}
}

func handleTrade(e *g.Intercept) {
	tradeInitiator := e.Packet.ReadString()
	if username == "" {
		return
	}

	if tradeInitiator == username && !selftrade {
		return
	}

	f, err := os.Open("trade.mp3")
	if err != nil {
		log.Fatal(err)
	}
	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	defer streamer.Close()
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
