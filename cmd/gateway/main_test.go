package main

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name      string
		expSubMsg string
		ip        string
		expErr    bool
	}{
		{
			name:      "var not set",
			ip:        "",
			expErr:    true,
			expSubMsg: "must be set",
		},
		{
			name:      "invalid ip address",
			ip:        "invalid",
			expErr:    true,
			expSubMsg: "must be a valid",
		},
		{
			name:   "valid ip address",
			ip:     "1.2.3.4",
			expErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateIP(tc.ip)
			if !tc.expErr {
				g.Expect(err).ToNot(HaveOccurred())
			} else {
				g.Expect(err.Error()).To(ContainSubstring(tc.expSubMsg))
			}
		})
	}
}
