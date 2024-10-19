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

func toScreamingSnake(input string) string {
	// Regular expression to match camelCase words
	re := regexp.MustCompile("([a-z])([A-Z])")

	// Insert a space between camelCase words and replace spaces with underscores
	snake := re.ReplaceAllString(input, "${1}_${2}")
	snake = strings.ReplaceAll(snake, " ", "_")

	// Convert to upper case
	return strings.ToUpper(snake)
}

func constructPyList(strs []string, varName string) string {
	// Create a string that can be pasted into Python
	//
	// Maximum line length as per PEP-8:
	// https://peps.python.org/pep-0008#maximum-line-length
	maxLineLength := 79
	indentNumber := 4
	currentLineLength := 0
	var lines []string
	var currentLine strings.Builder
	for _, str := range strs {
		strStr := fmt.Sprintf("\"%s\",", str)

		// New line if the addition of the scheme will go over the maximum
		// line length as defined by PEP-8
		if currentLineLength+len(strStr) > maxLineLength {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLineLength = 0
		}

		// Add indent to each new line
		// https://stackoverflow.com/a/22979015
		//
		// Use spaces and indent of 4
		if currentLine.Len() == 0 {
			indent := strings.Repeat(" ", indentNumber)
			currentLine.WriteString(indent)
			currentLineLength = indentNumber
		}

		// Add space between elements of the list
		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
			currentLineLength += 1
		}

		// Add the scheme to the current line
		currentLine.WriteString(strStr)
		currentLineLength += len(strStr)
	}

	// Add the final line to the list
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	// Join the output
	varName = toScreamingSnake(varName)
	return fmt.Sprintf("%s = [\n%s\n]", varName, strings.Join(lines, "\n"))
}

func constructPyDict(keys []string, values []string, varName string) string {
	if len(keys) != len(values) {
		fmt.Printf("[ERROR] Keys and values must be the same length: keys length = %d, values length = %d", len(keys), len(values))
		os.Exit(1)
	}

	indentNumber := 4
	var lines []string

	// Each new key-value pair is on a new line
	// https://stackoverflow.com/a/18139301
	for i, key := range keys {
		indent := strings.Repeat(" ", indentNumber)
		lines = append(lines, fmt.Sprintf("%s\"%s\": \"%s\",", indent, key, values[i]))
	}

	return fmt.Sprintf("%s = {\n%s\n}", varName, strings.Join(lines, "\n"))
}

func constructPyDefangDict(schemes []string, varName string) string {
	var defangedSchemed []string

	for _, scheme := range schemes {
		defangedSchemed = append(defangedSchemed, defangScheme(scheme))
	}

	return constructPyDict(schemes, defangedSchemed, varName)
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
	var schemes []string
	for i := 0; i < len(table); i++ {
		scheme := table[i]
		err := scheme.Validate()
		if err != nil {
			fmt.Printf("[ERROR] Invalid Scheme struct: %s; Scheme: %v", err, scheme)
			os.Exit(1)
		}
		scheme.UriScheme = cleanScheme(scheme.UriScheme)
		if scheme.Status == Permanent {
			schemes = append(schemes, scheme.UriScheme)
		}
	}

	// Format the output as a Python list
	pyStr := constructPyList(schemes, "uriSchemes")
	fmt.Println(pyStr)

	pyDict := constructPyDefangDict(schemes, "uriSchemesDefangedMap")
	fmt.Println(pyDict)
}
