package wsclient

import "net/http"

// AuthTransportTransport is an http.RoundTripper that adds an Authorization header to each request.
type AuthTransport struct {
	Transport http.RoundTripper
	Token     string
}

// RoundTrip executes a single HTTP transaction.
func (c *AuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original request
	clonedReq := req.Clone(req.Context())
	// Add the Authorization header
	clonedReq.Header.Set("Authorization", "Bearer "+c.Token)
	// Use the underlying transport
	return c.Transport.RoundTrip(clonedReq)
}
