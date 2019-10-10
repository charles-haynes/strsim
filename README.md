# strsim

## String Similarity library

Provides a number of string similarty metrics. Each of them has the signature
```go
function(a,b string) float64
```

an return a value in [0.0..1.0] where 0.0 means "not similar at all" and 1.0 means
"identical"
