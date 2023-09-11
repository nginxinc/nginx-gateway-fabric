package runtime

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type transport struct{}

func (c transport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("42")),
		Header:     make(http.Header),
	}, nil
}

func getTestHTTPClient() *http.Client {
	ts := transport{}
	return &http.Client{
		Transport: ts,
	}
}

func TestVerifyClient(t *testing.T) {
	c := verifyClient{
		client:  getTestHTTPClient(),
		timeout: 25 * time.Millisecond,
	}

	ctx := context.Background()
	cancellingCtx, cancel := context.WithCancel(ctx)
	time.AfterFunc(1*time.Millisecond, cancel)

	tests := []struct {
		ctx             context.Context
		name            string
		expectedVersion int
		expectError     bool
	}{
		{
			ctx:             ctx,
			expectedVersion: 42,
			expectError:     false,
			name:            "normal case",
		},
		{
			ctx:             ctx,
			expectedVersion: 43,
			expectError:     true,
			name:            "wrong version",
		},
		{
			ctx:             cancellingCtx,
			expectedVersion: 0,
			expectError:     true,
			name:            "context canceled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := c.waitForCorrectVersion(test.ctx, test.expectedVersion)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
