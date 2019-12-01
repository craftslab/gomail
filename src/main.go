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
    "path/filepath"
    "strings"

    "gopkg.in/alecthomas/kingpin.v2"
    "gopkg.in/gomail.v2"
)

const (
    author  = "Jia Jia"
    version = "1.0.0"
)

const (
    host = "smtp.example.com"
    pass = "pass"
    port = 25
    sender = "<mail@example.com>"
    sep = ","
    user = "user"
)

var (
    contentTypeMap = map[string]string {
        "HTML": "text/html",
        "PLAIN_TEXT": "text/plain",
    }
)

var (
    app = kingpin.New("mailsender", "Mail sender written in Go").Author(author).Version(version)

    attachment = app.Flag("attachment", "Attachment, format: attach1,attach2,...").Short('a').String()
    body = app.Flag("body", "Body").Short('b').String()
    contentType = app.Flag("content_type", "Content type, format: HTML or PLAIN_TEXT (default)").
        Short('c').Default("PLAIN_TEXT").Enum("HTML", "PLAIN_TEXT")
    header = app.Flag("header", "Header").Short('e').String()
    recipients = app.Flag("recipients", "Recipients, format: alen@example.com,cc:bob@example.com").Short('r').Required().String()
    title = app.Flag("title", "Title").Short('t').String()
)

func checkFile(name string) (string, bool) {
    fi, err := os.Lstat(name)
    if err != nil {
        root, _ := os.Getwd()
        fullname := filepath.Join(root, name)
        fi, err = os.Lstat(fullname)
        if err != nil {
            return name, false
        }
        name = fullname
    }

    if fi == nil || !fi.Mode().IsRegular() {
        return name, false
    }

    return name, true
}

func sendMail(from string, to []string, cc []string, subject string, contentType string, body string, attachment []string) bool {
    msg := gomail.NewMessage()

    msg.SetAddressHeader("From", sender, from)
    msg.SetHeader("To", strings.Join(to, sep))
    msg.SetHeader("Cc", strings.Join(cc, sep))
    msg.SetHeader("Subject", subject)
    msg.SetBody(contentType, body)

    for _, item := range attachment {
        msg.Attach(item, gomail.Rename(filepath.Base(item)))
    }

    dialer := gomail.NewDialer(host, port, user, pass)

    if err := dialer.DialAndSend(msg); err != nil {
        return false
    }

    return true
}

func parseRecipients(data string) ([]string, []string) {
    var cc []string
    var to []string

    buf := strings.Split(data, sep)
    for _, item := range buf {
        if hasPrefix := strings.HasPrefix(item, "cc:"); hasPrefix {
            cc = append(cc, strings.ReplaceAll(item, "cc:", ""))
        } else {
            to = append(to, item)
        }
    }

    return cc, to
}

func parseContentType(data string) (string, bool) {
    buf, isPresent := contentTypeMap[data]
    if !isPresent {
        return "", false
    }

    return buf, true
}

func parseBody(name string) (string, bool) {
    name, status := checkFile(name)
    if !status {
        return name, false
    }

    buf, err := ioutil.ReadFile(name)
    if err != nil {
        return name, false
    }

    return string(buf), true
}

func parseAttachment(name string) ([]string, bool) {
    var names []string
    var status bool

    if len(name) == 0 {
        return names, true
    }

    names = strings.Split(name, sep)
    for i := 0; i < len(names); i++ {
        names[i], status = checkFile(names[i])
        if !status {
            return nil, false
        }
    }

    return names, true
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

    to, cc := parseRecipients(*recipients)

    status := sendMail(*header, to, cc, *title, contentType, body, attachment)
    if !status {
        log.Fatal("Failed to send mail")
    }

    os.Exit(0)
}
