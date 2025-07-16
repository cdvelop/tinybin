package tinybin

import . "github.com/cdvelop/tinystring"

// TinyBin specific error variables for consistent error handling.
var (
	ErrStructNotFound   = Err(D.Struct, D.Not, D.Found)
	ErrMaxDepthExceeded = Err(D.Maximum, D.Exceeds)
	ErrVarintOverflow   = Err(D.Overflow)
	ErrVarintTruncated  = Err(D.Missing)
	ErrInvalidProtocol  = Err(D.Invalid)
	ErrUnsupportedType  = Err(D.Type, D.Not, D.Supported)
)
