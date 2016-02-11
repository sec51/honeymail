### SMTP honeypot

STATUS: work in progress not ready for prime time.

### Acknowledgments

Part of the code is taken from the user: [Chris Siebenmann](https://github.com/siebenmann) specifically the:

1. command.go
2. config.go
3. limits.go

files have been copied from his [smtpd](https://github.com/siebenmann/smtpd) code base.

The `command.go` file has been slightly changed to accomodate the additional default response the server needs to send to the mail client.

### License

GPL v3 for now