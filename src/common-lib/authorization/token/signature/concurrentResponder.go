package signature

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	"go.uber.org/atomic"
)

// ONLY USED IN TESTS.
// Given a mock responder, return a new one that runs yours, but only after receiving *all* the requests expected.
// For example, if you want to send 100 requests, but don't want them to all finish at the exact same time, use this.
func pauseUntilConcurrency(totalRequests int, responder httpmock.Responder) httpmock.Responder {
	waitUntilAllRequestsAreIn := make(chan struct{})
	requestsReceived := atomic.Uint64{}
	return func(r *http.Request) (*http.Response, error) {
		// If all requests have been received, then let all the other threads know
		if int(requestsReceived.Add(1)) == totalRequests {
			go func() { // In a separate go routine because this one needs to also listen on the same channel
				for i := 0; i < totalRequests; i++ {
					waitUntilAllRequestsAreIn <- struct{}{}
				}
			}()
		}

		// Wait for all the other requests to get to this same spot
		<-waitUntilAllRequestsAreIn
		// Now actually send the result
		return responder(r)
	}
}
