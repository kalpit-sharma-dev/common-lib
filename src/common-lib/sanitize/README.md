<p align="center">
<img height=70px src="docs/images/logo.png">
<img height=70px src="docs/images/Go-Logo_Blue.png">
</p>

# Sanitize

Helper functions to sanitize input strings to protect against XSS, CSV injection, embedded HTML, etc

**Import Statement**

```go
import	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/sanitize"
```

HTML strips html tags with a very simple parser, replace common entities, and escape < and > in the result. The result is intended to be used as plain text.

```go
sanitize.HTML(s string) string
```

HTMLAllowing parses html and allow certain tags and attributes from the lists optionally specified by args -
args[0] is a list of allowed tags, Default Tags are : h1, h2, h3, h4, h5, h6, div, span, hr, p, br, b, i, strong, em, ol, ul, li, a, img, pre, code, blockquote, article, section
args[1] is a list of allowed attributes, Default Attributes are : id, class, src, href, title, alt, name, rel
If either is missing default sets are used.
Ignored Tags Are : title, script, style, iframe, frame, frameset, noframes, noembed, embed, applet, object, base

```go
sanitize.HTMLAllowing(s string, args ...[]string) (string, error)
```

Name makes a string safe to use in a file name, producing a sanitized basename replacing . or / with - No attempt is made to normalise a path or normalise case.

```go
sanitize.Name(s string) string
```

FileName makes a string safe to use in a file name, producing a sanitized basename replacing . or / with - then replacing non-ascii characters.

```go
sanitize.FileName(s string) string
```

Identifier makes a string safe to use in as metric name, producing a sanitized name replacing ., - or / with \_. No attempt is made to normalise a path or normalise case.

```go
sanitize.Identifier(s string) string
```

Path makes a string safe to use as an url path.

```go
sanitize.Path(s string) string
```

Number makes a string safe to use as a number or decimal.

```go
sanitize.Number(s string) string
```

**Banchmark**

```
BenchmarkIdentifier-2     	   30000	     48701 ns/op	    2496 B/op	      34 allocs/op
BenchmarkHTML-2           	  200000	      6437 ns/op	     424 B/op	      11 allocs/op
BenchmarkHTMLAllowing-2   	   50000	     32116 ns/op	   55343 B/op	      87 allocs/op
BenchmarkName-2           	   10000	    160895 ns/op	    9484 B/op	     223 allocs/op
BenchmarkFileName-2       	   10000	    107299 ns/op	    5903 B/op	     162 allocs/op
BenchmarkPath-2           	   30000	     53216 ns/op	    3575 B/op	     159 allocs/op
BenchmarkNumber-2         	  300000	      5633 ns/op	     496 B/op	      18 allocs/op
```

### Contribution

Any changes in this package should be communicated to Juno Team.
