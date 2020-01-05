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
    "encoding/json"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/pkg/errors"
    "gopkg.in/alecthomas/kingpin.v2"
    "gopkg.in/gomail.v2"
)

const (
    author  = "Jia Jia"
    version = "2.0.0"
)

type Config struct {
    Host string `json:"host"`
    Pass string `json:"pass"`
    Port int `json:"port"`
    Sender string `json:"sender"`
    Sep string `json:"sep"`
    User string `json:"user"`
}

type Mail struct {
    Attachment []string
    Body string
    Cc []string
    ContentType string
    From string
    Subject string
    To []string
}

var (
    contentTypeMap = map[string]string {
        "HTML": "text/html",
        "PLAIN_TEXT": "text/plain",
    }
)

var (
    app = kingpin.New("mailsender", "Mail sender written in Go").Author(author).Version(version)

    attachment = app.Flag("attachment", "Attachment files, format: attach1,attach2,...").Short('a').String()
    body = app.Flag("body", "Body text or file").Short('b').String()
    config = app.Flag("config", "Config file, format: .json").Short('c').String()
    contentType = app.Flag("content_type", "Content type, format: HTML or PLAIN_TEXT (default)").
        Short('e').Default("PLAIN_TEXT").Enum("HTML", "PLAIN_TEXT")
    header = app.Flag("header", "Header text").Short('r').String()
    recipients = app.Flag("recipients", "Recipients list, format: alen@example.com,cc:bob@example.com").Short('p').Required().String()
    title = app.Flag("title", "Title text").Short('t').String()
)

func checkFile(name string) (string, error) {
    buf := name

    fi, err := os.Lstat(name)
    if err != nil {
        root, _ := os.Getwd()
        fullname := filepath.Join(root, name)
        fi, err = os.Lstat(fullname)
        if err != nil {
            return buf, errors.Wrap(err, "lstat failed")
        }
        buf = fullname
    }

    if fi == nil || !fi.Mode().IsRegular() {
        return buf, errors.New("file invalid")
    }

    return buf, nil
}

func sendMail(config *Config, data *Mail) error {
    msg := gomail.NewMessage()

    msg.SetAddressHeader("From", config.Sender, data.From)
    msg.SetHeader("Cc", data.Cc[:]...)
    msg.SetHeader("Subject", data.Subject)
    msg.SetHeader("To", data.To[:]...)
    msg.SetBody(data.ContentType, data.Body)

    for _, item := range data.Attachment {
        msg.Attach(item, gomail.Rename(filepath.Base(item)))
    }

    dialer := gomail.NewDialer(config.Host, config.Port, config.User, config.Pass)

    if err := dialer.DialAndSend(msg); err != nil {
        return errors.Wrap(err, "send failed")
    }

    return nil
}

func collectDifference(data []string, other []string) []string {
    var buf []string
    key := make(map[string]bool)

    for _, item := range other {
        if _, isPresent := key[item]; !isPresent {
            key[item] = true
        }
    }

    for _, item := range data {
        if _, isPresent := key[item]; !isPresent {
            buf = append(buf, item)
        }
    }

    return buf
}

func removeDuplicates(data []string) []string {
    var buf []string
    key := make(map[string]bool)

    for _, item := range data {
        if _, isPresent := key[item]; !isPresent {
            key[item] = true
            buf = append(buf, item)
        }
    }

    return buf
}

func parseRecipients(config *Config, data string) ([]string, []string) {
    var cc []string
    var to []string

    buf := strings.Split(data, config.Sep)
    for _, item := range buf {
        if len(item) != 0 {
            if hasPrefix := strings.HasPrefix(item, "cc:"); hasPrefix {
                buf := strings.ReplaceAll(item, "cc:", "")
                if len(buf) != 0 {
                    cc = append(cc, buf)
                }
            } else {
                to = append(to, item)
            }
        }
    }

    cc = removeDuplicates(cc)
    to = removeDuplicates(to)
    cc = collectDifference(cc, to)

    return cc, to
}

func parseContentType(data string) (string, error) {
    buf, isPresent := contentTypeMap[data]
    if !isPresent {
        return "", errors.New("content type invalid")
    }

    return buf, nil
}

func parseBody(data string) (string, error) {
    _name, err := checkFile(data)
    if err != nil {
        return data, nil
    }

    buf, err := ioutil.ReadFile(_name)
    if err != nil {
        return data, errors.Wrap(err, "read failed")
    }

    return string(buf), nil
}

func parseAttachment(config *Config, name string) ([]string, error) {
    var err error
    var names []string

    if len(name) == 0 {
        return names, nil
    }

    names = strings.Split(name, config.Sep)
    for i := 0; i < len(names); i++ {
        names[i], err = checkFile(names[i])
        if err != nil {
            return nil, err
        }
    }

    return names, nil
}

func parseConfig(name string) (Config, error) {
    var config Config

    fi, err := os.Open(name)
    if err != nil {
        return config, errors.Wrap(err, "open failed")
    }

    defer fi.Close()

    buf, _ := ioutil.ReadAll(fi)
    if err := json.Unmarshal(buf, &config); err != nil {
        return config, errors.Wrap(err, "unmarshal failed")
    }

    return config, nil
}

func main() {
    kingpin.MustParse(app.Parse(os.Args[1:]))

    config, err := parseConfig(*config)
    if err != nil {
        log.Fatal("Invalid config")
    }

    attachment, err := parseAttachment(&config, *attachment)
    if err != nil {
        log.Fatal("Invalid attachment")
    }

    body, err := parseBody(*body)
    if err != nil {
        log.Fatal("Invalid body")
    }

    contentType, err := parseContentType(*contentType)
    if err != nil {
        log.Fatal("Invalid content_type")
    }

    cc, to := parseRecipients(&config, *recipients)
    if len(cc) == 0 && len(to) == 0 {
        log.Fatal("Invalid recipients")
    }

    mail := Mail {
        attachment,
        body,
        cc,
        contentType,
        *header,
        *title,
        to,
    }

    if err := sendMail(&config, &mail); err != nil {
        log.Fatal("Failed to send mail")
    }

    os.Exit(0)
}
