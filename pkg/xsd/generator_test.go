package xsd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleXML = `<catalog year="2023">
	<book id="bk101">
		<author>Gambardella, Matthew</author>
		<title>XML Developer's Guide</title>
		<genre>Computer</genre>
		<price>44.95</price>
		<publish_date>2000-10-01</publish_date>
		<description>An in-depth look at creating applications 
		with XML.</description>
	</book>
	<book id="bk102">
		<author>Ralls, Kim</author>
		<title>Midnight Rain</title>
		<genre>Fantasy</genre>
		<price>5.95</price>
		<publish_date>2000-12-16</publish_date>
		<description>A former architect battles corporate zombies, 
		an evil sorceress, and her own childhood to become queen 
		of the world.</description>
	</book>
</catalog>`

func TestGenerator(t *testing.T) {
	// 1. Analyze the sample XML
	r := strings.NewReader(sampleXML)
	schema, err := Analyze(r)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	// Prepare temp dir for validation files
	tmpDir, err := os.MkdirTemp("", "xsdgen_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	xmlPath := filepath.Join(tmpDir, "sample.xml")
	if err := os.WriteFile(xmlPath, []byte(sampleXML), 0644); err != nil {
		t.Fatalf("Failed to write sample xml: %v", err)
	}

	tests := []struct {
		name  string
		style Style
		check func(*testing.T, string) // optional specific checks on content
	}{
		{
			name:  "RussianDoll",
			style: StyleRussianDoll,
			check: func(t *testing.T, content string) {
				if strings.Contains(content, "complexType name=") {
					t.Error("Russian Doll should not have named complex types (mostly)")
				}
			},
		},
		{
			name:  "SalamiSlice",
			style: StyleSalamiSlice,
			check: func(t *testing.T, content string) {
				// Should have multiple top-level elements (catalog, book, author...)
				// We expect <xs:element name="book"> at top level
				if !strings.Contains(content, "<xs:element name=\"book\"") {
					t.Error("Salami Slice should define global elements")
				}
			},
		},
		{
			name:  "VenetianBlind",
			style: StyleVenetianBlind,
			check: func(t *testing.T, content string) {
				// Should have named complex types
				if !strings.Contains(content, "<xs:complexType name=\"bookType\"") {
					t.Error("Venetian Blind should define named complex types")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			// 2. Generate XSD
			if err := Generate(&buf, schema, tt.style); err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			output := buf.String()

			// 3. Check specific style requirements
			tt.check(t, output)

			// 4. Validate
			xsdPath := filepath.Join(tmpDir, tt.name+".xsd")
			if err := os.WriteFile(xsdPath, []byte(output), 0644); err != nil {
				t.Fatalf("Failed to write XSD: %v", err)
			}

			// Validate expects file paths
			if err := Validate(xmlPath, xsdPath); err != nil {
				t.Errorf("Validation failed for style %s: %v", tt.name, err)
			}
		})
	}
}
