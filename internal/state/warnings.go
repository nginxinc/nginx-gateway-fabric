package state

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Warnings stores a list of warnings for a given object.
type Warnings map[client.Object][]string

func newWarnings() Warnings {
	return make(map[client.Object][]string)
}

// AddWarningf adds a warning for the specified object using the provided format and arguments.
func (w Warnings) AddWarningf(obj client.Object, msgFmt string, args ...interface{}) {
	w[obj] = append(w[obj], fmt.Sprintf(msgFmt, args...))
}

// AddWarning adds a warning for the specified object.
func (w Warnings) AddWarning(obj client.Object, msg string) {
	w[obj] = append(w[obj], msg)
}

// Add adds new Warnings to the map.
// Warnings for the same object are merged.
func (w Warnings) Add(warnings Warnings) {
	for k, v := range warnings {
		w[k] = append(w[k], v...)
	}
}
