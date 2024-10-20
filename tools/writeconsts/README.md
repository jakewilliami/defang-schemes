# Write Defanged URI Consts

The base library will have URI schemes (and defanged variants) baked into it.  As such, every now and then, we should update the constants.  That's what this tool is for.

```bash
 $ go generate  # or go run tools/writeconsts/writeconsts.go
[INFO] Generating library file
[INFO] Found base module path at /Users/jakeireland/projects/defang-uri-schemes
[INFO] found table [columns [Range Registration Procedures] count 3]
[INFO] found table [columns [URI Scheme Template Description Status Well-Known URI Support Reference Notes] count 384]
[INFO] found table [columns [Range Registration Procedures Note] count 5]
[INFO] found table [columns [Name Range (dec) Range (hex) Range Length (Bits) Reference Change Controller] count 2]
[INFO] found table [columns [Range Registration Procedures] count 3]
[INFO] found table [columns [Value Description Reference] count 28]
[INFO] found table [columns [Range Registration Procedures] count 6]
[INFO] found table [columns [Value Description Reference] count 2]
[INFO] found table [columns [ID Name Organization Contact URI Last Updated] count 113]
[INFO] Wrote 86552 bytes to "/Users/jakeireland/projects/defang-uri-schemes/uri_scheme_consts.go"
[INFO] Successfully ran `go fmt` on output file "/Users/jakeireland/projects/defang-uri-schemes/uri_scheme_consts.go"
```
