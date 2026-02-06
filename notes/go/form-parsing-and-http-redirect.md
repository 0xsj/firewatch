# Form Parsing & HTTP Redirect

## Form Parsing

`net/http` has built-in form parsing for `application/x-www-form-urlencoded` and `multipart/form-data`.

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Must call ParseForm() before reading values
    r.ParseForm()

    // FormValue reads from both URL query and POST body
    // POST body takes precedence
    username := r.FormValue("log")
    password := r.FormValue("pwd")
}
```

### ParseForm vs ParseMultipartForm

| Method                  | Parses                     | Max Memory |
|-------------------------|----------------------------|------------|
| `ParseForm()`           | URL query + form body      | N/A        |
| `ParseMultipartForm(n)` | multipart (file uploads)   | n bytes    |

### Where values live after parsing

```go
r.Form          // url.Values — merged query + body
r.PostForm      // url.Values — body only (no query params)
r.FormValue(k)  // Shortcut: calls ParseForm, returns first value
r.PostFormValue(k) // Shortcut: body-only version
```

### FormValue is a shortcut

`FormValue(key)` implicitly calls `ParseForm()` if not already called. It returns the first value — empty string if missing. No error returned, so it silently handles missing fields.

```go
// These are equivalent:
r.ParseForm()
v := r.Form.Get("key")

// Same as:
v := r.FormValue("key")
```

### Used in Firewatch

WordPress login handler captures attacker credentials:
```go
func (wp *WordPress) handleLoginPost(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()
    username := r.FormValue("log")   // WordPress form field name
    password := r.FormValue("pwd")   // WordPress form field name
    // Log and record the brute force attempt...
}
```

## HTTP Redirect

```go
// http.Redirect sets Location header and writes status
http.Redirect(w, r, "/wp-login.php?redirect_to=%2Fwp-admin%2F", http.StatusFound)
```

### Common redirect status codes

| Code | Constant               | Meaning                        |
|------|------------------------|--------------------------------|
| 301  | `StatusMovedPermanently` | Permanent redirect (cached)  |
| 302  | `StatusFound`          | Temporary redirect (most common) |
| 303  | `StatusSeeOther`       | Redirect after POST → GET    |
| 307  | `StatusTemporaryRedirect` | Preserve method (POST stays POST) |
| 308  | `StatusPermanentRedirect` | Permanent + preserve method  |

### What Redirect does internally

```go
func Redirect(w ResponseWriter, r *Request, url string, code int) {
    // Sets Location header
    w.Header().Set("Location", url)
    // Writes a small HTML body with the link
    w.WriteHeader(code)
    // Writes: <a href="url">Status Text</a>
}
```

Key: Call `Redirect` before writing any other response. It calls `WriteHeader`, so you can't set status afterward.
