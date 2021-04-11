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

	"github.com/go-ldap/ldap/v3"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	author  = "Jia Jia"
	version = "2.0.7"
)

type Config struct {
	Base string `json:"base"`
	Host string `json:"host"`
	Pass string `json:"pass"`
	Port int    `json:"port"`
	Sep  string `json:"sep"`
	User string `json:"user"`
}

var (
	app = kingpin.New("parser", "Recipient parser").Author(author).Version(version)

	config     = app.Flag("config", "Config file, format: .json").Short('c').String()
	filter     = app.Flag("filter", "Filter list, format: @example1.com,@example2.com").Short('f').String()
	recipients = app.Flag("recipients", "Recipients list, format: alen,cc:bob@example.com").Short('r').Required().String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	config, err := parseConfig(*config)
	if err != nil {
		log.Println("Invalid config")
		os.Exit(1)
	}

	filter, err := parseFilter(&config, *filter)
	if err != nil {
		log.Println("Invalid filter")
		os.Exit(1)
	}

	cc, to := parseRecipients(&config, *recipients)
	if len(cc) == 0 && len(to) == 0 {
		log.Println("Invalid recipients")
		os.Exit(1)
	}

	cc, err = fetchAddress(&config, cc)
	if err != nil {
		log.Println("Failed to fetch cc address")
		os.Exit(1)
	}

	to, err = fetchAddress(&config, to)
	if err != nil {
		log.Println("Failed to fetch to address")
		os.Exit(1)
	}

	printAddress(cc, to, filter)

	os.Exit(0)
}

func parseConfig(name string) (Config, error) {
	var config Config

	fi, err := os.Open(name)
	if err != nil {
		return config, errors.Wrap(err, "open failed")
	}

	defer func() { _ = fi.Close() }()

	buf, _ := ioutil.ReadAll(fi)
	if err = json.Unmarshal(buf, &config); err != nil {
		return config, errors.Wrap(err, "unmarshal failed")
	}

	return config, nil
}

func parseFilter(config *Config, data string) ([]string, error) {
	var filter []string

	if data == "" {
		return filter, nil
	}

	buf := strings.Split(data, config.Sep)
	for _, item := range buf {
		if item != "" {
			filter = append(filter, item)
		}
	}

	filter = removeDuplicates(filter)

	return filter, nil
}

func parseRecipients(config *Config, data string) (cc, to []string) {
	buf := strings.Split(data, config.Sep)
	for _, item := range buf {
		if item != "" {
			if hasPrefix := strings.HasPrefix(item, "cc:"); hasPrefix {
				if buf := strings.ReplaceAll(item, "cc:", ""); buf != "" {
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

// nolint:gosec
func fetchAddress(config *Config, data []string) ([]string, error) {
	fetch := func(data string) string {
		buf := strings.Split(data, "@")
		if len(buf) == 0 {
			return ""
		}
		return buf[0]
	}

	query := func(filter, data string) (string, error) {
		l, err := ldap.DialURL(fmt.Sprintf("%s:%d", config.Host, config.Port))
		if err != nil {
			return "", errors.Wrap(err, "dial failed")
		}
		defer l.Close()
		if err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
			return "", errors.Wrap(err, "start failed")
		}
		if err = l.Bind(config.User, config.Pass); err != nil {
			return "", errors.Wrap(err, "bind failed")
		}
		searchRequest := ldap.NewSearchRequest(
			config.Base,
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			fmt.Sprintf("(%s=%s)", filter, data),
			[]string{"*"},
			nil,
		)
		result, err := l.Search(searchRequest)
		if err != nil {
			return "", errors.Wrap(err, "search failed")
		}
		if len(result.Entries) < 1 {
			return "", errors.New("search null")
		}
		return result.Entries[0].GetAttributeValue("mail"), nil
	}

	var buf []string

	for _, item := range data {
		address, err := query("mail", item)
		if err != nil {
			if addr, err := query("sAMAccountName", fetch(item)); err == nil {
				address = addr
			}
		}
		if address != "" {
			buf = append(buf, address)
		}
	}

	return buf, nil
}

func printAddress(cc, to, filter []string) {
	cc = removeDuplicates(cc)
	to = removeDuplicates(to)
	cc = collectDifference(cc, to)

	for _, item := range to {
		if err := filterAddress(item, filter); err == nil {
			fmt.Printf("%s,", item)
		}
	}

	if len(cc) == 0 {
		return
	}

	for index := 0; index < len(cc)-1; index++ {
		if err := filterAddress(cc[index], filter); err == nil {
			fmt.Printf("cc:%s,", cc[index])
		}
	}

	if err := filterAddress(cc[len(cc)-1], filter); err == nil {
		fmt.Printf("cc:%s\n", cc[len(cc)-1])
	}
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

func collectDifference(data, other []string) []string {
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

func filterAddress(data string, filter []string) error {
	err := errors.New("filter failed")

	for _, item := range filter {
		if endsWith := strings.HasSuffix(data, item); endsWith {
			if data != item {
				err = nil
			}
			break
		}
	}

	return err
}
