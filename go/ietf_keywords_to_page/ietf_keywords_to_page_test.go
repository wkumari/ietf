// ietf_keywords_to_page
//
// This program reads a csv file of WG names and keywords, and created a page of the form
// Keyword -> WGs.

package main

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

func TestReadCSV(t *testing.T) {
	in := `WG1,WG2,WG3
"kw1","kw2",kw3
kw4,,
,kw1,kw2
,kw5,
`
	expected := map[string][]string{"kw1": {"WG1", "WG2"}, "kw2": {"WG2", "WG3"}, "kw3": {"WG3"}, "kw4": {"WG1"}, "kw5": {"WG2"}}
	records := ReadCSV(bufio.NewReader(strings.NewReader(in)))
	if !reflect.DeepEqual(records, expected) {
		t.Error("Expected: ", expected, " got: ", records)
	}
}
