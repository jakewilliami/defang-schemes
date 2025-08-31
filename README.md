<h1 align="center">Defang URI Schemes</h1>

Define defanging algorithm for non-HTTP[S] URI schemes.

## Description

To defang a URI is to remove its fangs: make it un-harmful.  Much software supports URI hyperlinks, so that a URI can be easily clicked on.  However, this poses a security risk.

There is a convention in security to defang a URI scheme such that we write `hxxp[s]` instead of `http[s]`.  Unfortunately, many schemes exist that do not have a well-defined method of defanging.

The present library suggests a simple algorithm to defang any&trade; URI scheme, with the following goals:

  1. Invalidate the URI: the "defanged" scheme must not be a valid scheme; and
  2. One-to-one mapping: the defanged scheme should map to one and only one un-fanged scheme, so that it is unambiguous for the security researcher.

---

## Quick Start

Get the module:
```bash
$ go get github.com/jakewilliami/defang-schemes
```

Basic library usage:
```go
package main

import (
	"fmt"

	"github.com/jakewilliami/defang-schemes"
)

func main() {
	scheme := defang_schemes.Map["https"]
	defanged := scheme.DefangedScheme
	fmt.Printf("%v\n", defanged)  // "hxxps"
}
```

Types:
```go
type Scheme struct {
	Scheme              string
	DefangedScheme      string
	Template            string
	Description         string
	Status              Status
	WellKnownUriSupport string
	Reference           string
	Notes               string
}

const (
	Permanent   Status = "Permanent"
	Provisional Status = "Provisional"
	Historical  Status = "Historical"
)
```

Generating the library file and checking its validity:
```go
$ go generate
[INFO] Generating library file
[INFO] Found base module path at /Users/jakeireland/projects/defang-schemes
[INFO] found table [columns [Range Registration Procedures] count 3]
[INFO] found table [columns [URI Scheme Template Description Status Well-Known URI Support Reference Notes] count 384]
[INFO] found table [columns [Range Registration Procedures Note] count 5]
[INFO] found table [columns [Name Range (dec) Range (hex) Range Length (Bits) Reference Change Controller] count 2]
[INFO] found table [columns [Range Registration Procedures] count 3]
[INFO] found table [columns [Value Description Reference] count 28]
[INFO] found table [columns [Range Registration Procedures] count 6]
[INFO] found table [columns [Value Description Reference] count 2]
[INFO] found table [columns [ID Name Organization Contact URI Last Updated] count 113]
[INFO] Wrote 86552 bytes to "/Users/jakeireland/projects/defang-schemes/consts.go"
[INFO] Successfully ran `go fmt` on output file "/Users/jakeireland/projects/defang-schemes/consts.go"
[INFO] Checking library file meets defang safety requirements
[WARN] Only checking validity of permanent URI schemes
[INFO] Checking that the defang algorithm does not produce any valid schemes
[INFO] Checking that the defang algorithm is (kind of) invertible
```

```bash
$ go run tools/defangdump/main.go
Dumping Python code for defining schemes

URI_SCHEMES = [
     "aaa", "aaas", "about", "acap", "acct", "acd", "acr", "adiumxtra", "adt",
     "afp", "afs", "aim", "amss", "android", "appdata", "apt", "ar", "ark",
     ...,
     "z39.50r", "z39.50s",
]

URI_SCHEMES_DEFANGED_MAP = {
    "aaa": "axa",
    "aaas": "aaxs",
    ...
    "z39.50s": "z39[.]50s",
}
```

## Citation

If your research depends on `defang-schemes`, please consider giving us a formal citation: [`citation.bib`](./citation.bib)
