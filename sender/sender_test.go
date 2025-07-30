package main

import (
	"encoding/json"
	"reflect"
	"testing"
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
}

func TestSendMail(t *testing.T) {
	config, err := parseConfig("../config/sender.json")
	if err != nil {
		t.Error("FAIL")
	}

	mail := Mail{
		[]string{"../test/attach1.txt", "../test/attach2.text"},
		"../test/body.txt",
		[]string{"catherine@example.com"},
		"PLAIN_TEXT",
		"FROM",
		"SUBJECT",
		[]string{"alen@example.com, bob@example.com"},
	}

	_ = sendMail(&config, &mail)
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
		"test.@example.com",      // Local part ends with dot
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

func TestParseRecipientsWithValidation(t *testing.T) {
	config := Config{
		Sep: ",",
	}

	testCases := []struct {
		name            string
		recipients      string
		expectedCC      []string
		expectedTO      []string
		expectedValid   int
		expectedInvalid int
		expectedTotal   int
	}{
		{
			name:            "Valid emails with CC",
			recipients:      "alice@example.com,cc:bob@example.com,charlie@example.com",
			expectedCC:      []string{"bob@example.com"},
			expectedTO:      []string{"alice@example.com", "charlie@example.com"},
			expectedValid:   3,
			expectedInvalid: 0,
			expectedTotal:   3,
		},
		{
			name:            "Mixed valid and invalid emails",
			recipients:      "valid@example.com,invalid.email,cc:ccvalid@example.com,cc:invalid@",
			expectedCC:      []string{"ccvalid@example.com"},
			expectedTO:      []string{"valid@example.com"},
			expectedValid:   2,
			expectedInvalid: 2,
			expectedTotal:   4,
		},
		{
			name:            "Empty and duplicate emails",
			recipients:      "test@example.com,,test@example.com,cc:cc@example.com,cc:cc@example.com",
			expectedCC:      []string{"cc@example.com"},
			expectedTO:      []string{"test@example.com"},
			expectedValid:   2,
			expectedInvalid: 0,
			expectedTotal:   2,
		},
		{
			name:            "All invalid emails",
			recipients:      "test@,@example.com,cc:test..test@example.com",
			expectedCC:      []string{},
			expectedTO:      []string{},
			expectedValid:   0,
			expectedInvalid: 3,
			expectedTotal:   3,
		},
		{
			name:            "RFC 5322 compliant addresses",
			recipients:      `"quoted user"@example.com,cc:john@example.com,normal@example.com`,
			expectedCC:      []string{"john@example.com"},
			expectedTO:      []string{`"quoted user"@example.com`, "normal@example.com"},
			expectedValid:   3,
			expectedInvalid: 0,
			expectedTotal:   3,
		},
		{
			name:            "Complex invalid cases",
			recipients:      "test@,@example.com,cc:test..test@example.com,cc:test@example..com",
			expectedCC:      []string{},
			expectedTO:      []string{},
			expectedValid:   0,
			expectedInvalid: 4,
			expectedTotal:   4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cc, to, validation := parseRecipientsWithValidation(&config, tc.recipients)

			if !reflect.DeepEqual(cc, tc.expectedCC) {
				t.Errorf("CC mismatch: expected %v, got %v", tc.expectedCC, cc)
			}

			if !reflect.DeepEqual(to, tc.expectedTO) {
				t.Errorf("TO mismatch: expected %v, got %v", tc.expectedTO, to)
			}

			if validation.ValidCount != tc.expectedValid {
				t.Errorf("Valid count mismatch: expected %d, got %d", tc.expectedValid, validation.ValidCount)
			}

			if validation.InvalidCount != tc.expectedInvalid {
				t.Errorf("Invalid count mismatch: expected %d, got %d", tc.expectedInvalid, validation.InvalidCount)
			}

			if validation.TotalCount != tc.expectedTotal {
				t.Errorf("Total count mismatch: expected %d, got %d", tc.expectedTotal, validation.TotalCount)
			}

			// Verify that validation result can be marshaled to JSON
			jsonData, err := json.MarshalIndent(validation, "", "  ")
			if err != nil {
				t.Errorf("Failed to marshal validation result to JSON: %v", err)
			}

			// Verify that the JSON contains expected fields
			var result map[string]interface{}
			if err := json.Unmarshal(jsonData, &result); err != nil {
				t.Errorf("Failed to unmarshal validation JSON: %v", err)
			}

			expectedFields := []string{
				"valid_addresses", "invalid_addresses", "cc_addresses",
				"to_addresses", "total_count", "valid_count", "invalid_count",
			}

			for _, field := range expectedFields {
				if _, exists := result[field]; !exists {
					t.Errorf("Missing field %s in JSON output", field)
				}
			}
		})
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
		From:        "sender@example.com",
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

		// These should be invalid
		"test@":              false, // missing domain
		"@test.com":          false, // missing local part
		"test":               false, // no @ symbol
		"test@test@test.com": false, // multiple @ symbols
		"test.@test.com":     false, // local part ends with dot
		".test@test.com":     false, // local part starts with dot
	}

	for email, expected := range edgeCases {
		actual := isValidEmail(email)
		if actual != expected {
			t.Errorf("Email %s: expected %v, got %v", email, expected, actual)
		}
	}
}

// TestSMTPValidationWithMockConfig tests the SMTP validation function with mock data
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
		{"valid@example.com", true},
		{"another.valid@test.org", true},
		{"invalid.email", false},
		{"@invalid.com", false},
		{"invalid@", false},
		{"", false},
	}

	for _, tc := range testCases {
		result := isValidEmailWithSMTP(&mockConfig, tc.email)
		if result != tc.expected {
			t.Errorf("SMTP validation for %s: expected %v, got %v", tc.email, tc.expected, result)
		}
	}
}
