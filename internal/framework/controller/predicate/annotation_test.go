package predicate

import (
	"testing"

	. "github.com/onsi/gomega"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestAnnotationPredicate_Create(t *testing.T) {
	t.Parallel()
	annotation := "test"

	tests := []struct {
		event     event.CreateEvent
		name      string
		expUpdate bool
	}{
		{
			name: "object has annotation",
			event: event.CreateEvent{
				Object: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			name: "object does not have annotation",
			event: event.CreateEvent{
				Object: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"diff": "one",
						},
					},
				},
			},
			expUpdate: false,
		},
		{
			name:      "object does not have any annotations",
			event:     event.CreateEvent{Object: &apiext.CustomResourceDefinition{}},
			expUpdate: false,
		},
		{
			name:      "object is nil",
			event:     event.CreateEvent{Object: nil},
			expUpdate: false,
		},
	}

	p := AnnotationPredicate{Annotation: annotation}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			update := p.Create(test.event)
			g.Expect(update).To(Equal(test.expUpdate))
		})
	}
}

func TestAnnotationPredicate_Update(t *testing.T) {
	t.Parallel()
	annotation := "test"

	tests := []struct {
		event     event.UpdateEvent
		name      string
		expUpdate bool
	}{
		{
			name: "annotation changed",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
				ObjectNew: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "two",
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			name: "annotation deleted",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
				ObjectNew: &apiext.CustomResourceDefinition{},
			},
			expUpdate: true,
		},
		{
			name: "annotation added",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{},
				ObjectNew: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
			},
			expUpdate: true,
		},
		{
			name: "annotation has not changed",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
				ObjectNew: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
			},
			expUpdate: false,
		},
		{
			name: "different annotation changed",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"diff": "one",
						},
					},
				},
				ObjectNew: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"diff": "two",
						},
					},
				},
			},
			expUpdate: false,
		},
		{
			name: "no annotations",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{},
				ObjectNew: &apiext.CustomResourceDefinition{},
			},
			expUpdate: false,
		},
		{
			name: "old object is nil",
			event: event.UpdateEvent{
				ObjectOld: nil,
				ObjectNew: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
			},
			expUpdate: false,
		},
		{
			name: "new object is nil",
			event: event.UpdateEvent{
				ObjectOld: &apiext.CustomResourceDefinition{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							annotation: "one",
						},
					},
				},
				ObjectNew: nil,
			},
			expUpdate: false,
		},
		{
			name: "both objects are nil",
			event: event.UpdateEvent{
				ObjectOld: nil,
				ObjectNew: nil,
			},
			expUpdate: false,
		},
	}

	p := AnnotationPredicate{Annotation: annotation}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			update := p.Update(test.event)
			g.Expect(update).To(Equal(test.expUpdate))
		})
	}
}
