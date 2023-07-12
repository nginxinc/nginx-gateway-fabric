package index

import "sigs.k8s.io/controller-runtime/pkg/client"

// FieldIndices is a map of field names to their indexer functions.
type FieldIndices map[string]client.IndexerFunc
