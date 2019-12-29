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
    "crypto/tls"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strings"
    "unicode"

    "github.com/go-ldap/ldap/v3"
    "gopkg.in/alecthomas/kingpin.v2"
)

const (
    author  = "Jia Jia"
    version = "1.2.0"
)

type Config struct {
    Base string `json:"base"`
    Host string `json:"host"`
    Pass string `json:"pass"`
    Port int `json:"port"`
    Sep string `json:"sep"`
    User string `json:"user"`
}

var (
    app = kingpin.New("recipientparser", "Recipient parser written in Go").Author(author).Version(version)

    config = app.Flag("config", "Config file, format: .json").Short('c').String()
    filter = app.Flag("filter", "Filter list, format: @example1.com,@example2.com").Short('f').String()
    recipients = app.Flag("recipients", "Recipients list, format: alen,cc:bob@example.com").Short('r').Required().String()
)

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

func filterAddress(data string, filter []string) bool {
    status := false

    for _, item := range filter {
        if endsWith := strings.HasSuffix(data, item); endsWith {
            status = true
            break
        }
    }

    return status
}

func printAddress(cc []string, to []string, filter []string) {
    cc = removeDuplicates(cc)
    to = removeDuplicates(to)
    cc = collectDifference(cc, to)

    for _, item := range to {
        if status := filterAddress(item, filter); status {
            fmt.Printf("%s,", item)
        }
    }

    for index := 0; index < len(cc)-1; index++ {
        if status := filterAddress(cc[index], filter); status {
            fmt.Printf("cc:%s,", cc[index])
        }
    }

    if status := filterAddress(cc[len(cc)-1], filter); status {
        fmt.Printf("cc:%s\n", cc[len(cc)-1])
    }
}

func queryLdap(config *Config, data string) (string, bool) {
    l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", config.Host, config.Port))
    if err != nil {
        return "", false
    }

    defer l.Close()

    err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
    if err != nil {
        return "", false
    }

    err = l.Bind(config.User, config.Pass)
    if err != nil {
        return "", false
    }

    searchRequest := ldap.NewSearchRequest(
        config.Base,
        ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
        fmt.Sprintf("(sAMAccountName=%s)", data),
        []string{"*"},
        nil,
    )

    result, err := l.Search(searchRequest)
    if err != nil {
        return "", false
    }

    if len(result.Entries) != 1 {
        return "", false
    }

    // TODO
    return "", true
}

func parseId(data string) string {
    f := func(c rune) bool {
        return !unicode.IsNumber(c)
    }

    buf := strings.FieldsFunc(data, f)
    if len(buf) != 1 {
        return ""
    }

    return buf[0]
}

func fetchAddress(config *Config, data []string) ([]string, bool) {
    var address string
    var buf []string
    status := true

    for _, item := range data {
        if found := strings.Contains(item, "@"); found {
            buf = append(buf, item)
        } else {
            if id := parseId(item); len(id) != 0 {
                if address, status = queryLdap(config, id); !status {
                    break
                }
                if len(address) != 0 {
                    buf = append(buf, address)
                }
            }
        }
    }

    return buf, status
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

func parseFilter(config *Config, data string) ([]string, bool) {
    var filter []string

    if len(data) == 0 {
        return filter, true
    }

    buf := strings.Split(data, config.Sep)
    for _, item := range buf {
        if len(item) != 0 {
            filter = append(filter, item)
        }
    }

    filter = removeDuplicates(filter)

    return filter, true
}

func parseConfig(name string) (Config, bool) {
    var config Config

    fi, err := os.Open(name)
    if err != nil {
        return config, false
    }

    defer fi.Close()

    buf, _ := ioutil.ReadAll(fi)
    err = json.Unmarshal(buf, &config)
    if err != nil {
        return config, false
    }

    return config, true
}

func main() {
    kingpin.MustParse(app.Parse(os.Args[1:]))

    config, validConfig := parseConfig(*config)
    if !validConfig {
        log.Fatal("Invalid config")
    }

    filter, validFilter := parseFilter(&config, *filter)
    if !validFilter {
        log.Fatal("Invalid filter")
    }

    cc, to := parseRecipients(&config, *recipients)
    if len(cc) == 0 && len(to) == 0 {
        log.Fatal("Invalid recipients")
    }

    cc, validCc := fetchAddress(&config, cc)
    if !validCc {
        log.Fatal("Failed to fetch cc address")
    }

    to, validTo := fetchAddress(&config, to)
    if !validTo {
        log.Fatal("Failed to fetch to address")
    }

    printAddress(cc, to, filter)

    os.Exit(0)
}
