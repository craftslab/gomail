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

func TestCollectDifference(t *testing.T) {
	bufA := []string{"alen@example.com", "bob@example.com"}
	bufB := []string{"alen@example.com"}

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
	buf := []string{"alen@example.com", "bob@example.com", "alen@example.com"}
	buf = removeDuplicates(buf)

	if found := checkDuplicates(buf); found {
		t.Error("FAIL")
	}
}

func TestFilterAddress(t *testing.T) {
	address := "alen@example.com"
	filter := []string{"@example.com"}

	if err := filterAddress(address, filter); err != nil {
		t.Error("FAIL")
	}
}

func TestPrintAddress(t *testing.T) {
	filter := []string{"@example.com"}
	cc := []string{"alen@example.com"}
	to := []string{"bob@example.com"}

	printAddress(cc, to, filter)
}

func TestQueryLdap(t *testing.T) {
	config, err := parseConfig("../../config/parser.json")
	if err != nil {
		t.Error("FAIL")
	}

	if address, err := queryLdap(&config, parseId("alen10000001")); err != nil {
		if len(address) != 0 {
			t.Error("FAIL")
		}
	}
}

func TestParseId(t *testing.T) {
	if id := parseId("alen10000001"); id != "10000001" {
		t.Error("FAIL")
	}

	if id := parseId("alen00000000"); id != "00000000" {
		t.Error("FAIL")
	}
}

func TestFetchAddress(t *testing.T) {
	config, err := parseConfig("../../config/parser.json")
	if err != nil {
		t.Error("FAIL")
	}

	recipients := []string{"alen@example", "bob"}

	if address, _ := fetchAddress(&config, recipients); len(address) < 1 {
		t.Error("FAIL")
	}
}

func TestParseRecipients(t *testing.T) {
	config, err := parseConfig("../../config/parser.json")
	if err != nil {
		t.Error("FAIL")
	}

	recipients := "alen@example.com,cc:,cc:bob@example.com,"

	cc, to := parseRecipients(&config, recipients)
	if len(cc) == 0 || len(to) == 0 {
		t.Error("FAIL")
	}
}

func TestParseFilter(t *testing.T) {
	config, err := parseConfig("../../config/parser.json")
	if err != nil {
		t.Error("FAIL")
	}

	filter := "alen@example.com,,bob@example.com,"

	if _, err := parseFilter(&config, filter); err != nil {
		t.Error("FAIL")
	}
}

func TestParseConfig(t *testing.T) {
	if _, err := parseConfig("../../config/parser.json"); err != nil {
		t.Error("FAIL")
	}
}
