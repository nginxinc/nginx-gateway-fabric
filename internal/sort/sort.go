package sort

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// LessObjectMeta compares two ObjectMetas according to the Gateway API conflict resolution guidelines.
// See https://gateway-api.sigs.k8s.io/concepts/guidelines/?h=conflict#conflicts
func LessObjectMeta(meta1 *metav1.ObjectMeta, meta2 *metav1.ObjectMeta) bool {
	if meta1.CreationTimestamp.Equal(&meta2.CreationTimestamp) {
		if meta1.Namespace == meta2.Namespace {
			return meta1.Name < meta2.Name
		}
		return meta1.Namespace < meta2.Namespace
	}

	return meta1.CreationTimestamp.Before(&meta2.CreationTimestamp)
}
