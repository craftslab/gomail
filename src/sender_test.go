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
    config, valid := parseConfig("../config/sender.json")
    if !valid {
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

    sendMail(&config, &mail)
}

func TestCollectDifference(t *testing.T) {
    bufA := []string {"alen@example.com", "bob@example.com"}
    bufB := []string {"alen@example.com"}

    buf := collectDifference(bufA, bufB)
    if len(buf) != 1 {
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
    config, valid := parseConfig("../config/sender.json")
    if !valid {
        t.Error("FAIL")
    }

    recipients := "alen@example.com,cc:,cc:bob@example.com,"

    cc, to := parseRecipients(&config, recipients)
    if len(cc) == 0 || len(to) == 0 {
        t.Error("FAIL")
    }
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
    if _, valid := parseBody(""); !valid {
        t.Error("FAIL")
    }

    if _, valid := parseBody("body"); !valid {
        t.Error("FAIL")
    }

    if _, valid := parseBody("body.txt"); !valid {
        t.Error("FAIL")
    }

    if _, valid := parseBody("../test/body.txt"); !valid {
        t.Error("FAIL")
    }
}

func TestParseAttachment(t *testing.T) {
    config, valid := parseConfig("../config/sender.json")
    if !valid {
        t.Error("FAIL")
    }

    if _, valid := parseAttachment(&config, ""); !valid {
        t.Error("FAIL")
    }

    if _, valid := parseAttachment(&config, "attach1.txt,attach2.txt"); valid {
        t.Error("FAIL")
    }

    if _, valid := parseAttachment(&config, "../test/attach1.txt,../test/attach2.txt"); !valid {
        t.Error("FAIL")
    }
}

func TestParseConfig(t *testing.T) {
    if _, valid := parseConfig("../config/sender.json"); !valid {
        t.Error("FAIL")
    }
}