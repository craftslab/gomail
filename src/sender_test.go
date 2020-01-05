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

func TestSendMail(t *testing.T) {
    config, err := parseConfig("../config/sender.json")
    if err != nil {
        t.Error("FAIL")
    }

    mail := Mail {
        []string {"../test/attach1.txt", "../test/attach2.text"},
        "../test/body.txt",
        []string {"catherine@example.com"},
        "PLAIN_TEXT",
        "FROM",
        "SUBJECT",
        []string {"alen@example.com, bob@example.com"},
    }

    _ = sendMail(&config, &mail)
}

func TestCollectDifference(t *testing.T) {
    bufA := []string {"alen@example.com", "bob@example.com"}
    bufB := []string {"alen@example.com"}

    if buf := collectDifference(bufA, bufB); len(buf) != 1 {
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

func TestRemoveDuplicates(t *testing.T) {
    buf := []string {"alen@example.com", "bob@example.com", "alen@example.com"}
    buf = removeDuplicates(buf)

    if found := checkDuplicates(buf); found {
        t.Error("FAIL")
    }
}

func TestParseRecipients(t *testing.T) {
    config, err := parseConfig("../config/sender.json")
    if err != nil {
        t.Error("FAIL")
    }

    recipients := "alen@example.com,cc:,cc:bob@example.com,"

    cc, to := parseRecipients(&config, recipients)
    if len(cc) == 0 || len(to) == 0 {
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

func TestParseConfig(t *testing.T) {
    if _, err := parseConfig("../config/sender.json"); err != nil {
        t.Error("FAIL")
    }
}
