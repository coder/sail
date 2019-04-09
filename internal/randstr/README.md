# randstr

randstr provides a `Make` function for secure token generation which can be
given an arbitrary charset (`MakeCharset`) and length.

In Linux, Darwin, and FreeBSD reading from `crypto/rand` will not produce an error and so in the interest of API
simplicity, randstr panics when random data cannot be generated instead of returning an error.

See https://github.com/golang/go/wiki/CodeReviewComments#crypto-rand

```go
import (
    "crypto/rand"
    // "encoding/base64"
    // "encoding/hex"
    "fmt"
)

func Key() string {
    buf := make([]byte, 16)
    _, err := rand.Read(buf)
    if err != nil {
        panic(err)  // out of randomness, should never happen
    }
    return fmt.Sprintf("%x", buf)
    // or hex.EncodeToString(buf)
    // or base64.StdEncoding.EncodeToString(buf)
}
```

## Example

Generate a secure, random token of length 10:

```go
fmt.Printf("Token: %v", randstr.Make(20))
```
