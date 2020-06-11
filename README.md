# go-mta

[![GoDoc](https://godoc.org/github.com/HuguesGuilleus/go-mta/pkg?status.svg)](https://godoc.org/github.com/HuguesGuilleus/go-mta/pkg)

A simple Mail transfert Agent. You can use as a package or as a deamon.

## Deamon

You can build yourself or use the [Compiled binary](https://github.com/HuguesGuilleus/go-mta/releases/latest "Github Release")

### Configuration file

The configuration file is in `/etc/go-mta/config.ini` or it's the argument when you run the program.

```ini
; The file who contains login
login = /etc/login.txt
; The listen address
addrs = localhost:25 :446
; Log output directory, by default /var/etc/go-mta/
out = log/

; Every host has a section.
[host1]
dkim_key = /path/to/dkim/private/key.pem
dkim_selector = DKIM slector
crt = /path/to/tls/certificate.pem
key = /path/to/tls/private/key.pem
```

### Login file

Each line contain a login, separation blank(s) and a password. The login and the password can't contain blank or `#`. You can add some comments with `#` and blank lines.

### DKIM key generation

You must generate DKIM key you can do this:

```bash
openssl genrsa -out key.pem
openssl rsa -in key.pem -pubout -out pub.key
```

## TODO

-   [ ] StartTLS
