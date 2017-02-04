package main

import (
	"fmt"
	"log"
	"testing"
)

func TestValidateAPIURL(t *testing.T) {
	expectedResults := map[string]error{
		"localhost:7860":        fmt.Errorf("Scheme is no set in URL: localhost:7860"),
		"http://localhost:7860": nil,
	}

	for url, res := range expectedResults {
		got := validateAPIURL(url)
		if got != nil && res != nil && got.Error() != res.Error() {
			log.Printf("Got:\n%v\nExpected:\n%v\nURL:\n%s\n", got.Error(), res.Error(), url)
			t.FailNow()
		}
	}
}
