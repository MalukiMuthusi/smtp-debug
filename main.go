package main

import (
	"bufio"
	"errors"
	"log"
	"net"
	"net/url"
	"time"

	smtp "github.com/emersion/go-smtp"
	"golang.org/x/net/idna"
	"golang.org/x/net/proxy"
)

const (
	smtpTimeout = time.Second * 60
	smtpPort    = ":25"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {

	// The code works when I use a localhost proxy
	//   	socks5://dante:maluki@127.0.0.1:1080

	client, err := newSMTPClient("gmail.com", "socks5://dante:maluki@35.242.186.23:1080")

	if err != nil {
		log.Println(err)
		return
	}

	log.Println(client)
}

// newSMTPClient generates a new available SMTP client
func newSMTPClient(domain, proxyURI string) (*smtp.Client, error) {

	domain, err := idna.ToASCII(domain)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if len(mxRecords) == 0 {
		return nil, errors.New("No MX records found")
	}

	// Attempt to connect to SMTP servers
	for _, r := range mxRecords {

		// Simplified to make the code short

		addr := r.Host + smtpPort

		c, err := dialSMTP(addr, proxyURI)
		if err != nil {
			log.Println(err)
			continue
		}

		return c, err

	}

	return nil, errors.New("failed to created smtp.Client")
}

// dialSMTP is a timeout wrapper for smtp.Dial. It attempts to dial an
// SMTP server (socks5 proxy supported) and fails with a timeout if timeout is reached while
// attempting to establish a new connection
func dialSMTP(addr, proxyURI string) (*smtp.Client, error) {

	// Dial the new smtp connection

	u, err := url.Parse(proxyURI)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	var iface proxy.Dialer

	if u.User != nil {
		auth := proxy.Auth{}

		auth.User = u.User.Username()
		auth.Password, _ = u.User.Password()

		iface, err = proxy.SOCKS5("tcp", u.Host, &auth, &net.Dialer{Timeout: 30 * time.Second})
		if err != nil {
			log.Println(err)

			return nil, err
		}
	} else {
		iface, err = proxy.SOCKS5("tcp", u.Host, nil, proxy.FromEnvironment())
		if err != nil {
			log.Println(err)

			return nil, err
		}
	}

	dialfunc := iface.Dial

	netConn, err := dialfunc("tcp", addr)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		log.Println(err)

		return nil, err
	}

	buf := bufio.NewReader(netConn)
	bytes, err := buf.ReadBytes('\n')
	if err != nil {

		log.Println(err)
		return nil, err
	}
	log.Printf("%s\n", bytes)

	client, err := smtp.NewClient(netConn, host)

	if err != nil {
		log.Println(err)

		return nil, err
	}

	return client, nil
}
