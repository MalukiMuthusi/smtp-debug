# SMTP DEBUG

I am trying to create an SMTP client that uses a proxy connection, SOCKS5.

When I use a local host proxy the code successfully creates an SMTP client.

When I try to use a remote proxy, I am getting an `TTL expired`. I am also getting a `EOF` error when try to use a different proxy connection.

## Proxy

I have set up a proxy server in my localhost, `socks5://dante:maluki@127.0.0.1:1080`

I have also set up an identical proxy server on my remote VM, `socks5://dante:maluki@35.242.186.23:1080`
