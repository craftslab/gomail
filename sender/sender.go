package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	dryRun     = app.Flag("dry-run", "Only output recipient validation JSON and exit; do not send").Short('n').Bool()
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

	cc, to = parseRecipients(&config, *recipients)

	if len(cc) == 0 && len(to) == 0 {
		if *dryRun {
			_, _, validation := parseRecipientsWithValidation(&config, *recipients)
			jsonOutput, err := json.MarshalIndent(validation, "", "  ")
			if err != nil {
				log.Println("Error marshaling validation results:", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonOutput))
			os.Exit(0)
		}
		log.Println("No valid recipients found")
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

	// In dry-run mode, output validation JSON (SMTP recipient checks when possible) and exit without sending
	if *dryRun {
		all := append([]string{}, cc...)
		all = append(all, to...)
		all = removeDuplicates(all)
		var validAddrs []string
		var invalidAddrs []string
		for _, addr := range all {
			if !isValidEmail(addr) {
				invalidAddrs = append(invalidAddrs, addr)
				continue
			}
			if smtpRecipientExists(&config, addr) {
				validAddrs = append(validAddrs, addr)
			} else {
				invalidAddrs = append(invalidAddrs, addr)
			}
		}
		// Build JSON result; cc/to remain as parsed, valid/invalid reflect checks
		validation := ValidationResult{
			ValidAddresses:   removeDuplicates(validAddrs),
			InvalidAddresses: removeDuplicates(invalidAddrs),
			CcAddresses:      cc,
			ToAddresses:      to,
			TotalCount:       len(all),
			ValidCount:       len(removeDuplicates(validAddrs)),
			InvalidCount:     len(removeDuplicates(invalidAddrs)),
		}
		jsonOutput, err := json.MarshalIndent(validation, "", "  ")
		if err != nil {
			log.Println("Error marshaling validation results:", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
		os.Exit(0)
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
	// For verbose mode, we only do format validation
	// The actual SMTP recipient validation is complex due to various server
	// configurations, TLS requirements, and security policies.
	// We'll let the actual sending process handle recipient validation.
	return isValidEmail(email)
}

// smtpRecipientExists tries to validate an email via SMTP RCPT TO without sending mail.
// It returns false only when the server clearly rejects the recipient (e.g., 550 no such user).
// On connection/TLS/auth errors, it returns true to avoid false negatives in environments
// where validation is not allowed.
func smtpRecipientExists(config *Config, email string) bool {
	if !isValidEmail(email) {
		return false
	}

	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Try implicit TLS first when port is 465
	if config.Port == 465 {
		tlsConn, err := tls.Dial("tcp", address, &tls.Config{ServerName: config.Host})
		if err == nil {
			defer func() { _ = tlsConn.Close() }()
			if rcptAcceptedTLS(tlsConn, config, email) {
				return true
			}
			// If clearly rejected, return false; otherwise fall through to lenient true
			return false
		}
		// On error, fall back to plain path below
	}

	// Plain TCP then STARTTLS if supported
	conn, err := net.DialTimeout("tcp", address, 8*time.Second)
	if err != nil {
		return true
	}
	defer func() { _ = conn.Close() }()

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		return true
	}
	defer func() { _ = client.Quit() }()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
			return true
		}
	}

	// Do not authenticate for RCPT probe; many servers allow RCPT without AUTH and
	// some explicitly block AUTH for probing. Proceed directly to MAIL/RCPT.

	if err = client.Mail(config.Sender); err != nil {
		return true
	}

	if err = client.Rcpt(email); err != nil {
		errStr := strings.ToLower(err.Error())
		// Clear rejections
		if strings.Contains(errStr, "no such user") ||
			strings.Contains(errStr, "user unknown") ||
			strings.Contains(errStr, "recipient rejected") ||
			strings.Contains(errStr, "mailbox unavailable") ||
			strings.Contains(errStr, "recipient address rejected") ||
			strings.Contains(errStr, "invalid recipient") ||
			strings.Contains(errStr, "unknown user") ||
			strings.Contains(errStr, "does not exist") ||
			strings.Contains(errStr, "550") {
			return false
		}
		return true
	}

	return true
}

// rcptAcceptedTLS is a helper for implicit TLS connections (port 465)
func rcptAcceptedTLS(tlsConn net.Conn, config *Config, email string) bool {
	client, err := smtp.NewClient(tlsConn, config.Host)
	if err != nil {
		return true
	}
	defer func() { _ = client.Quit() }()

	// Skip AUTH for implicit TLS probe as well; go straight to MAIL/RCPT
	if err = client.Mail(config.Sender); err != nil {
		return true
	}

	if err = client.Rcpt(email); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "550") || strings.Contains(errStr, "no such user") || strings.Contains(errStr, "invalid recipient") {
			return false
		}
		return true
	}

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
		// Check if this is a recipient validation error
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "no such user") ||
			strings.Contains(errStr, "user unknown") ||
			strings.Contains(errStr, "recipient rejected") ||
			strings.Contains(errStr, "550") {
			// Try to identify which specific recipients are invalid
			invalidRecipients, _ := identifyInvalidRecipients(config, data)
			if len(invalidRecipients) > 0 {
				return errors.Errorf("send failed - invalid recipients: %v (original error: %v)", invalidRecipients, err)
			}
			// If we can't identify specific invalid recipients, return the original error
			allRecipients := append(data.To, data.Cc...)
			return errors.Wrapf(err, "send failed - invalid recipient(s) detected among: %v", allRecipients)
		}
		return errors.Wrap(err, "send failed")
	}

	return nil
}

func identifyInvalidRecipients(config *Config, data *Mail) ([]string, []string) {
	var invalidRecipients []string
	var validRecipients []string

	allRecipients := append(data.To, data.Cc...)

	// Test each recipient individually by attempting to send
	for _, recipient := range allRecipients {
		if testRecipientBySending(config, data, recipient) {
			validRecipients = append(validRecipients, recipient)
		} else {
			invalidRecipients = append(invalidRecipients, recipient)
		}
	}

	return invalidRecipients, validRecipients
}

func testRecipientBySending(config *Config, data *Mail, recipient string) bool {
	// Use SMTP RCPT TO validation without actually sending emails
	return smtpRecipientExists(config, recipient)
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
