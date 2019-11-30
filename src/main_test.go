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
	"testing"
)

func TestCheckFile(t *testing.T) {
	if _, valid := checkFile("body.txt"); valid {
		t.Error("FAIL")
	}

	if _, valid := checkFile("test"); valid {
		t.Error("FAIL")
	}

	if _, valid := checkFile("../test/body.txt"); !valid {
		t.Error("FAIL")
	}
}

func TestSendMail(t *testing.T) {
	header := "header"
	to := []string {"alen@example.com,bob@example.com"}
	cc := []string {"catherine@example.com"}
	title := "title"
	contentType := "PLAIN_TEXT"
	body := "../test/body.txt"
	attachment := []string {"../test/attach1.txt", "../test/attach2.text"}

	sendMail(header, to, cc, title, contentType, body, attachment)
}

func TestParseRecipients(t *testing.T) {
	recipients := "alen@example.com,cc:bob@example.com"
	parseRecipients(recipients)
}

func TestParseContentType(t *testing.T) {
	if _, valid := parseContentType("FOO"); valid {
		t.Error("FAIL")
	}

	if _, valid := parseContentType("HTML"); !valid {
		t.Error("FAIL")
	}

	if _, valid := parseContentType("PLAIN_TEXT"); !valid {
		t.Error("FAIL")
	}
}

func TestParseBody(t *testing.T) {
	if _, valid := parseBody(""); valid {
		t.Error("FAIL")
	}

	if _, valid := parseBody("body.txt"); valid {
		t.Error("FAIL")
	}

	if _, valid := parseBody("../test/body.txt"); !valid {
		t.Error("FAIL")
	}
}

func TestParseAttachment(t *testing.T) {
	if _, valid := parseAttachment(""); !valid {
		t.Error("FAIL")
	}

	if _, valid := parseAttachment("attach1.txt,attach2.txt"); valid {
		t.Error("FAIL")
	}

	if _, valid := parseAttachment("../test/attach1.txt,../test/attach2.txt"); !valid {
		t.Error("FAIL")
	}
}
