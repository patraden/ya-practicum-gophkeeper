package roundtrip

import "net/http"

type TripFunc func(req *http.Request) *http.Response

func (f TripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestHTTPClient(fn TripFunc) *http.Client {
	return &http.Client{Transport: fn}
}
