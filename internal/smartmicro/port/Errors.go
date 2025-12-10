package port

import "github.com/pkg/errors"

var ErrTransportHeaderTooSmall = errors.New("buffer too small for Transport Header")
var ErrTransportHeaderStartPattern = errors.New("invalid Transport Header start pattern")
var ErrPayloadTooSmall = errors.New("buffer too small for Payload")
var ErrUnsupportedProtocol = errors.New("unsupported protocol")
