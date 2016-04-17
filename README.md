[![Build Status](https://travis-ci.org/sec51/honeymail.svg?branch=master)](https://travis-ci.org/sec51/honeymail)

### SMTP honeypot

UPDATE: the project is being actively developed. Here a list of some of the features which have been implemented, but not pushed to master yet.

- [x] Automatically extract several information from the email, like: list of urls, source domain, country, attachments, email parts (HTML or TXT) and more
- [x] Email storage to BOLTDB, with the possibility to add different storage backends (by implementing the storage interface)
- [x] API to retrieve the stored emails in a JSON format for further analysis or to display it in a web UI

### Status

**Work in progress.**
This honeypot has been tested by **Sec51** however **we cannot guarantee that it's bug free!**
Attackers may be able to gain access to your honeypot in case of severe bug. **Use at your own risk !**
We are not responsible for any damage caused by this software. For more information see the license.

### How to run it:

1) Generate a public/private key via:

`openssl req -newkey rsa:2048 -nodes -keyout smtp.key -x509 -days 365 -out smtp.crt`

2) Move the newly created certificates to a `cert` folder.

3) Configure your remote ip address or ip address list in the `conf/development.conf` or `conf/production.conf` INI config file.
This will allow only your IP to connect to the API. In addition set the path of the certificates.

4) Run the binary via:

`setcap 'cap_net_bind_service=+ep' honeymail`

5) Access the api via:

To see today's emails:

- `/api/emails`

To see a spefici email (you can find the id from the list return from /api/emails):

- `/api/email?id=49689cfcb7fcbf83ed95df3a65ae6d9047678ca1`

Please report any bugs you will encounter.

### Dependencies

The project is now using go vendoring. So all dependencies are inside the `vendor` folder

### License

Copyright (c) 2016 Sec51.com <info@sec51.com>

Permission to use, copy, modify, and distribute this software for any
purpose with or without fee is hereby granted, provided that the above 
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE. 