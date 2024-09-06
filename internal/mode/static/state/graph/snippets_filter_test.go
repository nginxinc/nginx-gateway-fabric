package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

func TestProcessSnippetsFilters(t *testing.T) {
	filter1NsName := types.NamespacedName{Namespace: "test", Name: "filter-1"}
	filter2NsName := types.NamespacedName{Namespace: "other", Name: "filter-2"}
	invalidFilterNsName := types.NamespacedName{Namespace: "default", Name: "invalid"}

	filter1 := &ngfAPI.SnippetsFilter{
		Spec: ngfAPI.SnippetsFilterSpec{
			Snippets: []ngfAPI.Snippet{
				{
					Context: ngfAPI.NginxContextMain,
					Value:   "main snippet",
				},
				{
					Context: ngfAPI.NginxContextHTTP,
					Value:   "http snippet",
				},
			},
		},
	}

	invalidFilter := &ngfAPI.SnippetsFilter{
		Spec: ngfAPI.SnippetsFilterSpec{
			Snippets: []ngfAPI.Snippet{
				{
					Context: ngfAPI.NginxContextMain,
					Value:   "main snippet",
				},
				{
					Context: "invalid context",
					Value:   "invalid snippet",
				},
			},
		},
	}

	filter2 := &ngfAPI.SnippetsFilter{
		Spec: ngfAPI.SnippetsFilterSpec{
			Snippets: []ngfAPI.Snippet{
				{
					Context: ngfAPI.NginxContextHTTPServerLocation,
					Value:   "location snippet",
				},
			},
		},
	}

	tests := []struct {
		snippetsFilters      map[types.NamespacedName]*ngfAPI.SnippetsFilter
		expProcessedSnippets map[types.NamespacedName]*SnippetsFilter
		msg                  string
	}{
		{
			msg:                  "no snippets filters",
			snippetsFilters:      nil,
			expProcessedSnippets: nil,
		},
		{
			msg: "mix valid and invalid snippets filters",
			snippetsFilters: map[types.NamespacedName]*ngfAPI.SnippetsFilter{
				filter1NsName:       filter1,
				invalidFilterNsName: invalidFilter,
				filter2NsName:       filter2,
			},
			expProcessedSnippets: map[types.NamespacedName]*SnippetsFilter{
				filter1NsName: {
					Source:     filter1,
					Conditions: nil,
					Valid:      true,
				},
				filter2NsName: {
					Source:     filter2,
					Conditions: nil,
					Valid:      true,
				},
				invalidFilterNsName: {
					Source: invalidFilter,
					Conditions: []conditions.Condition{staticConds.NewSnippetsFilterInvalid(
						"spec.snippets[1].context: Unsupported value: \"invalid context\": " +
							"supported values: \"main\", \"http\", \"http.server\", \"http.server.location\"",
					)},
					Valid: false,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)

			processedSnippetsFilters := processSnippetsFilters(test.snippetsFilters)
			g.Expect(processedSnippetsFilters).To(BeEquivalentTo(test.expProcessedSnippets))
		})
	}
}

func TestValidateSnippetsFilter(t *testing.T) {
	tests := []struct {
		msg     string
		filter  *ngfAPI.SnippetsFilter
		expCond conditions.Condition
	}{
		{
			msg: "valid filter",
			filter: &ngfAPI.SnippetsFilter{
				Spec: ngfAPI.SnippetsFilterSpec{
					Snippets: []ngfAPI.Snippet{
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main snippet",
						},
						{
							Context: ngfAPI.NginxContextHTTP,
							Value:   "http snippet",
						},
					},
				},
			},
			expCond: conditions.Condition{},
		},
		{
			msg:    "empty filter",
			filter: &ngfAPI.SnippetsFilter{},
			expCond: staticConds.NewSnippetsFilterInvalid(
				"spec.snippets: Required value: at least one snippet must be provided",
			),
		},
		{
			msg: "invalid filter; invalid snippet context",
			filter: &ngfAPI.SnippetsFilter{
				Spec: ngfAPI.SnippetsFilterSpec{
					Snippets: []ngfAPI.Snippet{
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main snippet",
						},
						{
							Context: ngfAPI.NginxContextHTTP,
							Value:   "http snippet",
						},
						{
							Context: "invalid context",
							Value:   "invalid",
						},
					},
				},
			},
			expCond: staticConds.NewSnippetsFilterInvalid(
				"spec.snippets[2].context: Unsupported value: \"invalid context\": " +
					"supported values: \"main\", \"http\", \"http.server\", \"http.server.location\"",
			),
		},
		{
			msg: "invalid filter; multiple invalid snippet contexts",
			filter: &ngfAPI.SnippetsFilter{
				Spec: ngfAPI.SnippetsFilterSpec{
					Snippets: []ngfAPI.Snippet{
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main snippet",
						},
						{
							Context: "invalid context",
							Value:   "invalid",
						},
						{
							Context: "", // empty context
							Value:   "invalid too",
						},
					},
				},
			},
			expCond: staticConds.NewSnippetsFilterInvalid(
				"[spec.snippets[1].context: Unsupported value: \"invalid context\": supported values: " +
					"\"main\", \"http\", \"http.server\", \"http.server.location\", spec.snippets[2].context: " +
					"Unsupported value: \"\": supported values: \"main\", \"http\", " +
					"\"http.server\", \"http.server.location\"]",
			),
		},
		{
			msg: "invalid filter; duplicate contexts",
			filter: &ngfAPI.SnippetsFilter{
				Spec: ngfAPI.SnippetsFilterSpec{
					Snippets: []ngfAPI.Snippet{
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main snippet",
						},
						{
							Context: ngfAPI.NginxContextHTTP,
							Value:   "http snippet",
						},
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main again",
						},
					},
				},
			},
			expCond: staticConds.NewSnippetsFilterInvalid(
				"spec.snippets[2].context: Invalid value: \"main\": only one snippet is allowed per context",
			),
		},
		{
			msg: "invalid filter; duplicate contexts and invalid context",
			filter: &ngfAPI.SnippetsFilter{
				Spec: ngfAPI.SnippetsFilterSpec{
					Snippets: []ngfAPI.Snippet{
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main snippet",
						},
						{
							Context: ngfAPI.NginxContextHTTP,
							Value:   "http snippet",
						},
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main again",
						},
						{
							Context: "invalid context",
							Value:   "invalid",
						},
					},
				},
			},
			expCond: staticConds.NewSnippetsFilterInvalid(
				"[spec.snippets[2].context: Invalid value: \"main\": only one snippet is allowed per context, " +
					"spec.snippets[3].context: Unsupported value: \"invalid context\": supported values: \"main\", " +
					"\"http\", \"http.server\", \"http.server.location\"]",
			),
		},
		{
			msg: "invalid filter; empty value",
			filter: &ngfAPI.SnippetsFilter{
				Spec: ngfAPI.SnippetsFilterSpec{
					Snippets: []ngfAPI.Snippet{
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "main snippet",
						},
						{
							Context: ngfAPI.NginxContextMain,
							Value:   "", // empty value
						},
					},
				},
			},
			expCond: staticConds.NewSnippetsFilterInvalid(
				"spec.snippets[1].value: Required value: value cannot be empty",
			),
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)

			cond := validateSnippetsFilter(test.filter)
			if test.expCond != (conditions.Condition{}) {
				g.Expect(cond).ToNot(BeNil())
				g.Expect(*cond).To(Equal(test.expCond))
			} else {
				g.Expect(cond).To(BeNil())
			}
		})
	}
}
