package defang_schemes

import (
	"fmt"
	"os"
	"regexp"

	"github.com/go-playground/validator/v10"
)

// Generate new const library file with go generate
//
//go:generate echo "[INFO] Generating library file"
//go:generate go run tools/writeconsts/main.go
//go:generate echo "[INFO] Checking library file meets defang safety requirements"
//go:generate go run tools/defangcheck/main.go

// Status types
// https://stackoverflow.com/a/71934535
type Status string

const (
	Permanent   Status = "Permanent"
	Provisional Status = "Provisional"
	Historical  Status = "Historical"
)

type Scheme struct {
	Scheme              string `validate:"required"`
	DefangedScheme      string `validate:"required"`
	Template            string
	Description         string
	Status              Status `validate:"oneof=Permanent Provisional Historical"`
	WellKnownUriSupport string
	Reference           string
	Notes               string
}

// As well as [a-z], these characters are allowed in URI schemes
// https://github.com/JuliaWeb/URIs.jl/blob/dce395c3/src/URIs.jl#L91-L108
// TODO: handle user info and IPv6 hosts
var ADDITIONAL_ALLOWED_SCHEME_CHARS = []rune{'-', '+', '.'}
var ADDITIONAL_ALLOWED_SCHEME_CHARS_PATTERN = additionalAllowedSchemeCharsPattern()
var SCHEME_PATTERN = schemePattern()

// Validate Scheme struct
// https://stackoverflow.com/a/71934231
func (s *Scheme) Validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate.Struct(s)
}

func additionalAllowedSchemeCharsPattern() *regexp.Regexp {
	var allowedChars string
	for _, char := range ADDITIONAL_ALLOWED_SCHEME_CHARS {
		allowedChars += string(char)
	}
	pattern := fmt.Sprintf(`[%s]+`, regexp.QuoteMeta(allowedChars))
	return regexp.MustCompile(pattern)
}

// Construct scheme pattern to use in validation/cleaning step
func schemePattern() *regexp.Regexp {
	var allowedChars string
	for _, char := range ADDITIONAL_ALLOWED_SCHEME_CHARS {
		allowedChars += string(char)
	}
	pattern := fmt.Sprintf(`[\w%s]+`, regexp.QuoteMeta(allowedChars))
	return regexp.MustCompile(pattern)
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

func defangAtPositions(s string, positions []int) string {
	return replaceAtPositions(s, positions, rune('x'))
}

// The goal of defanging is to malform the URI such that it does not open if clicked.
//
// However, as there is a *[re]fang* option in the Tomtils library, we need an algorithm
// to map invertibly fanged and defanged schemes.  Many libraries do not support schemes
// beyond http[s] [1, 2], as browsers do not support many different schemes.  However,
// it may be the case that different schemes are supported on different non-browser
// applications, so we *should* support defanging.
//
// There is also consideration to have enough information in a defanged stream such that
// it is invertible* to its original scheme.  Actually, not invertible, as there will not
// always be enough information just from the defanged scheme to reconstruct the scheme
// without having the list of valid schemes.  So what we need is for the defanged scheme
// to be one-to-one, so that given a defanged scheme, you know that there is a single
// valid scheme.
//
// [1]: https://stackoverflow.com/a/56150152
// [2]: https://github.com/ioc-fang/ioc_fanger
func DefangScheme(scheme string) string {
	// Case 0: check for (hopefully invalid) scheme of length 1
	if len(scheme) == 1 {
		fmt.Printf("[ERROR] Unhandled scheme \"%s\" of length 1 in defang algorithm\n", scheme)
		os.Exit(1)
	}

	// Case 1: well-defined base case
	// TODO: another case where we only remove t?
	if scheme == "http" || scheme == "https" {
		return defangAtPositions(scheme, []int{1, 2})
	}

	// Case 2: classical defanging of additional characters to produce invalid schemes
	if ADDITIONAL_ALLOWED_SCHEME_CHARS_PATTERN.MatchString(scheme) {
		return ADDITIONAL_ALLOWED_SCHEME_CHARS_PATTERN.ReplaceAllStringFunc(scheme, func(match string) string {
			return fmt.Sprintf("[%s]", match)
		})
	}

	// Case 3: for 3-letter schemes, we can remove the middle one
	if len(scheme) == 3 {
		return defangAtPositions(scheme, []int{1})
	}

	// Case 4: for 2-letter schemes, defang the second character
	if len(scheme) == 2 {
		return defangAtPositions(scheme, []int{1})
	}

	// Case 5: for 4-letter schemes, there should be enough nuance to them to defang only one letter
	// whilst removing the possibility that a valid scheme remains.  We choose to remove the third
	// letter, because removing the second would produce ambiguous results (e.g., with icap and imap)
	if len(scheme) == 4 {
		return defangAtPositions(scheme, []int{2})
	}

	// Default case: all remaining schemes should have length > 4, and hence enough information
	// to naïvely defang as we do HTTP[S]
	return defangAtPositions(scheme, []int{1, 2})
}
