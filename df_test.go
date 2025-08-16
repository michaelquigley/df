package df

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "a"},
		{"A", "a"},
		{"lowercase", "lowercase"},
		{"CamelCase", "camel_case"},
		{"XMLHttpRequest", "xml_http_request"},
		{"HTMLParser", "html_parser"},
		{"UserID", "user_id"},
		{"HTTPSProxy", "https_proxy"},
		{"SimpleField", "simple_field"},
		{"ABC", "abc"},
		{"ID", "id"},
		{"getHTTPResponseCode", "get_http_response_code"},
		{"APIToken", "api_token"},
		{"URLName", "url_name"},
		{"JSONData", "json_data"},
		{"HTTPSURLParser", "httpsurl_parser"},
		{"IOError", "io_error"},
		{"PDFFile", "pdf_file"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := toSnakeCase(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestStripIndices(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"simple.path", "simple.path"},
		{"Root.Items[0]", "Root.Items"},
		{"Root.Items[0].Action", "Root.Items.Action"},
		{"Root.Items[0].Nested[1].Field", "Root.Items.Nested.Field"},
		{"Root.Items[123].Action", "Root.Items.Action"},
		{"Root[0]", "Root"},
		{"Items[0][1][2]", "Items"},
		{"Root.Items[0].Nested[99].Deep[1].Field", "Root.Items.Nested.Deep.Field"},
		{"NoIndices", "NoIndices"},
		{"Path.With.No.Indices", "Path.With.No.Indices"},
		{"[0]", ""},
		{"Root.[0]", "Root."},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := stripIndices(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

type nestedType struct {
	Name  string
	Count int
}

// test helpers for Dynamic
type dynA struct{ Name string }

func (d *dynA) Type() string          { return "a" }
func (d *dynA) ToMap() map[string]any { return map[string]any{"type": "a", "name": d.Name} }

type dynB struct{ Count int }

func (d *dynB) Type() string          { return "b" }
func (d *dynB) ToMap() map[string]any { return map[string]any{"type": "b", "count": d.Count} }
