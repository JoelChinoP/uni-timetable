package main

import (
	"encoding/json"
	"strconv"
	"testing"
)

func TestGenerateShareID(t *testing.T) {
	seen := map[string]bool{}
	for range 200 {
		id, err := generateShareID()
		if err != nil {
			t.Fatal(err)
		}
		if len(id) != shareIDLength {
			t.Fatalf("id length = %d, want %d", len(id), shareIDLength)
		}
		if !shareIDPattern.MatchString(id) {
			t.Fatalf("id %q has unexpected characters", id)
		}
		if seen[id] {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = true
	}
}

func TestNormalizeSelection(t *testing.T) {
	raw := map[string]json.RawMessage{"12": json.RawMessage("34"), "7": json.RawMessage("9")}
	out, err := normalizeSelection(raw)
	if err != nil {
		t.Fatalf("valid selection rejected: %v", err)
	}
	var decoded map[string]int
	if err := json.Unmarshal(out, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["12"] != 34 || decoded["7"] != 9 {
		t.Fatalf("decoded = %v", decoded)
	}

	cases := map[string]map[string]json.RawMessage{
		"empty":          {},
		"bad course key": {"abc": json.RawMessage("1")},
		"zero course":    {"0": json.RawMessage("1")},
		"negative group": {"1": json.RawMessage("-3")},
		"null group":     {"1": json.RawMessage("null")},
		"string group":   {"1": json.RawMessage(`"9"`)},
	}
	for name, payload := range cases {
		if _, err := normalizeSelection(payload); err == nil {
			t.Fatalf("%s: expected rejection", name)
		}
	}
}

func TestNormalizeSelectionOverflow(t *testing.T) {
	huge := map[string]json.RawMessage{}
	for index := 1; index <= maxShareItems+1; index++ {
		huge[strconv.Itoa(index)] = json.RawMessage("1")
	}
	if _, err := normalizeSelection(huge); err == nil {
		t.Fatal("oversized selection must be rejected")
	}
}
