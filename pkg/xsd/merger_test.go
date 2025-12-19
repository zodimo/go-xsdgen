package xsd

import (
	"testing"
)

func TestMergeSchemas_Identical(t *testing.T) {
	s1 := &Schema{Root: NewElement("root")}
	s1.Root.SimpleType = TypeInt

	s2 := &Schema{Root: NewElement("root")}
	s2.Root.SimpleType = TypeInt

	merged, report, err := MergeSchemas(s1, s2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if merged.Root.SimpleType != TypeInt {
		t.Errorf("Expected TypeInt, got %s", merged.Root.SimpleType)
	}

	if len(report.FieldsAdded) > 0 || len(report.TypesGeneralized) > 0 || len(report.ConstraintsRelaxed) > 0 {
		t.Errorf("Expected empty report for identical schemas, got: %+v", report)
	}
}

func TestMergeSchemas_TypeGeneralization(t *testing.T) {
	s1 := &Schema{Root: NewElement("root")}
	s1.Root.SimpleType = TypeInt

	s2 := &Schema{Root: NewElement("root")}
	s2.Root.SimpleType = TypeString

	merged, report, err := MergeSchemas(s1, s2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if merged.Root.SimpleType != TypeString {
		t.Errorf("Expected TypeString (Int + String), got %s", merged.Root.SimpleType)
	}

	if len(report.TypesGeneralized) != 1 {
		t.Errorf("Expected 1 type generalization, got %d", len(report.TypesGeneralized))
	}
}

func TestMergeSchemas_NewAttribute(t *testing.T) {
	s1 := &Schema{Root: NewElement("root")}
	s1.Root.Attributes["attr1"] = &Attribute{Name: "attr1", Type: TypeInt, Required: true}

	s2 := &Schema{Root: NewElement("root")}
	// s2 has no attributes

	merged, report, err := MergeSchemas(s1, s2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if attr, ok := merged.Root.Attributes["attr1"]; !ok {
		t.Error("attr1 missing in merged schema")
	} else if attr.Required {
		t.Error("attr1 should be optional (missing in s2)")
	}

	foundConstraint := false
	for _, c := range report.ConstraintsRelaxed {
		if c == "root[@attr1]: present in s1 only -> optional" {
			foundConstraint = true
			break
		}
	}
	if !foundConstraint {
		t.Errorf("Expected constraint relaxation report for attr1, got: %v", report.ConstraintsRelaxed)
	}
}

func TestMergeSchemas_NewChild(t *testing.T) {
	s1 := &Schema{Root: NewElement("root")}

	s2 := &Schema{Root: NewElement("root")}
	child := NewElement("child")
	s2.Root.Children["child"] = child
	s2.Root.ChildrenOrder = append(s2.Root.ChildrenOrder, "child")

	merged, report, err := MergeSchemas(s1, s2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, ok := merged.Root.Children["child"]; !ok {
		t.Error("child missing in merged schema")
	}

	if merged.Root.Children["child"].MinOccurs != 0 {
		t.Errorf("child MinOccurs should be 0, got %d", merged.Root.Children["child"].MinOccurs)
	}

	foundField := false
	for _, f := range report.FieldsAdded {
		if f == "root.child" {
			foundField = true
			break
		}
	}
	if !foundField {
		t.Errorf("Expected field added report for child, got: %v", report.FieldsAdded)
	}
}
