package xsd

import (
	"strings"
	"testing"
)

func TestAnalyzerOrdering(t *testing.T) {
	// Case 1: Interleaved new elements should be inserted correctly
	// First instance: <A> <C>
	// Second instance: <A> <B> <C>
	// Expected Order: A, B, C (B inserted between A and C)
	xmlData := `<Root>
		<Item>
			<A>1</A>
			<C>3</C>
		</Item>
		<Item>
			<A>1</A>
			<B>2</B>
			<C>3</C>
		</Item>
	</Root>`

	r := strings.NewReader(xmlData)
	schema, err := Analyze(r)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	root := schema.Root
	if root == nil {
		t.Fatal("Root is nil")
	}

	item, ok := root.Children["Item"]
	if !ok {
		t.Fatal("Item child not found")
	}

	expectedOrder := []string{"A", "B", "C"}
	if len(item.ChildrenOrder) != len(expectedOrder) {
		t.Errorf("Expected %d children, got %d", len(expectedOrder), len(item.ChildrenOrder))
	}

	for i, name := range expectedOrder {
		if i < len(item.ChildrenOrder) && item.ChildrenOrder[i] != name {
			t.Errorf("Order mismatch at index %d: expected %s, got %s. Full list: %v", i, name, item.ChildrenOrder[i], item.ChildrenOrder)
		}
	}
}

func TestInferType(t *testing.T) {
	tests := []struct {
		val      string
		expected SimpleType
	}{
		{"123", TypeInteger},
		{"0", TypeInteger},
		{"-123", TypeInteger},
		{"2147483647", TypeInteger},
		{"2147483648", TypeInteger}, // Overflow int32
		{"0123", TypeString},        // Leading zero -> String
		{"007", TypeString},         // Leading zeros -> String
		{"0.123", TypeFloat},        // Leading zero but decimal -> Float
		{"12.34", TypeFloat},
		{"true", TypeBool},
		{"FALSE", TypeBool},
		{"2023-12-25", TypeDate},
		{"2023-12-25T12:00:00Z", TypeDateTime},
		{"some string", TypeString},
		{"", TypeString},
	}

	for _, tt := range tests {
		got := inferType(tt.val)
		if got != tt.expected {
			t.Errorf("inferType(%q) = %v; want %v", tt.val, got, tt.expected)
		}
	}
}

func TestAnalyzerEmptyElement(t *testing.T) {
	// Case 1: Int then Empty -> Should stay Int but Nullable
	xmlData1 := `<Root><Val>123</Val><Val></Val></Root>`
	r1 := strings.NewReader(xmlData1)
	schema1, err := Analyze(r1)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	val1 := schema1.Root.Children["Val"]
	if val1.SimpleType != TypeInteger {
		t.Errorf("Case 1: Expected TypeInteger, got %v", val1.SimpleType)
	}
	if !val1.Nullable {
		t.Errorf("Case 1: Expected Nullable=true")
	}

	// Case 2: Empty then Int -> Should be Int and Nullable (Order independence)
	xmlData2 := `<Root><Val></Val><Val>123</Val></Root>`
	r2 := strings.NewReader(xmlData2)
	schema2, err := Analyze(r2)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	val2 := schema2.Root.Children["Val"]
	if val2.SimpleType != TypeInteger {
		t.Errorf("Case 2: Expected TypeInteger, got %v", val2.SimpleType)
	}
	if !val2.Nullable {
		t.Errorf("Case 2: Expected Nullable=true")
	}

	// Case 3: Just Empty -> default to String, Nullable
	xmlData3 := `<Root><Val></Val></Root>`
	r3 := strings.NewReader(xmlData3)
	schema3, err := Analyze(r3)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}
	val3 := schema3.Root.Children["Val"]
	if val3.SimpleType != TypeString && val3.SimpleType != TypeUnknown {
		t.Errorf("Case 3: Expected TypeString or Unknown, got %v", val3.SimpleType)
	}
	if !val3.Nullable {
		t.Errorf("Case 3: Expected Nullable=true")
	}
}
