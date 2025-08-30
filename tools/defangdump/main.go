package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jakewilliami/defang-schemes"
)

type Scheme = defang_schemes.Scheme

var SchemeMap = defang_schemes.Map

type ByScheme []Scheme

// Implement the sort.Interface for ByScheme
func (a ByScheme) Len() int           { return len(a) }
func (a ByScheme) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScheme) Less(i, j int) bool { return a[i].Scheme < a[j].Scheme }

// For formatting "constant" variables in Python
func toScreamingSnake(input string) string {
	// Regular expression to match camelCase words
	re := regexp.MustCompile("([a-z])([A-Z])")

	// Insert a space between camelCase words and replace spaces with underscores
	snake := re.ReplaceAllString(input, "${1}_${2}")
	snake = strings.ReplaceAll(snake, " ", "_")

	// Convert to upper case
	return strings.ToUpper(snake)
}

// Create a string that can be pasted into Python
//
// Maximum line length as per PEP-8:
// https://peps.python.org/pep-0008#maximum-line-length
func constructPyList(strs []string, varName string) string {
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

func constructPySchemeList(schemes []Scheme, varName string) string {
	var rawSchemes []string

	for _, scheme := range schemes {
		rawSchemes = append(rawSchemes, scheme.Scheme)
	}

	return constructPyList(rawSchemes, varName)
}

func constructPyDict(keys []string, values []string, varName string) string {
	if len(keys) != len(values) {
		fmt.Printf("[ERROR] Keys and values must be the same length: keys length = %d, values length = %d\n", len(keys), len(values))
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

	varName = toScreamingSnake(varName)
	return fmt.Sprintf("%s = {\n%s\n}", varName, strings.Join(lines, "\n"))
}

func constructPyDefangSchemeDict(schemes []Scheme, varName string) string {
	var rawSchemes []string
	var defangedSchemes []string

	for _, scheme := range schemes {
		rawSchemes = append(rawSchemes, scheme.Scheme)
		defangedSchemes = append(defangedSchemes, scheme.DefangedScheme)
	}

	return constructPyDict(rawSchemes, defangedSchemes, varName)
}

func main() {
	// Get schemes as list
	schemes := make([]Scheme, 0, len(SchemeMap))
	for _, scheme := range SchemeMap {
		schemes = append(schemes, scheme)
	}
	sort.Sort(ByScheme(schemes))

	fmt.Print("Dumping Python code for defining schemes\n\n")
	pyStr := constructPySchemeList(schemes, "schemes")
	fmt.Print(pyStr, "\n\n")
	pyDict := constructPyDefangSchemeDict(schemes, "schemesDefangedMap")
	fmt.Println(pyDict)
}
