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

	"code.google.com/p/go-imap/go1/imap"
	"code.google.com/p/gopass"
)

var (
	login   = flag.String("login", "", "Your login")
	mailbox = flag.String("mailbox", "[Gmail]/All Mail", "The mailbox to watch")
	server  = flag.String("server", "imap.gmail.com", "Your IMAP server with TLS")
)

func main() {
	flag.Parse()
	if *login == "" {
		fmt.Println("I need at least your login!\n")
		flag.Usage()
		os.Exit(1)
	}

	pass, err := gopass.GetPass("Password: ")
	if err != nil {
		log.Fatal(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, os.Kill)
	go func() {
		<-quit
		os.Exit(0)
	}()

	for {
		c, err := imap.DialTLS(*server, &tls.Config{})
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.Auth(imap.PlainAuth(*login, pass, ""))
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Successfully authed")

		cmd, err := c.Select(*mailbox, true)
		if err != nil {
			log.Println("select")
			log.Fatal(err)
		}
		if _, err = cmd.Result(imap.OK); err != nil {
			log.Fatal(err)
		}

		log.Println("Successfully selected ", *mailbox)

		cmd, err = c.Idle()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Starting idle...")
		c.Data = nil

		for cmd.InProgress() {
			err := c.Recv(-1)
			if err != nil {
				if err != io.EOF {
					log.Println(err)
				}
				break
			}

			for _, rsp := range c.Data {
				if rsp.Label == "EXISTS" {
					log.Println("New message, running sync...")
					cmd := exec.Command("offlineimap", "-u", "Quiet")
					cmd.Stdout = os.Stderr
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						log.Printf("Error running sync: %s\n", err)
					}
					log.Println("Ran sync")
				}
			}

			c.Data = nil
		}
	}
}
