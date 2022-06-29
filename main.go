package main

import (
	"errors"
	"log"
	"net"
	"net/smtp"
	"net/url"
	"time"

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

// establishProxyConnection connects to the address on the named network address
// via proxy protocol
func establishProxyConnection(addr, proxyURI string) (net.Conn, error) {
	// return socks.Dial(proxyURI)("tcp", addr)

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

	return dialfunc("tcp", addr)
}

// newSMTPClient generates a new available SMTP client
func newSMTPClient(domain, proxyURI string) (*smtp.Client, error) {
	domain = domainToASCII(domain)
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
	// Channel holding the new smtp.Client or error
	ch := make(chan interface{}, 1)

	// Dial the new smtp connection
	go func() {
		var conn net.Conn
		var err error

		conn, err = establishProxyConnection(addr, proxyURI)
		if err != nil {
			log.Println(err)
		}

		if err != nil {
			ch <- err
			return
		}

		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			log.Println(err)
		}

		client, err := smtp.NewClient(conn, host)
		log.Println(client)

		if err != nil {
			log.Println(err)

			ch <- err
			return
		}
		ch <- client
	}()

	// Retrieve the smtp client from our client channel or timeout
	select {
	case res := <-ch:
		switch r := res.(type) {
		case *smtp.Client:
			return r, nil
		case error:
			return nil, r
		default:
			return nil, errors.New("Unexpected response dialing SMTP server")
		}
	case <-time.After(smtpTimeout):
		return nil, errors.New("Timeout connecting to mail-exchanger")
	}
}

// domainToASCII converts any internationalized domain names to ASCII
// reference: https://en.wikipedia.org/wiki/Punycode
func domainToASCII(domain string) string {
	asciiDomain, err := idna.ToASCII(domain)
	if err != nil {
		return domain
	}
	return asciiDomain

}
