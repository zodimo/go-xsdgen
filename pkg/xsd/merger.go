package xsd

import (
	"fmt"
)

// MergeReport contains details about what happened during the merge
type MergeReport struct {
	FieldsAdded        []string
	TypesGeneralized   []string
	ConstraintsRelaxed []string
}

// MergeSchemas combines two schemas into a more general one that validates both.
// It returns the merged schema and a report of changes.
func MergeSchemas(s1, s2 *Schema) (*Schema, *MergeReport, error) {
	if s1.Root.Name != s2.Root.Name {
		return nil, nil, fmt.Errorf("cannot merge schemas with different root elements: %s != %s", s1.Root.Name, s2.Root.Name)
	}

	report := &MergeReport{
		FieldsAdded:        make([]string, 0),
		TypesGeneralized:   make([]string, 0),
		ConstraintsRelaxed: make([]string, 0),
	}

	mergedRoot := mergeElements(s1.Root, s2.Root, report, "")

	return &Schema{Root: mergedRoot}, report, nil
}

func mergeElements(e1, e2 *Element, report *MergeReport, path string) *Element {
	currentPath := e1.Name
	if path != "" {
		currentPath = path + "." + e1.Name
	}

	// Create a new element for the result
	merged := NewElement(e1.Name)
	merged.Complex = e1.Complex || e2.Complex
	merged.MinOccurs = min(e1.MinOccurs, e2.MinOccurs)
	merged.MaxOccurs = max(e1.MaxOccurs, e2.MaxOccurs)
	merged.Nullable = e1.Nullable || e2.Nullable

	// Report Nullable change
	if merged.Nullable && (!e1.Nullable || !e2.Nullable) {
		report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s: Nullable false -> true", currentPath))
	}

	// If e1 has different min/max than merged (which is the relaxation of both), report it?
	// Actually, strict relaxation is: min = min(min1, min2), max = max(max1, max2)
	// If e1.Min=1, e2.Min=0 -> merged.Min=0. Constraint relaxed.
	if merged.MinOccurs < e1.MinOccurs || merged.MinOccurs < e2.MinOccurs {
		report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s: MinOccurs %d -> %d", currentPath, max(e1.MinOccurs, e2.MinOccurs), merged.MinOccurs))
	}
	// MaxOccurs: -1 is infinity.
	// If e1.Max=1, e2.Max=2 -> merged.Max=2.
	// Logic for max: if either is -1, result is -1. Else max(a,b).
	merged.MaxOccurs = mergeMaxOccurs(e1.MaxOccurs, e2.MaxOccurs)

	// Merge SimpleType
	oldType := e1.SimpleType
	merged.SimpleType = MergeSimpleTypes(e1.SimpleType, e2.SimpleType)
	if merged.SimpleType != oldType && e1.SimpleType != TypeUnknown {
		report.TypesGeneralized = append(report.TypesGeneralized, fmt.Sprintf("%s: %s + %s -> %s", currentPath, e1.SimpleType, e2.SimpleType, merged.SimpleType))
	}

	// Merge Attributes
	for name, attr1 := range e1.Attributes {
		if attr2, ok := e2.Attributes[name]; ok {
			// Attribute exists in both
			mergedAttr := &Attribute{
				Name:     name,
				Required: attr1.Required && attr2.Required,
				Type:     MergeSimpleTypes(attr1.Type, attr2.Type),
			}
			merged.Attributes[name] = mergedAttr

			if !mergedAttr.Required && (attr1.Required || attr2.Required) {
				report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s[@%s]: required -> optional", currentPath, name))
			}
			if mergedAttr.Type != attr1.Type {
				report.TypesGeneralized = append(report.TypesGeneralized, fmt.Sprintf("%s[@%s]: %s -> %s", currentPath, name, attr1.Type, mergedAttr.Type))
			}

		} else {
			// Attribute only in S1
			mergedAttr := &Attribute{
				Name:     name,
				Required: false, // Must be optional now
				Type:     attr1.Type,
			}
			merged.Attributes[name] = mergedAttr
			report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s[@%s]: present in s1 only -> optional", currentPath, name))
		}
	}

	for name, attr2 := range e2.Attributes {
		if _, ok := e1.Attributes[name]; !ok {
			// Attribute only in S2
			mergedAttr := &Attribute{
				Name:     name,
				Required: false,
				Type:     attr2.Type,
			}
			merged.Attributes[name] = mergedAttr
			report.FieldsAdded = append(report.FieldsAdded, fmt.Sprintf("%s[@%s]", currentPath, name))
		}
	}

	// Merge Children with intelligent ordering
	merged.ChildrenOrder = mergeChildOrder(e1.ChildrenOrder, e2.ChildrenOrder)
	for _, childName := range merged.ChildrenOrder {
		child1, in1 := e1.Children[childName]
		child2, in2 := e2.Children[childName]

		if in1 && in2 {
			merged.Children[childName] = mergeElements(child1, child2, report, currentPath)
		} else if in1 {
			mergedChild := cloneElement(child1)
			mergedChild.MinOccurs = 0
			merged.Children[childName] = mergedChild
			report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s.%s: present in s1 only -> optional", currentPath, childName))
		} else if in2 {
			mergedChild := cloneElement(child2)
			mergedChild.MinOccurs = 0
			merged.Children[childName] = mergedChild
			report.FieldsAdded = append(report.FieldsAdded, fmt.Sprintf("%s.%s", currentPath, childName))
		}
	}

	// Post-processing: Handle mixed-order duplicates
	// If the merge logic produced duplicates in ChildrenOrder (due to mixed element orders in inputs),
	// we have two options:
	// 1. Switch to <xs:all> (if maxOccurs constraints allow).
	// 2. Keep <xs:sequence> but make elements optional (MinOccurs=0).

	hasDuplicates := false
	counts := make(map[string]int)
	for _, name := range merged.ChildrenOrder {
		counts[name]++
		if counts[name] > 1 {
			hasDuplicates = true
		}
	}

	if hasDuplicates {
		// potential for xs:all?
		canUseAll := true
		for _, child := range merged.Children {
			// xs:all requires minOccurs=0|1 and maxOccurs=1
			if child.MaxOccurs > 1 || child.MaxOccurs == -1 {
				canUseAll = false
				break
			}
		}

		if canUseAll {
			// Switch to xs:all
			merged.Compositor = CompositorAll
			report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s: mixed order -> switched to xs:all", currentPath))

			// Deduplicate order
			uniqueOrder := make([]string, 0)
			seen := make(map[string]bool)
			for _, name := range merged.ChildrenOrder {
				if !seen[name] {
					uniqueOrder = append(uniqueOrder, name)
					seen[name] = true
				}
			}
			merged.ChildrenOrder = uniqueOrder

			// Ensure children are minOccurs=0 or 1.
			// (If they were duplicates, they likely should be optional in xs:all if strictly they weren't everywhere,
			// but xs:all inherently implies order independence.
			// Check logic: if it was present in S1 and S2, minOccurs might be 1.
			// If it was only in one, minOccurs=0. This is preserved.)
		} else {
			// Fallback to "messy" sequence
			for name, count := range counts {
				if count > 1 {
					if child, ok := merged.Children[name]; ok {
						if child.MinOccurs != 0 {
							child.MinOccurs = 0
							report.ConstraintsRelaxed = append(report.ConstraintsRelaxed, fmt.Sprintf("%s.%s: mixed order detected -> optional", currentPath, name))
						}
					}
				}
			}
		}
	}

	return merged
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mergeMaxOccurs(m1, m2 int) int {
	if m1 == -1 || m2 == -1 {
		return -1
	}
	return max(m1, m2)
}

func cloneElement(e *Element) *Element {
	// Deep copy needed to avoid modifying original schema
	// For simplicity, we can just return a new struct with same values
	// BUT recursive children need to be cloned too if we are going to modify them (e.g. setting MinOccurs)
	// Since we set MinOccurs=0 on the returned element immediately, a shallow copy of the struct is enough
	// IF we don't modify its pointer-fields.
	// However, we are putting it into a new tree.
	// Let's do a safe shallow copy of the struct, but we need to create a new map for Attributes/Children if we were to modify them.
	// In the use case above, we only modify MinOccurs on the struct itself.
	// But if we ever deep-merge down the line, we might have issues.
	// For now, this is "good enough" for the specific use case of "make optional".
	// But wait, if we modify 'mergedChild.MinOccurs=0', and 'mergedChild' points to the same children maps as 'e',
	// that's fine as long as we don't modify those children maps.

	newE := *e
	newE.Attributes = make(map[string]*Attribute)
	for k, v := range e.Attributes {
		// copy attribute
		attr := *v
		newE.Attributes[k] = &attr
	}
	newE.Children = make(map[string]*Element)
	for k, v := range e.Children {
		newE.Children[k] = cloneElement(v)
	}
	newE.ChildrenOrder = make([]string, len(e.ChildrenOrder))
	copy(newE.ChildrenOrder, e.ChildrenOrder)

	return &newE
}

// mergeChildOrder combines two lists of children names, preserving order and interleaving new elements.
func mergeChildOrder(base, other []string) []string {
	result := make([]string, 0, len(base)+len(other))
	result = append(result, base...)

	// We weave 'other' into 'base'.
	// This is a naive heuristic: if we find an element from 'other' already in 'result', we synchronize our position.
	// If not, we insert it at the current position.

	// BUT, inserting into a slice is expensive. Let's rebuild 'result' from scratch?
	// Or use a more robust "diff/patch" approach?

	// Let's try the "insertion cursor" approach on a copy of base.

	// Since we already copied 'base' to 'result', let's iterate 'other'.
	// We need to know where each element of 'result' is to check existence efficiently?
	// For small lists, linear scan is fine.

	insertIdx := 0
	for _, name := range other {
		// Does 'name' exist in 'result' *after* insertIdx?
		foundIdx := -1
		for i := insertIdx; i < len(result); i++ {
			if result[i] == name {
				foundIdx = i
				break
			}
		}

		if foundIdx != -1 {
			// Found it. Advance cursor to just after it.
			insertIdx = foundIdx + 1
		} else {
			// Not found (forward of current pos). Insert it at insertIdx.
			result = append(result[:insertIdx], append([]string{name}, result[insertIdx:]...)...)
			insertIdx++
		}
	}

	return result
}
