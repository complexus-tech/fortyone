package safehttp

import "errors"

var (
	ErrInvalidEndpoint    = errors.New("safe http endpoint is invalid")
	ErrInsecureScheme     = errors.New("safe http endpoint must use https")
	ErrCredentialsInURL   = errors.New("safe http endpoint must not contain credentials")
	ErrFragmentInURL      = errors.New("safe http endpoint must not contain a fragment")
	ErrUnsupportedPort    = errors.New("safe http endpoint port is not allowed")
	ErrIPAddressHost      = errors.New("safe http endpoint must use a DNS hostname")
	ErrUnsafeAddress      = errors.New("safe http endpoint resolved to a non-public address")
	ErrNoAddresses        = errors.New("safe http endpoint did not resolve to an address")
	ErrRedirectDenied     = errors.New("safe http redirects are disabled")
	ErrResponseTooLarge   = errors.New("safe http response exceeded the configured limit")
	ErrUnexpectedStatus   = errors.New("safe http response status is not successful")
	ErrUnsupportedRequest = errors.New("safe http request is unsupported")
)
