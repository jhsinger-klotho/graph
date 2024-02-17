package graph

var (
	_ Hash[string, string] = StringHash
	_ Hash[int, int]       = IntHash
)

// StringHash is a hashing function that accepts a string and uses that exact
// string as a hash value. Using it as Hash will yield a Graph[string, string].
func StringHash(v string) string {
	return v
}

// IntHash is a hashing function that accepts an integer and uses that exact
// integer as a hash value. Using it as Hash will yield a Graph[int, int].
func IntHash(v int) int {
	return v
}
