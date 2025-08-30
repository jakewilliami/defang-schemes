package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jakewilliami/defang-schemes"
)

type Scheme = defang_schemes.Scheme

var SchemeMap = defang_schemes.Map

// Importantly, confirm that a defanged scheme is not still a valid scheme
func defangedSchemeIsKnown(scheme Scheme, knownSchemes []Scheme) bool {
	for _, knownScheme := range knownSchemes {
		if scheme.DefangedScheme == knownScheme.Scheme {
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
			if scheme.Scheme == "http" || scheme.Scheme == "hxxp" || scheme.Scheme == "https" || scheme.Scheme == "hxxps" {
				if !http_warned {
					fmt.Println("[WARN] HTTP[S] defangs into a valid (albeit provisional) scheme.  Given that this is a common defang method, we will allow this")
					http_warned = true
				}
			} else {
				// Non-edge case error discovered.  Log and exit
				fmt.Printf("[ERROR] Defanged scheme \"%s\" is still a valid scheme\n", scheme.DefangedScheme)
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
		if _, exists := seenDefangedSchemes[scheme.DefangedScheme]; exists {
			// Warn on known edge-case
			if scheme.Scheme == "http" || scheme.Scheme == "hxxp" || scheme.Scheme == "https" || scheme.Scheme == "hxxps" {
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
					if scheme1.DefangedScheme == scheme.DefangedScheme {
						duplicateSchemes = append(duplicateSchemes, scheme1.Scheme)
					}
				}
				duplicates := strings.Join(duplicateSchemes, ", ")

				// Log duplicates error
				fmt.Printf("[ERROR] Defanged scheme \"%s\" is duplicated, meaning that re-fanging would be ambiguous due to the following offenders: %s\n", scheme.DefangedScheme, duplicates)
				os.Exit(1)
			}
		}
		seenDefangedSchemes[scheme.DefangedScheme] = struct{}{}
	}
}

func main() {
	// Get schemes as list
	schemes := make([]Scheme, 0, len(SchemeMap))
	for _, scheme := range SchemeMap {
		schemes = append(schemes, scheme)
	}

	// Only check validity of permanent schemes (for now?)
	fmt.Println("[WARN] Only checking validity of permanent URI schemes")
	var permanentSchemes []Scheme
	for _, scheme := range schemes {
		if scheme.Status == defang_schemes.Permanent {
			permanentSchemes = append(permanentSchemes, scheme)
		}
	}

	// Perform safety checks on defang algorithm
	defangedSchemesAreNotValid(permanentSchemes)
	defangedSchemesAreOneToOne(permanentSchemes)
}
