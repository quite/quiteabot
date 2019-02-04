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
	Verbose       bool
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

var conf *Config

func xmppsend(c *xmpp.Client, msg string) {
	if _, err := c.Send(xmpp.Chat{Remote: conf.XMPPTarget, Type: "chat", Text: msg}); err != nil {
		log.Println(err)
	}
}

func main() {
	var xmppc *xmpp.Client
	var err error

	conf, err = parseconfig("quiteabot.yaml")
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
	log.Println("xmpp: connected")

	telec, err := tb.NewBot(tb.Settings{
		Token:  conf.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("telegram: connected")

	telec.Handle(tb.OnText, func(m *tb.Message) {
		var from string
		for user, userid := range conf.TelegramUsers {
			if userid == m.Sender.ID {
				from = user
				break
			}
		}
		if from == "" {
			from = fmt.Sprintf("\"%s %s\" @%s <%d>", m.Sender.FirstName,
				m.Sender.LastName, m.Sender.Username, m.Sender.ID)
		}
		log.Printf("%s -> %s\n", from, conf.XMPPTarget)
		if conf.Verbose {
			fmt.Printf(">%s\n", m.Text)
		}
		xmppsend(xmppc, fmt.Sprintf("%s: %s", from, m.Text))
	})

	go func() {
		for {
			chat, err := xmppc.Recv()
			if err != nil {
				log.Fatal(err)
			}
			switch v := chat.(type) {
			case xmpp.Chat:
				// only care about msgs from our xmpptarget
				if !strings.HasPrefix(v.Remote, conf.XMPPTarget) {
					log.Printf("%s : ignored\n", v.Remote)
					if conf.Verbose {
						fmt.Printf(">%s\n", v.Text)
					}
					continue
				}
				usermsg := strings.SplitN(v.Text, ":", 2)
				if len(usermsg) < 2 {
					msg := "expected format: user:the msg"
					log.Printf("%s : %s", v.Remote, msg)
					if conf.Verbose {
						fmt.Printf(">%s\n", v.Text)
					}
					xmppsend(xmppc, msg)
					continue
				}
				userid := conf.TelegramUsers[usermsg[0]]
				if userid == 0 || usermsg[1] == "" {
					msg := "unlisted user or empty msg"
					log.Printf("%s : %s", v.Remote, msg)
					if conf.Verbose {
						fmt.Printf(">%s\n", v.Text)
					}
					xmppsend(xmppc, msg)
					continue
				}
				telec.Send(&tb.User{ID: userid}, usermsg[1], tb.NoPreview)
				log.Printf("%s -> %s <%d>\n", v.Remote, usermsg[0], userid)
				if conf.Verbose {
					fmt.Printf(">%s\n", usermsg[1])
				}
			}
		}
	}()

	telec.Start()
}
