package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	// https://stackoverflow.com/a/74328802
	"github.com/go-playground/validator/v10"
	"github.com/nfx/go-htmltable"
)

// Status types
// https://stackoverflow.com/a/71934535
type SchemeStatus string

const (
	Permanent   SchemeStatus = "Permanent"
	Provisional SchemeStatus = "Provisional"
	Historical  SchemeStatus = "Historical"
)

type Scheme struct {
	UriScheme           string       `header:"URI Scheme"`
	Template            string       `header:"Template"`
	Description         string       `header:"Description"`
	Status              SchemeStatus `header:"Status" validate:"oneof=Permanent Provisional Historical"`
	WellKnownUriSupport string       `header:"Well-Known URI Support"`
	Reference           string       `header:"Reference"`
	Notes               string       `header:"Notes"`
}

// Validate Scheme struct
// https://stackoverflow.com/a/71934231
func (s *Scheme) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

// Within s, replace characters at `positions' with the rune defined in `replacement`
//
// For example:
// ```go
// replaceAtPositions("hello", []int{1, 2}, rune('x')) == "hxxlo"
// ```
func replaceAtPositions(s string, positions []int, replacement rune) string {
	runes := []rune(s)

	for _, pos := range positions {
		if pos >= 0 && pos < len(runes) {
			runes[pos] = replacement
		}
	}

	return string(runes)
}

// The goal of defanging is to malform the URI such that it does not open if clicked
//
// However, as there is a *[re]fang* option in the Tomtils library, we need an algorithm
// to map invertibly fanged and defanged schemes.  Many libraries do not support schemes
// beyond http[s] [1, 2], as browsers do not support many different schemes.  However,
// it may be the case that different schemes are supported on different non-browser
// applications, so we *should* support defanging.
//
// [1]: https://stackoverflow.com/a/56150152
// [2]: https://github.com/ioc-fang/ioc_fanger
func defangScheme(scheme string) string {
	// TODO
	return ""
}

// Mostly, the `URI Scheme` field is good, but there is a scheme called `shttp (OBSOLETE)`,
// which we need to clean up
func cleanScheme(schemeRaw string) string {
	// Find the index of the first parenthesis
	startIndex := strings.Index(schemeRaw, "(")
	if startIndex == -1 {
		// Return the original string if there's no parenthesis
		return schemeRaw
	}
	// Extract the substring up to the first parenthesis
	return strings.TrimSpace(schemeRaw[:startIndex])
}

func main() {
	htmltable.Logger = func(_ context.Context, msg string, fields ...any) {
		fmt.Printf("[INFO] %s %v\n", msg, fields)
	}

	// Get URI Scheme table from IANA (based on RFC 7595)
	// https://stackoverflow.com/a/42289198
	url := "https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml"
	table, err := htmltable.NewSliceFromURL[Scheme](url)
	if err != nil {
		fmt.Printf("[ERROR] Could not get table by %s: %s", url, err)
		os.Exit(1)
	}

	// Collect URI schemes into a string list
	for i := 0; i < len(table); i++ {
		scheme := table[i]
		err := scheme.Validate()
		if err != nil {
			fmt.Printf("[ERROR] Invalid Scheme struct: %s; Scheme: %v", err, scheme)
			os.Exit(1)
		}
		scheme.UriScheme = cleanScheme(scheme.UriScheme)
		if scheme.Status == Permanent {
			fmt.Println(scheme)
		}
	}
}
