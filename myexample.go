package main

import (
	"bufio"
	"fmt"
	"flag"
	"github.com/mattn/go-xmpp"
	"log"
	"strings"
	"encoding/json"
	"os"
	"os/exec"
)

var server   = flag.String("server", "talk.google.com:5223", "server")
var username = flag.String("username", "", "username")
var password = flag.String("password", "", "password")
var notls = flag.Bool("notls", false, "No TLS")
var talk *xmpp.Client

// XXX: Allow only me to talk to this bot.
var me string

func Parsecmd(chat_txt string) {
	fmt.Println("parsecmd:", chat_txt)
	var line string
	// First token is the cmd. Second token is the value/arg.
	token := strings.SplitN(chat_txt, " ", 2)
	
	// "Driver" returns a JSON formatted string
	var trane struct {
		Zwavestr string
		Temp int
	}

	switch token[0] {

//	case default:
	//	talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: line})

	case "Gettemp", "Get temp", "gettemp":
		fmt.Printf("cmd: %s\n", chat_txt)
		//XXX: Needs a better PATH
		getcmd := exec.Command("/opt/github/open-zwave/cpp/examples/linux/MinOZW/test", "-d", "/dev/ttyUSB0", "-g")
		stdout, err := getcmd.StdoutPipe()
	//	fmt.Println("Get current temp\n")
		
		if err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}

		if err := getcmd.Start(); err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}
		
		if err := json.NewDecoder(stdout).Decode(&trane); err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}

		if err := getcmd.Wait(); err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}
	
		line = fmt.Sprintf("%s is %d\n", trane.Zwavestr, trane.Temp)
		fmt.Println(line)
		talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: line})

	case "Settemp":
		fmt.Printf("cmd: %s\n", chat_txt)
		fmt.Printf("Setting new temp to: %s\n", token[1])
		setcmd := exec.Command("/opt/github/open-zwave/cpp/examples/linux/MinOZW/test", "-d", "/dev/ttyUSB0", "-s", token[1])
		stdout, err := setcmd.StdoutPipe()
		if err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}

		if err := setcmd.Start(); err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}
		
		if err := json.NewDecoder(stdout).Decode(&trane); err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}

		if err := setcmd.Wait(); err != nil {
			talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: err.Error()})
			log.Fatal(err)
		}
		line = fmt.Sprintf("%s is %d\n", trane.Zwavestr, trane.Temp)
		fmt.Println(line)
		talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: line})
	}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: example [options]\n")
		flag.PrintDefaults()
		os.Exit(2)
	}

	flag.Parse()

	if *username == "" || *password == "" {
		flag.Usage()
	}

	var err error

	if *notls {
		talk, err = xmpp.NewClientNoTLS(*server, *username, *password)
	} else {
		talk, err = xmpp.NewClient(*server, *username, *password)
	}
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			chat, err := talk.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Chat:
			//	fmt.Println(v.Remote, v.Text)
				fmt.Println("-->")
				if v.Text != "" {
					Parsecmd(v.Text)
				}
				
			case xmpp.Presence:
			//	fmt.Println(v.From, v.Show)
				fmt.Println("<--")
				//XXX: Improve this in the future to really allow only me. Otherwise
				// me = first person to come online.
				me = v.From
			}
		}
	}()

	for {
		in := bufio.NewReader(os.Stdin)
		line, err := in.ReadString('\n')
		if err != nil {
			continue
		}
		line = strings.TrimRight(line, "\n")

		fmt.Println(me)
		// Anything you type on the cmdline after this app runs will be sent to "me".
		talk.Send(xmpp.Chat{Remote: me, Type: "chat", Text: line})
	}
}
