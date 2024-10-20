package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jakewilliami/defang-uri-schemes"
)

type Scheme = defang_uri_schemes.Scheme

var UriSchemeMap = defang_uri_schemes.UriSchemeMap

// Importantly, confirm that a defanged scheme is not still a valid scheme
func defangedSchemeIsKnown(scheme Scheme, knownSchemes []Scheme) bool {
	for _, knownScheme := range knownSchemes {
		if scheme.DefangedUriScheme == knownScheme.UriScheme {
			return true
		}
	}
	return false
}

// Confirm that no defanged schemes are known!
func defangedSchemesAreNotValid(schemes []Scheme) {
	fmt.Println("[INFO] Checking that the defang algorithm does not produce any valid schemes")
	http_warned := false
	for _, scheme := range schemes {
		if defangedSchemeIsKnown(scheme, schemes) {
			// Warn on known edge-case
			if scheme.UriScheme == "http" || scheme.UriScheme == "hxxp" || scheme.UriScheme == "https" || scheme.UriScheme == "hxxps" {
				if !http_warned {
					fmt.Println("[WARN] HTTP[S] defangs into a valid (albeit provisional) scheme.  Given that this is a common defang method, we will allow this")
					http_warned = true
				}
			} else {
				// Non-edge case error discovered.  Log and exit
				fmt.Printf("[ERROR] Defanged scheme \"%s\" is still a valid scheme\n", scheme.DefangedUriScheme)
				os.Exit(1)
			}
		}
	}
}

// Confirm that there exists a one-to-one mapping between a scheme and its defanged variant
func defangedSchemesAreOneToOne(schemes []Scheme) {
	fmt.Println("[INFO] Checking that the defang algorithm is (kind of) invertible")
	http_warned := false
	seenDefangedSchemes := make(map[string]struct{})
	for _, scheme := range schemes {
		if _, exists := seenDefangedSchemes[scheme.DefangedUriScheme]; exists {
			// Warn on known edge-case
			if scheme.UriScheme == "http" || scheme.UriScheme == "hxxp" || scheme.UriScheme == "https" || scheme.UriScheme == "hxxps" {
				if !http_warned {
					fmt.Println("[WARN] HTTP[S] defanges into HXXP[S], which are valid (albeit provisional) schemes.  Given that these are provisional, we will allow this edge case")
					http_warned = true
				}
			} else {
				// Non-edge case error discovered
				//
				// Collect duplicate schemes for logging
				var duplicateSchemes []string
				for _, scheme1 := range schemes {
					if scheme1.DefangedUriScheme == scheme.DefangedUriScheme {
						duplicateSchemes = append(duplicateSchemes, scheme1.UriScheme)
					}
				}
				duplicates := strings.Join(duplicateSchemes, ", ")

				// Log duplicates error
				fmt.Printf("[ERROR] Defanged scheme \"%s\" is duplicated, meaning that re-fanging would be ambiguous due to the following offenders: %s\n", scheme.DefangedUriScheme, duplicates)
				os.Exit(1)
			}
		}
		seenDefangedSchemes[scheme.DefangedUriScheme] = struct{}{}
	}
}

func main() {
	// Get schemes as list
	schemes := make([]Scheme, 0, len(UriSchemeMap))
	for _, scheme := range UriSchemeMap {
		schemes = append(schemes, scheme)
	}

	// Only check validity of permanent schemes (for now?)
	fmt.Println("[WARN] Only checking validity of permanent URI schemes")
	var permanentSchemes []Scheme
	for _, scheme := range schemes {
		if scheme.Status == defang_uri_schemes.Permanent {
			permanentSchemes = append(permanentSchemes, scheme)
		}
	}

	// Perform safety checks on defang algorithm
	defangedSchemesAreNotValid(permanentSchemes)
	defangedSchemesAreOneToOne(permanentSchemes)
}
