package envelope

import (
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strconv"
	"strings"
	"unicode"
)

// "Content-Type":["application/MSWord; name=MN2R84RLB.doc"] (attachment)
// "Content-Type":["text/plain; charset=UTF-8"] (email text)
/*
 mechanism := "7bit" / "8bit" / "binary" /
                  "quoted-printable" / "base64" /
                  ietf-token / x-token
not case sensitive
default is 7bit unless specified
*/
type emailPart struct {
	MimeFileType    string
	Charset         string
	FileName        string
	Size            int
	CreationDate    string
	ModificatioDate string
	ReadDate        string
	Encoding        string // UTF-8, 8bit, Base64, quoted-printable
	Description     string
	IsAttachment    bool
	Data            []byte
}

func (p *emailPart) DataAsString() string {
	return string(p.Data)
}

func (p *emailPart) DataAsBytes() []byte {

	if p.Encoding == "base64" {
		bytesData, err := base64.StdEncoding.DecodeString(string(p.Data))
		if err != nil {
			return p.Data
		}
		return bytesData
	}

	if p.Encoding == "quoted-printable" || p.Encoding == "" {
		r := bytes.NewReader(p.Data)
		bodyReader := quotedprintable.NewReader(r) //base64.StdEncoding.DecodeString(string(p.Data))
		bytesData, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			return p.Data
		}
		return bytesData
	}

	return p.Data
}

func trim(s string) string {
	return strings.TrimLeftFunc(strings.TrimRightFunc(s, unicode.IsSpace), unicode.IsSpace)
}

func parseEmailParts(msg mail.Message) ([]emailPart, error) {
	var parts []emailPart
	var err error

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Could not parse the Content TYPE !!!")
		return parts, err
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			p, err := mr.NextPart()

			if err == io.EOF {
				return parts, nil
			}
			if err != nil {
				return parts, err
			}

			part := emailPart{}

			part.Data, err = ioutil.ReadAll(p)
			if err != nil {
				return parts, err
			}

			if disposition, ok := p.Header["Content-Disposition"]; ok && len(disposition) > 0 {
				if mediaType, params, err := mime.ParseMediaType(disposition[0]); err == nil {
					if strings.ToLower(mediaType) == "attachment" {
						//log.Println("Detected attachment", mediaType)
						part.IsAttachment = true
					}

					// size
					if size, ok := params["size"]; ok && len(size) > 0 {
						if sizeInt, err := strconv.Atoi(size); err == nil {
							part.Size = sizeInt
						}
					}

					// calculate the size otherwise by checking the length of the data
					if part.Size == 0 {
						part.Size = len(part.Data)
					}

					// creation-date
					if creationDate, ok := params["creation-date"]; ok && len(creationDate) > 0 {
						part.CreationDate = creationDate
					}

					// modification-date
					if modificationDate, ok := params["modification-date"]; ok && len(modificationDate) > 0 {
						part.ModificatioDate = modificationDate
					}

					// read-date
					if readDate, ok := params["read-date"]; ok && len(readDate) > 0 {
						part.ReadDate = readDate
					}

					// read-date
					if fileName, ok := params["filename"]; ok && len(fileName) > 0 {
						part.FileName = fileName
					}
				}

			}

			if contentType, ok := p.Header["Content-Type"]; ok && len(contentType) > 0 {
				// type
				if mediaType, params, err := mime.ParseMediaType(contentType[0]); err == nil {
					part.MimeFileType = mediaType

					// name
					if fileName, ok := params["name"]; ok && len(fileName) > 0 && part.FileName == "" {
						part.FileName = fileName
					}

					// charset
					if fileName, ok := params["charset"]; ok && len(fileName) > 0 && part.FileName == "" {
						part.Charset = fileName
					}
				}
			}

			if encoding, ok := p.Header["Content-Transfer-Encoding"]; ok && len(encoding) > 0 {
				part.Encoding = strings.ToLower(trim(encoding[0]))
			}

			if description, ok := p.Header["Content-Description"]; ok && len(description) > 0 {
				part.Description = trim(description[0])
			}

			// append to parts array
			parts = append(parts, part)
			//fmt.Printf("Part %q\n", part)
		}
	}

	return parts, nil

}
