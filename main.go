package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"strconv"
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
	DownloadPath  string
}

func (c *Config) resolveUser(user *tb.User) string {
	var from string
	for nameduser, userid := range c.TelegramUsers {
		if userid == user.ID {
			from = nameduser
			break
		}
	}
	if from == "" {
		from = fmt.Sprintf("\"%s %s\" @%s <%d>", user.FirstName,
			user.LastName, user.Username, user.ID)
	}
	return from
}

func newConfig(filename string) (conf *Config, err error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	conf = new(Config)
	if err = yaml.Unmarshal(contents, &conf); err != nil {
		return
	}

	if conf.DownloadPath == "" {
		conf.DownloadPath = "."
	}
	var f *os.File
	fname := path.Join(conf.DownloadPath, strconv.Itoa(rand.Int()))
	if f, err = os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0600); err != nil {
		err = fmt.Errorf("downloadpath `%s` not writable: %s", conf.DownloadPath, err)
		return
	}
	f.Close()
	os.Remove(fname)

	return
}

var conf *Config

func xmppsend(c *xmpp.Client, msg string) {
	if _, err := c.Send(xmpp.Chat{Remote: conf.XMPPTarget, Type: "chat", Text: msg}); err != nil {
		log.Printf("xmpp send FAIL: %s\n", err)
	}
}

func hostFromSRV(XMPPUser string) (string, error) {
	parts := strings.Split(XMPPUser, "@")
	if len(parts) != 2 {
		return "", fmt.Errorf("xmppuser not a jabber ID: %s", XMPPUser)
	}
	_, addrs, err := net.LookupSRV("xmpp-client", "tcp", parts[1])
	if err != nil {
		return "", err
	}
	// just picking the first srv record
	host, port := addrs[0].Target, addrs[0].Port
	if host == "" || port <= 0 {
		return "", fmt.Errorf("bad SRV record: %s:%d", host, port)
	}
	hostPort := net.JoinHostPort(strings.TrimSuffix(host, "."), strconv.Itoa(int(port)))
	return hostPort, nil
}

func main() {
	var xmppc *xmpp.Client
	var err error

	rand.Seed(time.Now().UnixNano())

	conf, err = newConfig("quiteabot.yaml")
	if err != nil {
		fmt.Printf("config failed: %s\n", err)
		os.Exit(1)
	}

	host := conf.XMPPServer
	if host == "" {
		host, err = hostFromSRV(conf.XMPPUser)
		if err != nil {
			panic(err)
		}
	}

	options := xmpp.Options{
		Host:          host,
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
		from := conf.resolveUser(m.Sender)

		log.Printf("%s -> %s\n", from, conf.XMPPTarget)
		if conf.Verbose {
			fmt.Printf(">%s\n", m.Text)
		}
		xmppsend(xmppc, fmt.Sprintf("%s> %s", from, m.Text))
	})

	telec.Handle(tb.OnPhoto, func(m *tb.Message) {
		from := conf.resolveUser(m.Sender)

		// TODO? should send xmpp about errors?
		if m.Photo.FileID == "" {
			log.Printf("Error: from: %s, photo but empty fileid!\n", from)
			return
		}

		file, err := telec.FileByID(m.Photo.FileID)
		if err != nil {
			log.Printf("Error: from: %s, FileByID(%v): %v\n", from, m.Photo.FileID, err)
			return
		}

		// photos are always jpg it seems
		fpath := path.Join(conf.DownloadPath,
			fmt.Sprintf("from_%s_%s.jpg", from, m.Time().Format("20060102T150405")))

		if err = telec.Download(&file, fpath); err != nil {
			log.Printf("Error: from: %s, Download(%v,...): %v\n", from, file, err)
			return
		}

		log.Printf("%s: downloaded photo: %s\n", from, fpath)
		if conf.Verbose {
			fmt.Printf(">[caption: %s]\n", m.Caption)
		}
		xmppsend(xmppc, fmt.Sprintf("%s> [downloaded photo: %s caption: %s]",
			from, fpath, m.Caption))
	})

	telec.Handle(tb.OnDocument, func(m *tb.Message) {
		from := conf.resolveUser(m.Sender)

		// TODO? should send xmpp about errors?
		if m.Document.FileID == "" {
			log.Printf("Error: from: %s, document but empty fileid!\n", from)
			return
		}

		file, err := telec.FileByID(m.Document.FileID)
		if err != nil {
			log.Printf("Error: from: %s, FileByID(%v): %v\n", from, m.Document.FileID, err)
			return
		}

		fpath := path.Join(conf.DownloadPath,
			fmt.Sprintf("from_%s_%s_%s", from, m.Time().Format("20060102T150405"), m.Document.FileName))

		if err = telec.Download(&file, fpath); err != nil {
			log.Printf("Error: from: %s, Download(%v,...): %v\n", from, file, err)
			return
		}

		log.Printf("%s: downloaded document (%s): %s\n", from, m.Document.MIME, fpath)
		if conf.Verbose {
			fmt.Printf(">[mime: %s, caption: %s]\n", m.Document.MIME, m.Caption)
		}
		xmppsend(xmppc, fmt.Sprintf("%s> [downloaded document: %s mime: %s caption: %s]",
			from, fpath, m.Document.MIME, m.Caption))
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
				if _, err := telec.Send(&tb.User{ID: userid}, usermsg[1], tb.NoPreview); err != nil {
					log.Printf("teleg send FAIL: %s\n", err)
				}
				log.Printf("%s -> %s <%d>\n", v.Remote, usermsg[0], userid)
				if conf.Verbose {
					fmt.Printf(">%s\n", usermsg[1])
				}
			}
		}
	}()

	telec.Start()
}
