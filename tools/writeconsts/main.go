package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	// https://stackoverflow.com/a/74328802
	"github.com/nfx/go-htmltable"

	"github.com/jakewilliami/defang-schemes"
)

// Get file path at runtime
// https://stackoverflow.com/a/38644571
var (
	_, b, _, _ = runtime.Caller(0)
	basepath   = filepath.Dir(b)
	rootpath   = filepath.Dir(filepath.Dir(basepath))
)

type Scheme struct {
	Scheme              string                `header:"URI Scheme"`
	Template            string                `header:"Template"`
	Description         string                `header:"Description"`
	Status              defang_schemes.Status `header:"Status"`
	WellKnownUriSupport string                `header:"Well-Known URI Support"`
	Reference           string                `header:"Reference"`
	Notes               string                `header:"Notes"`
}

func cleanNulls(scheme Scheme) Scheme {
	val := reflect.ValueOf(&scheme).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			if field.String() == "-" {
				field.SetString("")
			}
		}
	}
	return scheme
}

var CLEAN_SCHEME_PATTERN = cleanSchemePattern()

// Schemes from IANA can contain additional information in parentheses
func cleanSchemePattern() *regexp.Regexp {
	pattern := fmt.Sprintf(`^(%s)(?:\s+\((.*)\))?$`, defang_schemes.SCHEME_PATTERN)
	return regexp.MustCompile(pattern)
}

// Conveninence function to check for error after writing to file
func checkWriterErr(err error, file string) {
	if err != nil {
		fmt.Printf("[ERROR] Could not write line to file \"%s\": %s\n", file, err)
		os.Exit(1)
	}
}

// Mostly, the `URI Scheme` field is good, but there is a scheme called `shttp (OBSOLETE)`,
// which we need to clean up
func cleanScheme(scheme Scheme) Scheme {
	scheme = cleanNulls(scheme)

	schemeRaw := scheme.Scheme
	matches := CLEAN_SCHEME_PATTERN.FindStringSubmatch(schemeRaw)

	if matches == nil || len(matches) == 0 {
		fmt.Printf("[ERROR] Invalid scheme for \"%s\"\n", schemeRaw)
		os.Exit(1)
	}

	// Set the first match to the URI scheme
	// NOTE: we start counting from 1 because the first element is the entire match
	scheme.Scheme = matches[1]

	// If the URI scheme holds additional information, add it to notes
	if len(matches) > 2 && matches[2] != "" {
		scheme.Notes = matches[2]
	}

	// Confirm we don't have any unhandled matching information
	if len(matches) > 3 {
		fmt.Printf("[ERROR] Unhandled matching groups in scheme regex for \"%s\"\n", schemeRaw)
		os.Exit(1)
	}

	// Ensure scheme is lowercase
	scheme.Scheme = strings.ToLower(scheme.Scheme)

	// Return the (potentially modified) scheme
	return scheme
}

func main() {
	fmt.Printf("[INFO] Found base module path at %s\n", rootpath)

	htmltable.Logger = func(_ context.Context, msg string, fields ...any) {
		fmt.Printf("[INFO] %s %v\n", msg, fields)
	}

	// Get URI Scheme table from IANA (based on RFC 7595)
	// https://stackoverflow.com/a/42289198
	url := "https://www.iana.org/assignments/uri-schemes/uri-schemes.xhtml"
	table, err := htmltable.NewSliceFromURL[Scheme](url)
	if err != nil {
		fmt.Printf("[ERROR] Could not get table by %s: %s\n", url, err)
		os.Exit(1)
	}

	// Collect URI schemes into a map
	schemeMap := make(map[string]defang_schemes.Scheme, len(table))
	for i := 0; i < len(table); i++ {
		scheme := cleanScheme(table[i])

		schemeMap[scheme.Scheme] = defang_schemes.Scheme{
			Scheme:              scheme.Scheme,
			DefangedScheme:      defang_schemes.DefangScheme(scheme.Scheme),
			Template:            scheme.Template,
			Description:         scheme.Description,
			Status:              scheme.Status,
			WellKnownUriSupport: scheme.WellKnownUriSupport,
			Reference:           scheme.Reference,
			Notes:               scheme.Notes,
		}
		schemeToValidate := schemeMap[scheme.Scheme]
		err := (&schemeToValidate).Validate()
		if err != nil {
			fmt.Printf("[ERROR] Invalid Scheme struct: %s; Scheme: %v\n", err, scheme)
			os.Exit(1)
		}
	}

	// Create a sorted list of schemes
	schemeKeyVec := make([]string, len(schemeMap))
	i := 0
	for key, _ := range schemeMap {
		schemeKeyVec[i] = key
		i++
	}
	sort.Strings(schemeKeyVec)

	// Write to Go file
	// TODO: document this section
	// TODO: get package meta info dynamically
	pkgName := "defang_schemes"
	dataMapName := "Map"
	outFile := filepath.Join(rootpath, "consts.go")

	file, err := os.Create(outFile)
	if err != nil {
		fmt.Printf("[ERROR] Cannot open file \"%s\": %s\n", outFile, err)
		os.Exit(1)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write consts package header
	_, err = writer.WriteString(fmt.Sprintf("package %s\n\n", pkgName))
	checkWriterErr(err, outFile)

	// Write generated header
	// Idea comes from Simon Sawert:
	// https://github.com/bombsimon/tld-validator/blob/c0d0fbf9/cmd/tld-generator/main.go#L19
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err = writer.WriteString("/*\nTHIS FILE WAS AUTOMATICALLY GENERATED AT " + now + "\n\nDo not edit this file.  Run \"go generate\" to re-generate this file with an\nupdated version of URI schemes from:\n    iana.org/assignments/uri-schemes/uri-schemes.xhtml.\n*/\n\n")
	checkWriterErr(err, outFile)

	// Write map
	_, err = writer.WriteString("var " + dataMapName + " = map[string]Scheme{\n")
	checkWriterErr(err, outFile)

	for _, key := range schemeKeyVec {
		scheme := schemeMap[key]
		_, err = writer.WriteString(fmt.Sprintf("\"%s\": Scheme{\nScheme: \"%s\",\nDefangedScheme: \"%s\",\nTemplate: %s,\nDescription: %s,\nStatus: %s,\nWellKnownUriSupport: %s,\nReference: %s,\nNotes: %s,\n},\n", scheme.Scheme, scheme.Scheme, scheme.DefangedScheme, strconv.Quote(scheme.Template), strconv.Quote(scheme.Description), scheme.Status, strconv.Quote(scheme.WellKnownUriSupport), strconv.Quote(scheme.Reference), strconv.Quote(scheme.Notes)))
		checkWriterErr(err, outFile)
	}

	_, err = writer.WriteString("}\n\n")
	checkWriterErr(err, outFile)

	err = writer.Flush()
	if err != nil {
		fmt.Printf("[ERROR] Could not flush file writer: %s", err)
		os.Exit(1)
	}

	fileInfo, err := os.Stat(outFile)
	if err == nil {
		fmt.Printf("[INFO] Wrote %d bytes to \"%s\"\n", fileInfo.Size(), outFile)
	}

	// TODO: Would like to do this without calling to external command
	// Consider using: https://github.com/mvdan/gofumpt
	cmd := exec.Command("go", "fmt", outFile)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("[WARNING] Failed to run `go fmt` on output file \"%s\": %s\n", outFile, err)
	} else {
		fmt.Printf("[INFO] Successfully ran `go fmt` on output file \"%s\"\n", outFile)
	}
}
