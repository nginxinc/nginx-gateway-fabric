package shared

// Map defines an NGINX map.
type Map struct {
	Source       string
	Variable     string
	Parameters   []MapParameter
	UseHostnames bool
}

// MapParameter defines a Value and Result pair in a Map.
type MapParameter struct {
	Value  string
	Result string
}
