// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
    "io/ioutil"
    "log"
    "os"
    "strings"

    "gopkg.in/alecthomas/kingpin.v2"
)

const (
    AUTHOR  = "Jia Jia"
    VERSION = "1.0.0"
)

const (
    PASS = "PASS"
    PORT = 25
    SENDER = "<mail@example.com>"
    SEP = ","
    SMTP = "smtp.example.com"
    USER = "USER"
)

var (
	contentTypeMap = map[string]string {
	    "HTML": "html",
        "PLAIN_TEXT": "text",
	}
)

var (
    app = kingpin.New("mailsender", "Mail sender written in Go").Author(AUTHOR).Version(VERSION)

    attachment = app.Flag("attachment", "Attachment, format: attach1,attach2,...").Short('a').String()
    body = app.Flag("body", "Body").Short('b').String()
    contentType = app.Flag("content_type", "Content type, format: HTML or PLAIN_TEXT (default)").
        Short('c').Default("PLAIN_TEXT").Enum("HTML", "PLAIN_TEXT")
    header = app.Flag("header", "Header").Short('e').String()
    recipients = app.Flag("recipients", "Recipients, format: alen@example.com,cc:catherine@example.com").Short('r').Required().String()
    title = app.Flag("title", "Title").Short('t').String()
)

func sendMail(attachment []string, body string, contentType string, header string, recipients string, title string) bool {
	return true
}

func parseContentType(data string) (string, bool) {
	buf, isPresent := contentTypeMap[data]
	if !isPresent {
	    return "", false
    }

    return buf, true
}

func parseBody(data string) (string, bool) {
    fi, err := os.Lstat(data)
    if err != nil || fi == nil || !fi.Mode().IsRegular() {
        return data, false
    }

    buf, err := ioutil.ReadFile(data)
    if err != nil {
        return data, false
    }

    return string(buf), true
}

func parseAttachment(data string) ([]string, bool) {
    var buf []string

	if len(data) == 0 {
		return buf, true
    }

	buf = strings.Split(data, ",")
	for _, item := range buf {
        fi, err := os.Lstat(item)
        if err != nil || fi == nil || !fi.Mode().IsRegular() {
            return nil, false
        }
    }

    return buf, true
}

func main() {
    kingpin.MustParse(app.Parse(os.Args[1:]))

    attachment, validAttachment := parseAttachment(*attachment)
    if !validAttachment {
        log.Fatal("Invalid attachment")
    }

    body, _ := parseBody(*body)

    contentType, validContentType := parseContentType(*contentType)
    if !validContentType {
        log.Fatal("Invalid content_type")
    }

    status := sendMail(attachment, body, contentType, *header, *recipients, *title)
    if status == false {
        log.Fatal("Failed to send mail")
    }

    os.Exit(0)
}
