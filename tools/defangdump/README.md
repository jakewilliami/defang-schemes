# Defang Dump

Helper tool to persist scheme data to disk.

```bash
$ go run main.go
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
