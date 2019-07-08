package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/v-braun/go2p"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func main() {
	logrus.SetFormatter(new(prefixed.TextFormatter))
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	localAddr := flag.String("laddr", "localhost:7071", "local ip address")
	flag.Parse()

	cyan := color.New(color.FgCyan).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	green := color.New(color.FgHiGreen).SprintFunc()
	white := color.New(color.FgHiWhite).SprintFunc()
	peerName := color.New(color.BgBlue, color.FgHiWhite).SprintFunc()

	net := go2p.NewNetworkConnectionTCP(*localAddr, &map[string]func(peer *go2p.Peer, msg *go2p.Message){
		"msg": func(peer *go2p.Peer, msg *go2p.Message) {
			fmt.Println(fmt.Sprintf("%s %s", peerName(peer.RemoteAddress()+" > "), msg.PayloadGetString()))
		},
	})

	err := net.Start()
	if err != nil {
		panic(err)
	}

	net.OnPeer(func(p *go2p.Peer) {
		fmt.Printf("%s %s\n", cyan("new peer:"), green(p.RemoteAddress()))
	})

	defer net.Stop()

	fmt.Println(cyan(`
local server started!

press:
	`))

	fmt.Printf("%s %s\n", blue("[q][ENTER]"), white("to exit"))
	fmt.Printf("%s %s\n", blue("[c {ip address : port}][ENTER]"), white("to connect to another peer"))
	fmt.Printf("%s %s\n", blue("[any message][ENTER]"), white("to send a message to all peers"))

	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "q" {
			return
		} else if strings.HasPrefix(text, "c ") {
			text = strings.TrimPrefix(text, "c ")
			net.ConnectTo("tcp", text)
		} else {
			net.SendBroadcast(go2p.NewMessageRoutedFromString("msg", text))
		}
	}

}
