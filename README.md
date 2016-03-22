[![Build Status](https://travis-ci.org/sec51/honeymail.svg?branch=master)](https://travis-ci.org/sec51/honeymail)

### SMTP honeypot

UPDATE: the project is being actively developed. Here a list of some of the features which have been implemented, but not pushed to master yet.

- [x] Automatically extract several information from the email, like: list of urls, source domain, country, attachments, email parts (HTML or TXT) and more
- [x] Email storage to BOLTDB, with the possibility to add different storage backends (by implementing the storage interface)
- [x] API to retrieve the stored emails in a JSON format for further analysis or to display it in a web UI

### Status

work in progress - NOT ready for prime time.

### Dependencies

```
  - go get github.com/Sirupsen/logrus  
  - go get github.com/boltdb/bolt/...
  - go get github.com/oschwald/geoip2-golang
  - go get github.com/mvdan/xurls
```

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