package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	xmpp "github.com/mattn/go-xmpp"
	tb "gopkg.in/tucnak/telebot.v2"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	XMPPServer    string
	XMPPUser      string
	XMPPPass      string
	XMPPTarget    string
	TelegramToken string
	TelegramUsers map[string]int
}

func parseconfig(filename string) (conf *Config, err error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	conf = new(Config)
	if err = yaml.Unmarshal(contents, &conf); err != nil {
		return
	}
	return
}

func main() {
	var xmppc *xmpp.Client
	var err error

	conf, err := parseconfig("quiteabot.yaml")
	if err != nil {
		panic(err)
	}

	options := xmpp.Options{
		Host:          conf.XMPPServer,
		User:          conf.XMPPUser,
		Password:      conf.XMPPPass,
		NoTLS:         true,
		StartTLS:      true,
		Debug:         false,
		Session:       false, // no server session
		Status:        "xa",
		StatusMessage: "i'm a bot",
	}

	xmppc, err = options.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	tbc, err := tb.NewBot(tb.Settings{
		Token: conf.TelegramToken,
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
		xmppc.Send(xmpp.Chat{Remote: conf.XMPPTarget, Type: "chat", Text: msg})
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
				if !strings.HasPrefix(v.Remote, conf.XMPPTarget) {
					fmt.Printf("ignored\n")
					continue
				}
				usermsg := strings.SplitN(v.Text, ":", 2)
				if len(usermsg) < 2 {
					xmppc.Send(xmpp.Chat{Remote: conf.XMPPTarget, Type: "chat", Text: "expected: user:the msg"})
					fmt.Printf("wrong format: %s\n", v.Text)
					continue
				}
				userid := conf.TelegramUsers[usermsg[0]]
				if userid == 0 || len(usermsg[1]) == 0 {
					xmppc.Send(xmpp.Chat{Remote: conf.XMPPTarget, Type: "chat", Text: "unknown user"})
					fmt.Printf("unknown user: %s (or empty msg)\n", usermsg[0])
					continue
				}
				tbc.Send(&tb.User{ID: userid}, usermsg[1], tb.NoPreview)
				fmt.Printf("relayed to <%d>\n", userid)
				// case xmpp.Presence:
				// 	fmt.Println(v.From, v.Show)
			}
		}
	}()

	tbc.Start()
}
