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
