package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	xmpp "github.com/mattn/go-xmpp"
	tb "gopkg.in/tucnak/telebot.v2"
)

var server = "picapica.lublin.se:5222"
var username = "quiteabot@lublin.se"
var password = "zDhN9F2EiYm79HZOpZFDP594V"
var status = "xa"
var statusMessage = "woot"
var debug = false
var session = false //not server session

func serverName(host string) string {
	return strings.Split(host, ":")[0]
}

func main() {
	var xmppc *xmpp.Client
	var err error
	options := xmpp.Options{
		Host:          server,
		User:          username,
		Password:      password,
		NoTLS:         true,
		StartTLS:      true,
		Debug:         debug,
		Session:       session,
		Status:        status,
		StatusMessage: statusMessage,
	}

	xmppc, err = options.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	token, err := ioutil.ReadFile("quiteabot.telegramtoken")
	if err != nil {
		panic(err)
	}
	tbc, err := tb.NewBot(tb.Settings{
		Token: strings.TrimSpace(string(token)),
		// // You can also set custom API URL. If field is empty it equals to "https://api.telegram.org"
		// URL:    "http://195.129.111.17:8012",
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	tbc.Handle(tb.OnText, func(m *tb.Message) {
		// 780111986 me
		msg := fmt.Sprintf("%s%s<%d>: %s", m.Sender.FirstName, m.Sender.LastName, m.Sender.ID, m.Text)
		fmt.Printf("---\nfrom: %s\n", msg)
		xmppc.Send(xmpp.Chat{Remote: "daniel@lublin.se", Type: "chat", Text: msg})
		fmt.Printf("relayed to xmpp\n")
	})

	go func() {
		for {
			chat, err := xmppc.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Chat:
				fmt.Printf("---\nfrom: %s\n", v.Remote)
				if !strings.HasPrefix(v.Remote, "daniel@lublin.se") {
					fmt.Printf("ignored\n")
					continue
				}
				usermsg := strings.SplitN(v.Text, ":", 2)
				if len(usermsg) < 2 {
					xmppc.Send(xmpp.Chat{Remote: "daniel@lublin.se", Type: "chat", Text: "expected: user:the msg"})
					fmt.Printf("wrong format: %s\n", v.Text)
					continue
				}
				users := map[string]int{"me": 780111986}
				userid := users[usermsg[0]]
				if userid == 0 || len(usermsg[1]) == 0 {
					xmppc.Send(xmpp.Chat{Remote: "daniel@lublin.se", Type: "chat", Text: "unknown user"})
					fmt.Printf("unknown user: %s (or empty msg)\n", usermsg[0])
					continue
				}
				tbc.Send(&tb.User{ID: userid}, usermsg[1], tb.NoPreview)
				fmt.Printf("relayed to teleg\n")
				// case xmpp.Presence:
				// 	fmt.Println(v.From, v.Show)
			}
		}
	}()

	tbc.Start()
}
