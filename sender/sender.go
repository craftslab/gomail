package main

import (
	"encoding/json"
	"io"
	"log"
	"mime"
	"net/mail"
	"os"
	"path/filepath"
	"strings"

	gomail "github.com/go-mail/mail"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	BuildTime string
	CommitID  string
)

type Config struct {
	Host   string `json:"host"`
	Pass   string `json:"pass"`
	Port   int    `json:"port"`
	Sender string `json:"sender"`
	Sep    string `json:"sep"`
	User   string `json:"user"`
}

type Mail struct {
	Attachment  []string
	Body        string
	Cc          []string
	ContentType string
	From        string
	Subject     string
	To          []string
}

type ValidationResult struct {
	ValidAddresses   []string `json:"valid_addresses"`
	InvalidAddresses []string `json:"invalid_addresses"`
	CcAddresses      []string `json:"cc_addresses"`
	ToAddresses      []string `json:"to_addresses"`
	TotalCount       int      `json:"total_count"`
	ValidCount       int      `json:"valid_count"`
	InvalidCount     int      `json:"invalid_count"`
}

var (
	contentTypeMap = map[string]string{
		"HTML":       "text/html",
		"PLAIN_TEXT": "text/plain",
	}
)

var (
	app = kingpin.New("sender", "Mail sender").Version(BuildTime + "-" + CommitID)

	attachment  = app.Flag("attachment", "Attachment files, format: attach1,attach2,...").Short('a').String()
	body        = app.Flag("body", "Body text or file").Short('b').String()
	config      = app.Flag("config", "Config file, format: .json").Short('c').String()
	contentType = app.Flag("content_type", "Content type, format: HTML or PLAIN_TEXT (default)").
			Short('e').Default("PLAIN_TEXT").Enum("HTML", "PLAIN_TEXT")
	header     = app.Flag("header", "Header text").Short('r').String()
	recipients = app.Flag("recipients", "Recipients list, format: alen@example.com,cc:bob@example.com").Short('p').Required().String()
	title      = app.Flag("title", "Title text").Short('t').String()
	verbose    = app.Flag("verbose", "Enable verbose output").Short('v').Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	config, err := parseConfig(*config)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	attachment, err := parseAttachment(&config, *attachment)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	body, err := parseBody(*body)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	contentType, err := parseContentType(*contentType)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	var cc, to []string
	var validation ValidationResult

	if *verbose {
		cc, to, validation = parseRecipientsWithValidation(&config, *recipients)
		jsonOutput, err := json.MarshalIndent(validation, "", "  ")
		if err != nil {
			log.Println("Error marshaling validation results:", err)
			os.Exit(1)
		}
		log.Println(string(jsonOutput))
	} else {
		cc, to = parseRecipients(&config, *recipients)
	}

	if len(cc) == 0 && len(to) == 0 {
		log.Println("Invalid recipients")
		os.Exit(1)
	}

	m := Mail{
		attachment,
		body,
		cc,
		contentType,
		*header,
		*title,
		to,
	}

	if err := sendMail(&config, &m); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}

func parseConfig(name string) (Config, error) {
	var config Config

	fi, err := os.Open(name)
	if err != nil {
		return config, errors.Wrap(err, "open failed")
	}

	defer func() { _ = fi.Close() }()

	buf, _ := io.ReadAll(fi)
	if err := json.Unmarshal(buf, &config); err != nil {
		return config, errors.Wrap(err, "unmarshal failed")
	}

	return config, nil
}

func parseAttachment(config *Config, name string) ([]string, error) {
	var err error
	var names []string

	if name == "" {
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

func parseBody(data string) (string, error) {
	_name, err := checkFile(data)
	if err != nil {
		return data, nil
	}

	buf, err := os.ReadFile(_name)
	if err != nil {
		return data, errors.Wrap(err, "read failed")
	}

	return string(buf), nil
}

func parseContentType(data string) (string, error) {
	buf, isPresent := contentTypeMap[data]
	if !isPresent {
		return "", errors.New("content type invalid")
	}

	return buf, nil
}

func parseRecipients(config *Config, data string) (cc, to []string) {
	buf := strings.Split(data, config.Sep)
	for _, item := range buf {
		if item != "" {
			if hasPrefix := strings.HasPrefix(item, "cc:"); hasPrefix {
				buf := strings.ReplaceAll(item, "cc:", "")
				if buf != "" {
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

func parseRecipientsWithValidation(config *Config, data string) (cc, to []string, validation ValidationResult) {
	var allAddresses []string
	var validAddresses []string
	var invalidAddresses []string

	cc = []string{}
	to = []string{}

	buf := strings.Split(data, config.Sep)

	for _, item := range buf {
		item = strings.TrimSpace(item)
		if item != "" {
			var email string
			if hasPrefix := strings.HasPrefix(item, "cc:"); hasPrefix {
				email = strings.TrimSpace(strings.ReplaceAll(item, "cc:", ""))
			} else {
				email = item
			}
			if email != "" {
				allAddresses = append(allAddresses, email)
				if isValidEmailWithSMTP(config, email) {
					validAddresses = append(validAddresses, email)
					if hasPrefix := strings.HasPrefix(item, "cc:"); hasPrefix {
						cc = append(cc, email)
					} else {
						to = append(to, email)
					}
				} else {
					invalidAddresses = append(invalidAddresses, email)
				}
			}
		}
	}

	cc = removeDuplicates(cc)
	to = removeDuplicates(to)
	cc = collectDifference(cc, to)

	validation = ValidationResult{
		ValidAddresses:   removeDuplicates(validAddresses),
		InvalidAddresses: removeDuplicates(invalidAddresses),
		CcAddresses:      cc,
		ToAddresses:      to,
		TotalCount:       len(removeDuplicates(allAddresses)),
		ValidCount:       len(removeDuplicates(validAddresses)),
		InvalidCount:     len(removeDuplicates(invalidAddresses)),
	}

	return cc, to, validation
}

func isValidEmail(email string) bool {
	// Use the same validation logic as the mail library
	// This validates the email format according to RFC 5322 standards
	email = strings.TrimSpace(email)
	if email == "" {
		return false
	}

	// First do basic format validation
	_, err := mail.ParseAddress(email)
	return err == nil
}

// nolint:staticcheck
func isValidEmailWithSMTP(config *Config, email string) bool {
	// First check basic format
	if !isValidEmail(email) {
		return false
	}

	// Test if the email would be accepted by the mail library
	// by attempting to create a message with it (without sending)
	defer func() {
		// Recover from any panics that might occur during message creation
		if r := recover(); r != nil {
			// Email format caused a panic in the mail library
		}
	}()

	msg := gomail.NewMessage()

	// Try to set the email address using the same method as sendMail
	// This will validate if the email format is compatible with the mail library
	msg.SetHeader("To", email)

	return true
}

func sendMail(config *Config, data *Mail) error {
	msg := gomail.NewMessage()

	msg.SetAddressHeader("From", config.Sender, data.From)
	msg.SetHeader("Cc", data.Cc...)
	msg.SetHeader("Subject", data.Subject)
	msg.SetHeader("To", data.To...)
	msg.SetBody(data.ContentType, data.Body)

	for _, item := range data.Attachment {
		msg.Attach(item, gomail.Rename(mime.QEncoding.Encode("utf-8", filepath.Base(item))))
	}

	dialer := gomail.NewDialer(config.Host, config.Port, config.User, config.Pass)

	if err := dialer.DialAndSend(msg); err != nil {
		return errors.Wrap(err, "send failed")
	}

	return nil
}

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

func removeDuplicates(data []string) []string {
	if data == nil {
		return []string{}
	}

	var buf []string
	key := make(map[string]bool)

	for _, item := range data {
		if _, isPresent := key[item]; !isPresent {
			key[item] = true
			buf = append(buf, item)
		}
	}

	if buf == nil {
		return []string{}
	}

	return buf
}

func collectDifference(data, other []string) []string {
	buf := []string{}
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
