package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/rakoo/go-ini"

	"code.google.com/p/go-imap/go1/imap"
)

var (
	mailbox = flag.String("mailbox", "[Gmail]/All Mail", "The mailbox to watch")
)

func main() {
	flag.Parse()
	email, password := getCredentials()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	go func() {
		<-quit
		os.Exit(0)
	}()

	connect := func(server string) (c *imap.Client, err error) {
		c, err = imap.DialTLS(server, &tls.Config{})
		if err != nil {
			return
		}

		_, err = c.Auth(imap.PlainAuth(email, password, ""))
		if err != nil {
			return
		}

		log.Println("Successfully authed")

		cmd, err := c.Select(*mailbox, true)
		if err != nil {
			log.Println("Error selecting mailbox: ", err)
			return
		}
		_, err = cmd.Result(imap.OK)
		if err != nil {
			return
		}

		log.Println("Successfully selected ", *mailbox)

		_, err = c.Idle()
		if err != nil {
			return
		}

		log.Println("Starting idle...")
		c.Data = nil

		return
	}

loop:
	for {
		if err := cmd(); err != nil {
			log.Printf("Error running sync on new loop: %s\n", err)
			continue
		}

		c, err := connect("imap.gmail.com")
		if err != nil {
			continue
		}

		wait := true
		for wait {
			err = c.Recv(15 * time.Minute)
			switch err {
			case nil:
				fallthrough
			case io.EOF:
				// We received content from server -- sync mails
				wait = false
			case imap.ErrTimeout:
				// after the timeout, wakeup the connection
				if _, err := c.IdleTerm(); err != nil {
					log.Println("Error finishing idle:: ", err)
					continue loop
				}
				if _, err := imap.Wait(c.Noop()); err != nil {
					log.Println("Error nooping: ", err)
					continue loop
				}
				log.Println("Keeping it alive !")
				if _, err := c.Idle(); err != nil {
					log.Println("Error re-idling: ", err)
					continue loop
				}
			default:
				log.Println("Error while receiving content: ", err)
				continue loop
			}
		}

		for _, rsp := range c.Data {
			if rsp.Label == "EXISTS" {
				log.Println("New message, running sync...")
				if err := cmd(); err != nil {
					log.Printf("Error running sync: %s\n", err)
				}
				log.Println("Ran sync")
			}
		}

		c.Data = nil
	}
}

func getCredentials() (email, password string) {
	cfg, err := ini.LoadFile(filepath.Join(os.Getenv("HOME"),
		".offlineimaprc"))
	if err != nil {
		log.Fatal(err)
	}

	accounts, ok := cfg.Get("general", "accounts")
	if !ok {
		log.Fatal("No general/accounts string")
	}

	// Only take first one
	account := strings.Split(accounts, ",")[0]
	remoterepo, ok := cfg.Get("Account "+account, "remoterepository")
	if !ok {
		log.Fatal(fmt.Sprintf("No section 'Account %s'/remoterepository",
			account))
	}

	typ, ok := cfg.Get("Repository "+remoterepo, "type")
	if !ok {
		log.Fatal("No type in repo")
	}
	if typ != "Gmail" {
		log.Fatal("I can only manage Gmail repos; curren type is ", typ)
	}
	user, okuser := cfg.Get("Repository "+remoterepo, "remoteuser")
	password, okpass := cfg.Get("Repository "+remoterepo, "remotepass")
	if !okuser || !okpass {
		log.Fatal("Invalid config")
	}

	return user + "@gmail.com", password
}

func cmd() error {
	cmd := exec.Command("offlineimap", "-u", "Quiet")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
