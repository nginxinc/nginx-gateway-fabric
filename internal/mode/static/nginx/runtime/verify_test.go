package runtime

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

type transport struct{}

func (c transport) RoundTrip(_ *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
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

	newContents := []byte("4 5 6")

	readFileNew := func(string) ([]byte, error) {
		return newContents, nil
	}
	readFileError := func(string) ([]byte, error) {
		return nil, errors.New("error")
	}

	tests := []struct {
		ctx             context.Context
		readFile        readFileFunc
		name            string
		expectedVersion int
		expectError     bool
	}{
		{
			ctx:             ctx,
			expectedVersion: 42,
			readFile:        readFileNew,
			expectError:     false,
			name:            "normal case",
		},
		{
			ctx:             ctx,
			expectedVersion: 43,
			readFile:        readFileNew,
			expectError:     true,
			name:            "wrong version",
		},
		{
			ctx:             ctx,
			expectedVersion: 0,
			readFile:        readFileError,
			expectError:     true,
			name:            "no new workers",
		},
		{
			ctx:             cancellingCtx,
			expectedVersion: 0,
			readFile:        readFileNew,
			expectError:     true,
			name:            "context canceled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := c.waitForCorrectVersion(test.ctx, test.expectedVersion, "/childfile", []byte("1 2 3"), test.readFile)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestEnsureNewNginxWorkers(t *testing.T) {
	previousContents := []byte("1 2 3")
	newContents := []byte("4 5 6")

	readFileError := func(string) ([]byte, error) {
		return nil, errors.New("error")
	}

	readFilePrevious := func(string) ([]byte, error) {
		return previousContents, nil
	}

	readFileNew := func(string) ([]byte, error) {
		return newContents, nil
	}

	ctx := context.Background()

	cancellingCtx, cancel := context.WithCancel(ctx)
	time.AfterFunc(100*time.Millisecond, cancel)

	cancellingCtx2, cancel2 := context.WithCancel(ctx)
	time.AfterFunc(1*time.Millisecond, cancel2)

	tests := []struct {
		ctx              context.Context
		readFile         readFileFunc
		name             string
		previousContents []byte
		expectError      bool
	}{
		{
			ctx:              ctx,
			readFile:         readFileNew,
			previousContents: previousContents,
			expectError:      false,
			name:             "normal case",
		},
		{
			ctx:              ctx,
			readFile:         readFileError,
			previousContents: previousContents,
			expectError:      true,
			name:             "cannot read file",
		},
		{
			ctx:              cancellingCtx,
			readFile:         readFilePrevious,
			previousContents: previousContents,
			expectError:      true,
			name:             "timed out waiting for new workers",
		},
		{
			ctx:              cancellingCtx2,
			readFile:         readFilePrevious,
			previousContents: previousContents,
			expectError:      true,
			name:             "context canceled",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := ensureNewNginxWorkers(
				test.ctx,
				"/childfile",
				test.previousContents,
				test.readFile,
			)

			if test.expectError {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}
