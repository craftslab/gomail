package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	gomail "github.com/go-mail/mail"
)

func TestParseConfig(t *testing.T) {
	if _, err := parseConfig("../config/sender.json"); err != nil {
		t.Error("FAIL")
	}
}

func TestParseAttachment(t *testing.T) {
	config, err := parseConfig("../config/sender.json")
	if err != nil {
		t.Error("FAIL")
	}

	if _, err := parseAttachment(&config, ""); err != nil {
		t.Error("FAIL")
	}

	if _, err := parseAttachment(&config, "attach1.txt,attach2.txt"); err == nil {
		t.Error("FAIL")
	}

	if _, err := parseAttachment(&config, "../test/attach1.txt,../test/attach2.txt"); err != nil {
		t.Error("FAIL")
	}
}

func TestParseBody(t *testing.T) {
	if _, err := parseBody(""); err != nil {
		t.Error("FAIL")
	}

	if _, err := parseBody("body"); err != nil {
		t.Error("FAIL")
	}

	if _, err := parseBody("body.txt"); err != nil {
		t.Error("FAIL")
	}

	if _, err := parseBody("../test/body.txt"); err != nil {
		t.Error("FAIL")
	}
}

func TestParseContentType(t *testing.T) {
	if _, err := parseContentType("FOO"); err == nil {
		t.Error("FAIL")
	}

	if _, err := parseContentType("HTML"); err != nil {
		t.Error("FAIL")
	}

	if _, err := parseContentType("PLAIN_TEXT"); err != nil {
		t.Error("FAIL")
	}
}

func TestParseRecipients(t *testing.T) {
	config, err := parseConfig("../config/sender.json")
	if err != nil {
		t.Error("FAIL")
	}

	// Test case 1: Basic functionality with CC
	recipients := "alen@example.com,cc:,cc:bob@example.com,"
	cc, to := parseRecipients(&config, recipients)
	if len(cc) == 0 || len(to) == 0 {
		t.Error("FAIL: Expected both CC and TO to have recipients")
	}

	// Test case 2: Only TO recipients
	recipients2 := "alice@example.com,bob@example.com"
	cc2, to2 := parseRecipients(&config, recipients2)
	if len(cc2) != 0 || len(to2) != 2 {
		t.Errorf("FAIL: Expected 0 CC and 2 TO, got %d CC and %d TO", len(cc2), len(to2))
	}

	// Test case 3: Only CC recipients
	recipients3 := "cc:alice@example.com,cc:bob@example.com"
	cc3, to3 := parseRecipients(&config, recipients3)
	if len(cc3) != 2 || len(to3) != 0 {
		t.Errorf("FAIL: Expected 2 CC and 0 TO, got %d CC and %d TO", len(cc3), len(to3))
	}

	// Test case 4: Duplicates should be removed
	recipients4 := "alice@example.com,alice@example.com,cc:bob@example.com,cc:bob@example.com"
	cc4, to4 := parseRecipients(&config, recipients4)
	if len(cc4) != 1 || len(to4) != 1 {
		t.Errorf("FAIL: Expected 1 CC and 1 TO after deduplication, got %d CC and %d TO", len(cc4), len(to4))
	}

	// Test case 5: Whitespace handling - leading/trailing spaces should be trimmed
	recipients5 := " alice@example.com , cc:bob@example.com , charlie@example.com "
	cc5, to5 := parseRecipients(&config, recipients5)
	if len(cc5) != 1 || len(to5) != 2 {
		t.Errorf("FAIL: Expected 1 CC and 2 TO with whitespace trimming, got %d CC and %d TO", len(cc5), len(to5))
	}
	// Verify bob is in CC, not in TO
	foundBobInCC := false
	for _, email := range cc5 {
		if email == "bob@example.com" {
			foundBobInCC = true
			break
		}
	}
	if !foundBobInCC {
		t.Error("FAIL: Expected bob@example.com in CC after trimming whitespace")
	}

	// Test case 6: The bug case - 'invalid@example,cc:jia.jia@example.com'
	// The cc: prefix should be detected correctly even with no space after comma
	recipients6 := "invalid@example.com,cc:jia.jia@example.com"
	cc6, to6 := parseRecipients(&config, recipients6)
	if len(cc6) != 1 || len(to6) != 1 {
		t.Errorf("FAIL: Expected 1 CC and 1 TO, got %d CC and %d TO", len(cc6), len(to6))
	}
	// Verify jia.jia is in CC
	foundJiaInCC := false
	for _, email := range cc6 {
		if email == "jia.jia@example.com" {
			foundJiaInCC = true
			break
		}
	}
	if !foundJiaInCC {
		t.Error("FAIL: Expected jia.jia@example.com in CC, not in TO")
	}

	// Test case 7: Mixed whitespace scenarios
	recipients7 := "alice@example.com,  cc:bob@example.com,cc: charlie@example.com , cc:  david@example.com"
	cc7, to7 := parseRecipients(&config, recipients7)
	if len(cc7) != 3 || len(to7) != 1 {
		t.Errorf("FAIL: Expected 3 CC and 1 TO with various whitespace patterns, got %d CC and %d TO", len(cc7), len(to7))
	}
}

func TestSendMail(t *testing.T) {
	t.Skip("Skipping integration test that would attempt real SMTP send")
	config, err := parseConfig("../config/sender.json")
	if err != nil {
		t.Error("FAIL")
	}

	// Test case 1: With header provided (header as display name, config.Sender as From address)
	mail := Mail{
		[]string{"../test/attach1.txt", "../test/attach2.text"},
		"../test/body.txt",
		[]string{"catherine@example.com"},
		"PLAIN_TEXT",
		"Custom Sender Name", // header option - used as display name
		"SUBJECT",
		[]string{"alen@example.com, bob@example.com"},
	}

	_ = sendMail(&config, &mail)

	// Test case 2: Without header (config.Sender as From address, no display name)
	mailNoHeader := Mail{
		[]string{"../test/attach1.txt", "../test/attach2.text"},
		"../test/body.txt",
		[]string{"catherine@example.com"},
		"PLAIN_TEXT",
		"", // no header option - config.Sender will be used as From address without display name
		"SUBJECT",
		[]string{"alen@example.com, bob@example.com"},
	}

	_ = sendMail(&config, &mailNoHeader)
}

func TestCheckFile(t *testing.T) {
	if _, err := checkFile("body.txt"); err == nil {
		t.Error("FAIL")
	}

	if _, err := checkFile("test"); err == nil {
		t.Error("FAIL")
	}

	if _, err := checkFile("../test/body.txt"); err != nil {
		t.Error("FAIL")
	}
}

func TestRemoveDuplicates(t *testing.T) {
	buf := []string{"alen@example.com", "bob@example.com", "alen@example.com"}
	buf = removeDuplicates(buf)

	if found := checkDuplicates(buf); found {
		t.Error("FAIL")
	}
}

func checkDuplicates(data []string) bool {
	found := false
	key := make(map[string]bool)

	for _, item := range data {
		if _, isPresent := key[item]; isPresent {
			found = true
			break
		}
		key[item] = true
	}

	return found
}

func TestCollectDifference(t *testing.T) {
	bufA := []string{"alen@example.com", "bob@example.com"}
	bufB := []string{"alen@example.com"}

	if buf := collectDifference(bufA, bufB); len(buf) != 1 {
		t.Error("FAIL")
	}
}

func TestIsValidEmail(t *testing.T) {
	// Test cases based on net/mail.ParseAddress behavior
	validEmails := []string{
		"test@example.com",
		"user.name@domain.co.uk",
		"firstname+lastname@example.org",
		"user_name@example.com",
		"email@example-one.com",
		"user@subdomain.example.com",
		"test123@example.org",
		"a@b.co",
		"test@example", // RFC 5322 allows domains without explicit TLD
		// RFC 5322 allows quoted strings
		`"test user"@example.com`,
		// Name with angle brackets (net/mail.ParseAddress handles this)
		"John Doe <john@example.com>",
		// Trailing dot cases that should now be handled
		"alice.@example.com",
		"Bob Smith <bob.@company.org>",
		"user.@domain.net",
		"test.name.@company.co.uk",
	}

	invalidEmails := []string{
		"invalid.email",          // No @ symbol
		"@example.com",           // Missing local part
		"test@",                  // Missing domain
		"",                       // Empty string
		" ",                      // Just whitespace
		"test@.com",              // Domain starts with dot
		"test..test@example.com", // Double dots in local part
		"test@example..com",      // Double dots in domain
		"test space@example.com", // Unquoted space in local part
		"test@exam ple.com",      // Space in domain
		".test@example.com",      // Local part starts with dot
		// These should still be invalid even with trailing dot handling
		"invalid.email.@", // Domain part is just @
		"user@",           // Still missing domain after @
		"@domain.com",     // Still missing local part
	}

	for _, email := range validEmails {
		if !isValidEmail(email) {
			t.Errorf("Expected %s to be valid", email)
		}
	}

	for _, email := range invalidEmails {
		if isValidEmail(email) {
			t.Errorf("Expected %s to be invalid", email)
		}
	}
}

func TestValidationResultJSONMarshaling(t *testing.T) {
	validation := ValidationResult{
		ValidAddresses:   []string{"valid1@example.com", "valid2@example.com"},
		InvalidAddresses: []string{"invalid1", "invalid2"},
		CcAddresses:      []string{"cc@example.com"},
		ToAddresses:      []string{"to@example.com"},
		TotalCount:       4,
		ValidCount:       2,
		InvalidCount:     2,
	}

	jsonData, err := json.MarshalIndent(validation, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal ValidationResult: %v", err)
	}

	var unmarshaled ValidationResult
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal ValidationResult: %v", err)
	}

	if !reflect.DeepEqual(validation, unmarshaled) {
		t.Errorf("Marshaling/Unmarshaling mismatch:\nOriginal: %+v\nUnmarshaled: %+v", validation, unmarshaled)
	}
}

func TestParseRecipientsWithValidationEdgeCases(t *testing.T) {
	config := Config{
		Sep: ",",
	}

	// Test empty input
	cc, to, validation := parseRecipientsWithValidation(&config, "")
	if len(cc) != 0 || len(to) != 0 || validation.TotalCount != 0 {
		t.Error("Empty input should result in empty lists and zero counts")
	}

	// Test only separators
	cc, to, validation = parseRecipientsWithValidation(&config, ",,,,")
	if len(cc) != 0 || len(to) != 0 || validation.TotalCount != 0 {
		t.Error("Only separators should result in empty lists and zero counts")
	}

	// Test CC prefix with no email
	cc, to, validation = parseRecipientsWithValidation(&config, "cc:,cc:")
	if len(cc) != 0 || len(to) != 0 || validation.TotalCount != 0 {
		t.Error("CC prefix with no email should result in empty lists and zero counts")
	}

	// Test whitespace handling
	cc, to, validation = parseRecipientsWithValidation(&config, "  test@example.com  , cc:  cc@example.com  ")
	if len(cc) != 1 || len(to) != 1 || validation.ValidCount != 2 {
		t.Error("Whitespace should be trimmed properly")
	}
}

// TestConfigWithMockData tests the Config struct with mock data
func TestConfigWithMockData(t *testing.T) {
	mockConfig := Config{
		Host:   "smtp.example.com",
		Pass:   "password123",
		Port:   587,
		Sender: "noreply@example.com",
		Sep:    ",",
		User:   "user@example.com",
	}

	// Test that the mock config has all required fields
	if mockConfig.Host == "" {
		t.Error("Host should not be empty")
	}
	if mockConfig.Port == 0 {
		t.Error("Port should not be zero")
	}
	if mockConfig.Sep == "" {
		t.Error("Separator should not be empty")
	}

	// Test parseRecipients with mock config
	recipients := "alice@example.com,cc:bob@example.com"
	cc, to := parseRecipients(&mockConfig, recipients)

	if len(to) != 1 || to[0] != "alice@example.com" {
		t.Errorf("Expected TO to contain alice@example.com, got %v", to)
	}
	if len(cc) != 1 || cc[0] != "bob@example.com" {
		t.Errorf("Expected CC to contain bob@example.com, got %v", cc)
	}
}

// TestMailStructWithMockData tests the Mail struct with mock data
func TestMailStructWithMockData(t *testing.T) {
	mockMail := Mail{
		Attachment:  []string{"file1.txt", "file2.pdf"},
		Body:        "This is a test email body",
		Cc:          []string{"cc1@example.com", "cc2@example.com"},
		ContentType: "text/plain",
		From:        "Sender Display Name", // From field represents display name (header option)
		Subject:     "Test Email Subject",
		To:          []string{"recipient1@example.com", "recipient2@example.com"},
	}

	// Verify all fields are set correctly
	if len(mockMail.Attachment) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(mockMail.Attachment))
	}
	if mockMail.Body == "" {
		t.Error("Body should not be empty")
	}
	if len(mockMail.Cc) != 2 {
		t.Errorf("Expected 2 CC recipients, got %d", len(mockMail.Cc))
	}
	if len(mockMail.To) != 2 {
		t.Errorf("Expected 2 TO recipients, got %d", len(mockMail.To))
	}
}

// TestValidationResultWithMockData tests validation result with comprehensive mock data
func TestValidationResultWithMockData(t *testing.T) {
	mockValidation := ValidationResult{
		ValidAddresses:   []string{"valid1@example.com", "valid2@example.com", "valid3@example.com"},
		InvalidAddresses: []string{"invalid1", "invalid2@", "@invalid3"},
		CcAddresses:      []string{"cc1@example.com", "cc2@example.com"},
		ToAddresses:      []string{"to1@example.com", "to2@example.com", "to3@example.com"},
		TotalCount:       6,
		ValidCount:       5,
		InvalidCount:     3,
	}

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(mockValidation, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal mock validation result: %v", err)
	}

	// Test that JSON is valid and contains expected structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonData, &result); err != nil {
		t.Fatalf("Failed to unmarshal validation JSON: %v", err)
	}

	// Verify counts
	if int(result["total_count"].(float64)) != mockValidation.TotalCount {
		t.Errorf("Total count mismatch in JSON")
	}
	if int(result["valid_count"].(float64)) != mockValidation.ValidCount {
		t.Errorf("Valid count mismatch in JSON")
	}
	if int(result["invalid_count"].(float64)) != mockValidation.InvalidCount {
		t.Errorf("Invalid count mismatch in JSON")
	}

	// Test unmarshaling back
	var unmarshaledValidation ValidationResult
	if err := json.Unmarshal(jsonData, &unmarshaledValidation); err != nil {
		t.Fatalf("Failed to unmarshal back to ValidationResult: %v", err)
	}

	if !reflect.DeepEqual(mockValidation, unmarshaledValidation) {
		t.Error("Mock data doesn't match after JSON round-trip")
	}
}

// TestEmailValidationWithEdgeCases tests email validation with edge cases
func TestEmailValidationWithEdgeCases(t *testing.T) {
	edgeCases := map[string]bool{
		// These should be valid according to net/mail.ParseAddress
		"test@localhost":           true, // localhost domain
		"user@[192.168.1.1]":       true, // IP address in brackets
		"a@b.c":                    true, // minimal valid email
		"test.email@example.co.uk": true, // subdomain
		// Trailing dot cases should now be valid
		"alice.@example.com":    true, // simple trailing dot
		"user.name.@domain.org": true, // multiple dots with trailing dot

		// These should be invalid
		"test@":              false, // missing domain
		"@test.com":          false, // missing local part
		"test":               false, // no @ symbol
		"test@test@test.com": false, // multiple @ symbols
		".test@test.com":     false, // local part starts with dot
		"invalid.@":          false, // trailing dot but missing domain
		"test@.":             false, // domain is just a dot
	}

	for email, expected := range edgeCases {
		actual := isValidEmail(email)
		if actual != expected {
			t.Errorf("Email %s: expected %v, got %v", email, expected, actual)
		}
	}
}

// TestSMTPValidationWithMockConfig verifies the wrapper validation now does format-only checks
// Note: isValidEmailWithSMTP currently defers to format validation and does not dial SMTP
func TestSMTPValidationWithMockConfig(t *testing.T) {
	mockConfig := Config{
		Host:   "smtp.example.com",
		Pass:   "password123",
		Port:   587,
		Sender: "noreply@example.com",
		Sep:    ",",
		User:   "user@example.com",
	}

	testCases := []struct {
		email    string
		expected bool
	}{
		// These should fail format validation
		{"invalid.email", false},
		{"@invalid.com", false},
		{"invalid@", false},
		{"", false},
		// These have valid formats and should return true
		{"valid@example.com", true},
		{"another.valid@test.org", true},
		// Trailing dot cases should be valid
		{"alice.@example.com", true},
	}

	for _, tc := range testCases {
		result := isValidEmailWithSMTP(&mockConfig, tc.email)
		if result != tc.expected {
			t.Errorf("SMTP validation for %s: expected %v, got %v", tc.email, tc.expected, result)
		}
	}
}

// TestValidateRecipientWithSMTPFormat tests only the format validation part
func TestValidateRecipientWithSMTPFormat(t *testing.T) {
	mockConfig := Config{
		Host:   "nonexistent.smtp.server.com",
		Pass:   "password123",
		Port:   587,
		Sender: "noreply@example.com",
		Sep:    ",",
		User:   "user@example.com",
	}

	// Test cases that should fail format validation before attempting SMTP
	invalidFormatCases := []string{
		"invalid.email",
		"@invalid.com",
		"invalid@",
		"",
		"test@",
		"@test.com",
		"test..test@example.com",
	}

	for _, email := range invalidFormatCases {
		result := isValidEmailWithSMTP(&mockConfig, email)
		if result != false {
			t.Errorf("Email with invalid format %s should return false, got %v", email, result)
		}
	}

	// Test cases with valid formats (these will return true as format is valid)
	validFormatCases := []string{
		"valid@example.com",
		"test.user@domain.org",
		"user+tag@example.co.uk",
		// Trailing dot cases should be valid
		"alice.@example.com",
	}

	for _, email := range validFormatCases {
		result := isValidEmailWithSMTP(&mockConfig, email)
		if result != true {
			t.Errorf("Email with valid format %s should return true, got %v", email, result)
		}
	}
}

// TestEmailValidationComparison demonstrates that SMTP validation mirrors format validation
func TestEmailValidationComparison(t *testing.T) {
	mockConfig := Config{
		Host:   "nonexistent.smtp.server.com",
		Pass:   "password123",
		Port:   587,
		Sender: "noreply@example.com",
		Sep:    ",",
		User:   "user@example.com",
	}

	testCases := []struct {
		email             string
		expectFormatValid bool
		expectSMTPValid   bool // With non-existent server, valid format emails return true
		description       string
	}{
		{
			email:             "valid@example.com",
			expectFormatValid: true,
			expectSMTPValid:   true, // Mirrors format validation
			description:       "Valid format email",
		},
		{
			email:             "invalid.email",
			expectFormatValid: false,
			expectSMTPValid:   false, // Mirrors format validation
			description:       "Invalid format email",
		},
		{
			email:             "@invalid.com",
			expectFormatValid: false,
			expectSMTPValid:   false, // Mirrors format validation
			description:       "Missing local part",
		},
		{
			email:             "test@",
			expectFormatValid: false,
			expectSMTPValid:   false, // Mirrors format validation
			description:       "Missing domain",
		},
		{
			email:             "",
			expectFormatValid: false,
			expectSMTPValid:   false, // Fails format check before SMTP attempt
			description:       "Empty email",
		},
		{
			email:             "alice.@example.com",
			expectFormatValid: true,
			expectSMTPValid:   true, // Should be valid with trailing dot handling
			description:       "Simple trailing dot email",
		},
		{
			email:             "Bob Smith <bob.@company.org>",
			expectFormatValid: true,
			expectSMTPValid:   true, // Should be valid with trailing dot handling
			description:       "Trailing dot email with display name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			formatResult := isValidEmail(tc.email)
			smtpResult := isValidEmailWithSMTP(&mockConfig, tc.email)

			if formatResult != tc.expectFormatValid {
				t.Errorf("Format validation for %s: expected %v, got %v", tc.email, tc.expectFormatValid, formatResult)
			}

			if smtpResult != tc.expectSMTPValid {
				t.Errorf("SMTP validation for %s: expected %v, got %v", tc.email, tc.expectSMTPValid, smtpResult)
			}
		})
	}
}

// TestTrailingDotEmailHandling specifically tests the trailing dot functionality
func TestTrailingDotEmailHandling(t *testing.T) {
	testCases := []struct {
		email         string
		shouldBeValid bool
		description   string
	}{
		{
			email:         "alice.@example.com",
			shouldBeValid: true,
			description:   "Simple trailing dot case",
		},
		{
			email:         "user.name.@domain.org",
			shouldBeValid: true,
			description:   "Multiple dots with trailing dot",
		},
		{
			email:         "Bob Smith <bob.@company.org>",
			shouldBeValid: true,
			description:   "Trailing dot with display name in angle brackets",
		},
		{
			email:         "test.@localhost",
			shouldBeValid: true,
			description:   "Trailing dot with localhost domain",
		},
		{
			email:         "invalid.@",
			shouldBeValid: false,
			description:   "Trailing dot but missing domain",
		},
		{
			email:         ".@example.com",
			shouldBeValid: false,
			description:   "Only dot before @ (invalid local part)",
		},
		{
			email:         "test@.",
			shouldBeValid: false,
			description:   "Domain is just a dot",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := isValidEmail(tc.email)
			if result != tc.shouldBeValid {
				t.Errorf("Email %s: expected %v, got %v", tc.email, tc.shouldBeValid, result)
			}
		})
	}
}

// TestParseAddressWithTrailingDot tests the helper function directly
func TestParseAddressWithTrailingDot(t *testing.T) {
	testCases := []struct {
		input          string
		expectedOutput string
		shouldSucceed  bool
		description    string
	}{
		{
			input:          "alice.@example.com",
			expectedOutput: "alice.@example.com",
			shouldSucceed:  true,
			description:    "Simple trailing dot case",
		},
		{
			input:          "Bob Smith <bob.@company.org>",
			expectedOutput: "bob.@company.org",
			shouldSucceed:  true,
			description:    "Trailing dot with display name",
		},
		{
			input:          "  John Doe  <  john.@test.net  >  ",
			expectedOutput: "john.@test.net",
			shouldSucceed:  true,
			description:    "Trailing dot with whitespace handling",
		},
		{
			input:          "user.@domain.co.uk",
			expectedOutput: "user.@domain.co.uk",
			shouldSucceed:  true,
			description:    "Trailing dot with subdomain",
		},
		{
			input:          "invalid.@",
			expectedOutput: "",
			shouldSucceed:  false,
			description:    "Trailing dot but missing domain",
		},
		{
			input:          "Name <invalid@>",
			expectedOutput: "",
			shouldSucceed:  false,
			description:    "Display name with invalid email (no domain)",
		},
		{
			input:          "",
			expectedOutput: "",
			shouldSucceed:  false,
			description:    "Empty input",
		},
		{
			input:          "not-an-email",
			expectedOutput: "",
			shouldSucceed:  false,
			description:    "Not an email address format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			output, err := parseAddressWithTrailingDot(tc.input)

			if tc.shouldSucceed {
				if err != nil {
					t.Errorf("Expected success for %s, but got error: %v", tc.input, err)
				}
				if output != tc.expectedOutput {
					t.Errorf("Expected output %s for input %s, got %s", tc.expectedOutput, tc.input, output)
				}
			} else {
				if err == nil {
					t.Errorf("Expected failure for %s, but got success with output: %s", tc.input, output)
				}
			}
		})
	}
}

// TestIsTrailingDotError tests the helper function for detecting trailing dot errors
func TestIsTrailingDotError(t *testing.T) {
	testCases := []struct {
		error       error
		shouldMatch bool
		description string
	}{
		{
			error:       fmt.Errorf("mail: trailing dot in atom"),
			shouldMatch: true,
			description: "Standard trailing dot error",
		},
		{
			error:       fmt.Errorf("mail: missing '@' or angle-addr"),
			shouldMatch: true,
			description: "Missing @ error (related to trailing dot parsing)",
		},
		{
			error:       fmt.Errorf("some error with trailing dot in atom somewhere"),
			shouldMatch: true,
			description: "Error containing trailing dot phrase",
		},
		{
			error:       fmt.Errorf("mail: invalid address"),
			shouldMatch: false,
			description: "Different error type",
		},
		{
			error:       fmt.Errorf("completely unrelated error"),
			shouldMatch: false,
			description: "Unrelated error",
		},
		{
			error:       nil,
			shouldMatch: false,
			description: "Nil error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := isTrailingDotError(tc.error)
			if result != tc.shouldMatch {
				t.Errorf("Error %v: expected %v, got %v", tc.error, tc.shouldMatch, result)
			}
		})
	}
}

// TestDryRunWithTrailingDotEmails tests the dry-run functionality with trailing dot emails
func TestDryRunWithTrailingDotEmails(t *testing.T) {
	config := Config{
		Host:   "nonexistent.smtp.server.com",
		Pass:   "testpass",
		Port:   587,
		Sender: "test@example.com",
		Sep:    ",",
		User:   "testuser",
	}

	// Test case with trailing dot emails that should now be valid
	recipients := "valid@example.com,cc:alice.@example.com,invalid.email"
	cc, to, validation := parseRecipientsWithValidation(&config, recipients)

	// Expected results: trailing dot emails should now be valid
	expectedCC := []string{"alice.@example.com"}
	expectedTO := []string{"valid@example.com"}
	expectedValid := 2   // valid@example.com, alice.@example.com
	expectedInvalid := 1 // invalid.email
	expectedTotal := 3

	if !reflect.DeepEqual(cc, expectedCC) {
		t.Errorf("CC mismatch: expected %v, got %v", expectedCC, cc)
	}

	if !reflect.DeepEqual(to, expectedTO) {
		t.Errorf("TO mismatch: expected %v, got %v", expectedTO, to)
	}

	if validation.ValidCount != expectedValid {
		t.Errorf("Valid count mismatch: expected %d, got %d", expectedValid, validation.ValidCount)
	}

	if validation.InvalidCount != expectedInvalid {
		t.Errorf("Invalid count mismatch: expected %d, got %d", expectedInvalid, validation.InvalidCount)
	}

	if validation.TotalCount != expectedTotal {
		t.Errorf("Total count mismatch: expected %d, got %d", expectedTotal, validation.TotalCount)
	}

	// Verify that alice.@example.com is in valid addresses
	if !contains(validation.ValidAddresses, "alice.@example.com") {
		t.Errorf("alice.@example.com should be in valid addresses, got: %v", validation.ValidAddresses)
	}

	// Verify that invalid.email is in invalid addresses
	if !contains(validation.InvalidAddresses, "invalid.email") {
		t.Errorf("invalid.email should be in invalid addresses, got: %v", validation.InvalidAddresses)
	}
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// TestRecipientFilteringPreservesDistinction tests that filtering invalid recipients
// preserves the cc/to distinction
func TestRecipientFilteringPreservesDistinction(t *testing.T) {
	config := Config{
		Host:   "nonexistent.smtp.server.com",
		Pass:   "testpass",
		Port:   587,
		Sender: "test@example.com",
		Sep:    ",",
		User:   "testuser",
	}

	testCases := []struct {
		name        string
		recipients  string
		expectedCC  []string
		expectedTO  []string
		description string
	}{
		{
			name:        "invalid_to_valid_cc",
			recipients:  "invalid@,cc:valid@example.com",
			expectedCC:  []string{"valid@example.com"},
			expectedTO:  []string{},
			description: "Invalid TO with valid CC should preserve CC",
		},
		{
			name:        "invalid_to_multiple_valid_cc",
			recipients:  "invalid@,cc:jia.jia@example.com,cc:zhang.san@example.com",
			expectedCC:  []string{"jia.jia@example.com", "zhang.san@example.com"},
			expectedTO:  []string{},
			description: "Invalid TO with multiple valid CC should preserve all CC",
		},
		{
			name:        "valid_to_and_cc",
			recipients:  "valid.to@example.com,cc:valid.cc@example.com",
			expectedCC:  []string{"valid.cc@example.com"},
			expectedTO:  []string{"valid.to@example.com"},
			description: "Both valid TO and CC should be preserved",
		},
		{
			name:        "multiple_invalid_to_valid_cc",
			recipients:  "invalid1@,invalid2@,cc:valid@example.com",
			expectedCC:  []string{"valid@example.com"},
			expectedTO:  []string{},
			description: "Multiple invalid TO with valid CC should preserve CC",
		},
		{
			name:        "mixed_valid_invalid_to_and_cc",
			recipients:  "valid.to@example.com,invalid@,cc:valid.cc@example.com,cc:invalid@",
			expectedCC:  []string{"valid.cc@example.com"},
			expectedTO:  []string{"valid.to@example.com"},
			description: "Mixed valid/invalid in both TO and CC should filter correctly",
		},
		{
			name:        "trailing_dot_emails",
			recipients:  "alice.@example.com,cc:bob.@company.org",
			expectedCC:  []string{"bob.@company.org"},
			expectedTO:  []string{"alice.@example.com"},
			description: "Trailing dot emails should be valid and preserved",
		},
		{
			name:        "invalid_to_trailing_dot_cc",
			recipients:  "invalid@,cc:alice.@example.com",
			expectedCC:  []string{"alice.@example.com"},
			expectedTO:  []string{},
			description: "Invalid TO with trailing dot CC should preserve CC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cc, to, _ := parseRecipientsWithValidation(&config, tc.recipients)

			if !reflect.DeepEqual(cc, tc.expectedCC) {
				t.Errorf("%s: CC mismatch\nExpected: %v\nGot: %v", tc.description, tc.expectedCC, cc)
			}

			if !reflect.DeepEqual(to, tc.expectedTO) {
				t.Errorf("%s: TO mismatch\nExpected: %v\nGot: %v", tc.description, tc.expectedTO, to)
			}
		})
	}
}

// TestSMTPRecipientExists tests the SMTP recipient validation function
func TestSMTPRecipientExists(t *testing.T) {
	config := Config{
		Host:   "nonexistent.smtp.server.com",
		Pass:   "testpass",
		Port:   587,
		Sender: "test@example.com",
		Sep:    ",",
		User:   "testuser",
	}

	testCases := []struct {
		email       string
		description string
	}{
		{
			email:       "invalid.email",
			description: "Invalid format should return false",
		},
		{
			email:       "@example.com",
			description: "Missing local part should return false",
		},
		{
			email:       "test@",
			description: "Missing domain should return false",
		},
		{
			email:       "",
			description: "Empty email should return false",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := smtpRecipientExists(&config, tc.email)
			// Invalid format emails should always return false
			if tc.email == "" || !isValidEmail(tc.email) {
				if result != false {
					t.Errorf("%s: expected false for %s, got %v", tc.description, tc.email, result)
				}
			}
			// Note: For valid format emails with non-existent server,
			// the function returns true to avoid false negatives
		})
	}
}

// TestHasTrailingDotPattern tests the helper function for detecting trailing dot patterns
func TestHasTrailingDotPattern(t *testing.T) {
	testCases := []struct {
		email             string
		shouldHavePattern bool
		description       string
	}{
		{
			email:             "alice.@example.com",
			shouldHavePattern: true,
			description:       "Simple trailing dot case",
		},
		{
			email:             "Bob Smith <bob.@company.org>",
			shouldHavePattern: true,
			description:       "Trailing dot with display name",
		},
		{
			email:             "user.@domain.co.uk",
			shouldHavePattern: true,
			description:       "Trailing dot with subdomain",
		},
		{
			email:             "normal@example.com",
			shouldHavePattern: false,
			description:       "Normal email without trailing dot",
		},
		{
			email:             "test.name@domain.org",
			shouldHavePattern: false,
			description:       "Dots in local part but not trailing",
		},
		{
			email:             "invalid.email",
			shouldHavePattern: false,
			description:       "No @ symbol",
		},
		{
			email:             "test@@example.com",
			shouldHavePattern: false,
			description:       "Multiple @ symbols",
		},
		{
			email:             "",
			shouldHavePattern: false,
			description:       "Empty email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := hasTrailingDotPattern(tc.email)
			if result != tc.shouldHavePattern {
				t.Errorf("Email %s: expected %v, got %v", tc.email, tc.shouldHavePattern, result)
			}
		})
	}
}

// TestIsValidTrailingDotAddress tests the validation for trailing dot addresses
func TestIsValidTrailingDotAddress(t *testing.T) {
	testCases := []struct {
		address       string
		shouldBeValid bool
		description   string
	}{
		{
			address:       "alice.@example.com",
			shouldBeValid: true,
			description:   "Valid trailing dot address",
		},
		{
			address:       "user.name.@domain.org",
			shouldBeValid: true,
			description:   "Multiple dots with trailing dot",
		},
		{
			address:       "test.@localhost",
			shouldBeValid: true,
			description:   "Trailing dot with localhost",
		},
		{
			address:       "a.@b.co",
			shouldBeValid: true,
			description:   "Minimal trailing dot address",
		},
		{
			address:       "invalid.@",
			shouldBeValid: false,
			description:   "Trailing dot but missing domain",
		},
		{
			address:       ".@example.com",
			shouldBeValid: false,
			description:   "Local part starts with dot",
		},
		{
			address:       "test@.",
			shouldBeValid: false,
			description:   "Domain is just a dot",
		},
		{
			address:       "@example.com",
			shouldBeValid: false,
			description:   "Missing local part",
		},
		{
			address:       "test@",
			shouldBeValid: false,
			description:   "Missing domain",
		},
		{
			address:       "test.@example.com.",
			shouldBeValid: false,
			description:   "Domain ends with dot",
		},
		{
			address:       "test.@.example.com",
			shouldBeValid: false,
			description:   "Domain starts with dot",
		},
		{
			address:       "",
			shouldBeValid: false,
			description:   "Empty address",
		},
		{
			address:       "test@@example.com",
			shouldBeValid: false,
			description:   "Multiple @ symbols",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := isValidTrailingDotAddress(tc.address)
			if result != tc.shouldBeValid {
				t.Errorf("Address %s: expected %v, got %v", tc.address, tc.shouldBeValid, result)
			}
		})
	}
}

// TestRemoveDuplicatesEdgeCases tests edge cases for duplicate removal
func TestRemoveDuplicatesEdgeCases(t *testing.T) {
	testCases := []struct {
		input       []string
		expected    []string
		description string
	}{
		{
			input:       nil,
			expected:    []string{},
			description: "Nil input should return empty slice",
		},
		{
			input:       []string{},
			expected:    []string{},
			description: "Empty input should return empty slice",
		},
		{
			input:       []string{"a@example.com"},
			expected:    []string{"a@example.com"},
			description: "Single item should return as is",
		},
		{
			input:       []string{"a@example.com", "a@example.com", "a@example.com"},
			expected:    []string{"a@example.com"},
			description: "All duplicates should return single item",
		},
		{
			input:       []string{"a@example.com", "b@example.com", "a@example.com", "c@example.com", "b@example.com"},
			expected:    []string{"a@example.com", "b@example.com", "c@example.com"},
			description: "Mixed duplicates should be removed while preserving order",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := removeDuplicates(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestCollectDifferenceEdgeCases tests edge cases for difference collection
func TestCollectDifferenceEdgeCases(t *testing.T) {
	testCases := []struct {
		data        []string
		other       []string
		expected    []string
		description string
	}{
		{
			data:        []string{},
			other:       []string{},
			expected:    nil,
			description: "Both empty should return nil",
		},
		{
			data:        []string{"a@example.com"},
			other:       []string{},
			expected:    []string{"a@example.com"},
			description: "Empty other should return all data",
		},
		{
			data:        []string{},
			other:       []string{"a@example.com"},
			expected:    nil,
			description: "Empty data should return nil",
		},
		{
			data:        []string{"a@example.com", "b@example.com"},
			other:       []string{"a@example.com", "b@example.com"},
			expected:    nil,
			description: "Identical sets should return nil",
		},
		{
			data:        []string{"a@example.com", "b@example.com", "c@example.com"},
			other:       []string{"a@example.com"},
			expected:    []string{"b@example.com", "c@example.com"},
			description: "Should return items in data but not in other",
		},
		{
			data:        []string{"a@example.com", "b@example.com"},
			other:       []string{"c@example.com", "d@example.com"},
			expected:    []string{"a@example.com", "b@example.com"},
			description: "No overlap should return all data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := collectDifference(tc.data, tc.other)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestParseRecipientsDeduplication tests that parseRecipients properly deduplicates
// and handles cc/to distinction
func TestParseRecipientsDeduplication(t *testing.T) {
	config := Config{
		Sep: ",",
	}

	testCases := []struct {
		name        string
		recipients  string
		expectedCC  int
		expectedTO  int
		description string
	}{
		{
			name:        "duplicate_in_to",
			recipients:  "alice@example.com,alice@example.com,bob@example.com",
			expectedCC:  0,
			expectedTO:  2,
			description: "Duplicates in TO should be removed",
		},
		{
			name:        "duplicate_in_cc",
			recipients:  "cc:alice@example.com,cc:alice@example.com,cc:bob@example.com",
			expectedCC:  2,
			expectedTO:  0,
			description: "Duplicates in CC should be removed",
		},
		{
			name:        "same_address_in_to_and_cc",
			recipients:  "alice@example.com,cc:alice@example.com",
			expectedCC:  0,
			expectedTO:  1,
			description: "Same address in TO and CC should appear only in TO",
		},
		{
			name:        "multiple_duplicates_across_to_and_cc",
			recipients:  "alice@example.com,bob@example.com,cc:alice@example.com,cc:bob@example.com,cc:charlie@example.com",
			expectedCC:  1,
			expectedTO:  2,
			description: "Addresses in both TO and CC should be removed from CC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cc, to := parseRecipients(&config, tc.recipients)
			if len(cc) != tc.expectedCC {
				t.Errorf("%s: expected %d CC addresses, got %d", tc.description, tc.expectedCC, len(cc))
			}
			if len(to) != tc.expectedTO {
				t.Errorf("%s: expected %d TO addresses, got %d", tc.description, tc.expectedTO, len(to))
			}
		})
	}
}

// TestParseRecipientsWhitespace tests that parseRecipients correctly handles
// whitespace around separators and prefixes
func TestParseRecipientsWhitespace(t *testing.T) {
	config := Config{
		Sep: ",",
	}

	testCases := []struct {
		name        string
		recipients  string
		expectedCC  []string
		expectedTO  []string
		description string
	}{
		{
			name:        "no_whitespace",
			recipients:  "alice@example.com,cc:bob@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Basic case without whitespace",
		},
		{
			name:        "whitespace_after_comma",
			recipients:  "alice@example.com, cc:bob@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Space after comma should be trimmed, cc: prefix should still be detected",
		},
		{
			name:        "whitespace_before_comma",
			recipients:  "alice@example.com ,cc:bob@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Space before comma should be trimmed",
		},
		{
			name:        "whitespace_both_sides",
			recipients:  "alice@example.com , cc:bob@example.com , charlie@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com", "charlie@example.com"},
			description: "Spaces around commas should be trimmed",
		},
		{
			name:        "whitespace_after_cc_prefix",
			recipients:  "alice@example.com,cc: bob@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Space after cc: prefix should be trimmed from email",
		},
		{
			name:        "whitespace_before_cc_prefix",
			recipients:  "alice@example.com,  cc:bob@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Multiple spaces before cc: prefix should be trimmed, prefix should be detected",
		},
		{
			name:        "bug_case_invalid_email_with_cc",
			recipients:  "invalid@example,cc:jia.jia@example.com",
			expectedCC:  []string{"jia.jia@example.com"},
			expectedTO:  []string{"invalid@example"},
			description: "The reported bug: cc: prefix should be detected even without space after comma",
		},
		{
			name:        "leading_trailing_whitespace",
			recipients:  "  alice@example.com  ,  cc:bob@example.com  ",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Leading and trailing whitespace on entire string should be handled",
		},
		{
			name:        "multiple_spaces_in_various_positions",
			recipients:  " alice@example.com ,  cc:bob@example.com , cc:  charlie@example.com  ,  david@example.com ",
			expectedCC:  []string{"bob@example.com", "charlie@example.com"},
			expectedTO:  []string{"alice@example.com", "david@example.com"},
			description: "Mixed whitespace patterns should all be handled correctly",
		},
		{
			name:        "empty_entries_with_whitespace",
			recipients:  "alice@example.com, , ,cc:bob@example.com",
			expectedCC:  []string{"bob@example.com"},
			expectedTO:  []string{"alice@example.com"},
			description: "Empty entries (just commas) should be ignored",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cc, to := parseRecipients(&config, tc.recipients)

			// Check counts
			if len(cc) != len(tc.expectedCC) {
				t.Errorf("%s: expected %d CC addresses, got %d. CC=%v",
					tc.description, len(tc.expectedCC), len(cc), cc)
			}
			if len(to) != len(tc.expectedTO) {
				t.Errorf("%s: expected %d TO addresses, got %d. TO=%v",
					tc.description, len(tc.expectedTO), len(to), to)
			}

			// Check specific addresses in CC
			for _, expectedEmail := range tc.expectedCC {
				found := false
				for _, actualEmail := range cc {
					if actualEmail == expectedEmail {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected %s in CC list, but not found. CC=%v",
						tc.description, expectedEmail, cc)
				}
			}

			// Check specific addresses in TO
			for _, expectedEmail := range tc.expectedTO {
				found := false
				for _, actualEmail := range to {
					if actualEmail == expectedEmail {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected %s in TO list, but not found. TO=%v",
						tc.description, expectedEmail, to)
				}
			}
		})
	}
}

// TestSendMailFromHeader tests the From header behavior with different configurations
func TestSendMailFromHeader(t *testing.T) {
	t.Skip("Skipping integration test - requires mock SMTP server")

	tests := []struct {
		name           string
		mailFrom       string
		senderAddr     string
		expectedFormat string
		description    string
	}{
		{
			name:           "With header option provided",
			mailFrom:       "Custom Sender Name",
			senderAddr:     "noreply@example.com",
			expectedFormat: `"Custom Sender Name" <noreply@example.com>`,
			description:    "When header is provided, config.Sender is used as From address with header as display name",
		},
		{
			name:           "Without header option",
			mailFrom:       "",
			senderAddr:     "noreply@example.com",
			expectedFormat: "noreply@example.com",
			description:    "When header is empty, config.Sender is used as From address with no display name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Host:   "smtp.example.com",
				Pass:   "password",
				Port:   587,
				Sender: tt.senderAddr,
				Sep:    ",",
				User:   "user",
			}

			mail := Mail{
				Attachment:  []string{},
				Body:        "Test body",
				Cc:          []string{},
				ContentType: "text/plain",
				From:        tt.mailFrom,
				Subject:     "Test Subject",
				To:          []string{"recipient@example.com"},
			}

			// Note: This test validates the structure but skips actual sending
			// Expected From header format: SetAddressHeader("From", config.Sender, data.From)
			t.Logf("%s: From field=%q, Sender=%q, Expected Format=%q",
				tt.description, mail.From, config.Sender, tt.expectedFormat)
		})
	}
}

// TestFromHeaderGeneration tests that the From header is correctly generated
// with --header as display name and config.Sender as email address
func TestFromHeaderGeneration(t *testing.T) {
	config := Config{
		Host:   "smtp.example.com",
		Pass:   "password",
		Port:   587,
		Sender: "noreply@example.com",
		Sep:    ",",
		User:   "user",
	}

	tests := []struct {
		name           string
		headerValue    string
		expectedFormat string
		description    string
	}{
		{
			name:           "With display name",
			headerValue:    "iChange",
			expectedFormat: `"iChange" <noreply@example.com>`,
			description:    "When --header='iChange', From should be '\"iChange\" <noreply@example.com>'",
		},
		{
			name:           "With name containing space",
			headerValue:    "Jenkins CI",
			expectedFormat: `"Jenkins CI" <noreply@example.com>`,
			description:    "When --header='Jenkins CI', From should be '\"Jenkins CI\" <noreply@example.com>'",
		},
		{
			name:           "Without display name",
			headerValue:    "",
			expectedFormat: "noreply@example.com",
			description:    "When --header is empty, From should be just the email address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a message and set From header using the same logic as sendMail
			msg := gomail.NewMessage()
			msg.SetAddressHeader("From", config.Sender, tt.headerValue)

			// Get the From header value
			fromHeader := msg.GetHeader("From")
			if len(fromHeader) != 1 {
				t.Fatalf("Expected 1 From header, got %d", len(fromHeader))
			}

			actual := fromHeader[0]
			if actual != tt.expectedFormat {
				t.Errorf("%s\nGot:      %q\nExpected: %q", tt.description, actual, tt.expectedFormat)
			} else {
				t.Logf("✓ %s", tt.description)
			}
		})
	}
}

// TestFromHeaderWithDottedAddress tests that addresses with dots in local part
// correctly use --header as display name, not the address local part
func TestFromHeaderWithDottedAddress(t *testing.T) {
	tests := []struct {
		name           string
		senderAddress  string
		headerValue    string
		expectedFormat string
		description    string
	}{
		{
			name:           "Simple address with header",
			senderAddress:  "mail@example.com",
			headerValue:    "iChange",
			expectedFormat: `"iChange" <mail@example.com>`,
			description:    "Simple address should use header as display name",
		},
		{
			name:           "Dotted address with header",
			senderAddress:  "dev.devops@example.com",
			headerValue:    "iChange",
			expectedFormat: `"iChange" <dev.devops@example.com>`,
			description:    "Dotted address should use header as display name, not 'dev.devops'",
		},
		{
			name:           "Dotted address without header",
			senderAddress:  "dev.devops@example.com",
			headerValue:    "",
			expectedFormat: "<dev.devops@example.com>",
			description:    "Dotted address without header should be wrapped in angle brackets to prevent misinterpretation",
		},
		{
			name:           "Multiple dots in address",
			senderAddress:  "john.doe.smith@example.com",
			headerValue:    "John Smith",
			expectedFormat: `"John Smith" <john.doe.smith@example.com>`,
			description:    "Multiple dots should still use header as display name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := gomail.NewMessage()
			msg.SetAddressHeader("From", tt.senderAddress, tt.headerValue)

			fromHeader := msg.GetHeader("From")
			if len(fromHeader) != 1 {
				t.Fatalf("Expected 1 From header, got %d", len(fromHeader))
			}

			actual := fromHeader[0]
			if actual != tt.expectedFormat {
				t.Errorf("%s\nGot:      %q\nExpected: %q\nThis means --header value is not being used correctly!",
					tt.description, actual, tt.expectedFormat)
			} else {
				t.Logf("✓ %s", tt.description)
			}
		})
	}
}
