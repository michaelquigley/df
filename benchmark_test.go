package df

import (
	"testing"
)

func BenchmarkToSnakeCase(b *testing.B) {
	testCases := []string{
		"SimpleCase",
		"VeryLongFieldNameWithManyUpperCaseLetters",
		"HTMLParser",
		"XMLHttpRequest",
		"getUserIDFromToken",
		"A",
		"ABC",
		"getHTTPStatusCode",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = toSnakeCase(tc)
		}
	}
}

func BenchmarkStripIndices(b *testing.B) {
	testCases := []string{
		"Root.Items[0].Action",
		"Container.Users[42].Profile.Settings[1].Value",
		"Simple.Path.Without.Indices",
		"Deep[0].Nested[1].Array[2].Access[3].Pattern[4]",
		"Mixed.Path[100].With.Some[999].Indices",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = stripIndices(tc)
		}
	}
}

// test to verify our string optimizations produce correct results
func TestStringOptimizationsCorrectness(t *testing.T) {
	// test toSnakeCase correctness
	testCases := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"A", "a"},
		{"ABC", "abc"},
		{"SimpleCase", "simple_case"},
		{"HTMLParser", "html_parser"},
		{"getUserID", "get_user_id"},
		{"XMLHttpRequest", "xml_http_request"},
	}

	for _, tc := range testCases {
		result := toSnakeCase(tc.input)
		if result != tc.expected {
			t.Errorf("toSnakeCase(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}

	// test stripIndices correctness
	stripTestCases := []struct {
		input    string
		expected string
	}{
		{"Root.Items[0].Action", "Root.Items.Action"},
		{"Container.Users[42].Profile", "Container.Users.Profile"},
		{"Simple.Path.Without.Indices", "Simple.Path.Without.Indices"},
		{"Deep[0].Nested[1].Array[2]", "Deep.Nested.Array"},
		{"", ""},
		{"NoIndices", "NoIndices"},
	}

	for _, tc := range stripTestCases {
		result := stripIndices(tc.input)
		if result != tc.expected {
			t.Errorf("stripIndices(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
