package xsd

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"time"
)

// Analyze parses the XML from the reader and returns a Schema definition
func Analyze(r io.Reader) (*Schema, error) {
	decoder := xml.NewDecoder(r)
	var root *Element
	var stack []*Element

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch se := t.(type) {
		case xml.StartElement:
			name := se.Name.Local
			var current *Element

			if len(stack) == 0 {
				root = NewElement(name)
				current = root
				current.Count++
			} else {
				parent := stack[len(stack)-1]
				parent.Complex = true

				// Check if child already exists
				if existing, ok := parent.Children[name]; ok {
					current = existing
					current.MaxOccurs = -1 // If seen again, it can occur multiple times
					current.Count++
				} else {
					current = NewElement(name)
					parent.Children[name] = current
					parent.ChildrenOrder = append(parent.ChildrenOrder, name)
					current.Count = 1
				}
			}

			// Process Attributes
			for _, attr := range se.Attr {
				attrName := attr.Name.Local
				if _, exists := current.Attributes[attrName]; !exists {
					current.Attributes[attrName] = &Attribute{
						Name:  attrName,
						Type:  inferType(attr.Value),
						Count: 1,
					}
				} else {
					// Update existing attribute stats
					existing := current.Attributes[attrName]
					existing.Count++
					existing.Type = mergeTypes(existing.Type, inferType(attr.Value))
				}
			}

			// Mark missing attributes in previous instances as not required?
			// Actually, "Required" logic needs to be checked against Element.Count.
			// We can do a post-processing pass or check at the end. For now, track counts.

			stack = append(stack, current)

		case xml.EndElement:
			stack = stack[:len(stack)-1]

		case xml.CharData:
			val := strings.TrimSpace(string(se))
			if len(val) > 0 && len(stack) > 0 {
				current := stack[len(stack)-1]
				inferred := inferType(val)
				if current.SimpleType == TypeUnknown {
					current.SimpleType = inferred
				} else {
					current.SimpleType = mergeTypes(current.SimpleType, inferred)
				}
			}
		}
	}

	if root != nil {
		finalizeStructure(root)
	}

	return &Schema{Root: root}, nil
}

func inferType(val string) SimpleType {
	if _, err := strconv.Atoi(val); err == nil {
		return TypeInt
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return TypeFloat
	}
	if strings.ToLower(val) == "true" || strings.ToLower(val) == "false" {
		return TypeBool
	}
	// Try dates
	if _, err := time.Parse("2006-01-02", val); err == nil {
		return TypeDate
	}
	if _, err := time.Parse(time.RFC3339, val); err == nil {
		return TypeDateTime
	}
	return TypeString
}

func mergeTypes(t1, t2 SimpleType) SimpleType {
	if t1 == t2 {
		return t1
	}
	if t1 == TypeUnknown {
		return t2
	}
	if t2 == TypeUnknown {
		return t1
	}
	// If mixed numeric types, promote to float
	if (t1 == TypeInt && t2 == TypeFloat) || (t1 == TypeFloat && t2 == TypeInt) {
		return TypeFloat
	}
	// Fallback to string for mismatched types
	return TypeString
}

func finalizeStructure(el *Element) {
	// Update attribute requirements
	for _, attr := range el.Attributes {
		if attr.Count == el.Count {
			attr.Required = true
		} else {
			attr.Required = false
		}
	}

	// Update children occurrences
	// Determining minOccurs is tricky without tracking every instance sequence.
	// A simple heuristic: if Child.Count < Parent.Count, then MinOccurs = 0.
	// But valid list elements might appear multiple times per parent.
	// We really need to know "did this child appear in *every* instance of parent?".
	// For now, let's assume if it exists it's MinOccurs=1 unless we detect otherwise, but avoiding tracking *instance* IDs makes this hard.
	// Simplified logic: If Child.Count < Parent.Count, it definitely was missing in some parents (unless it can repeat, which makes it complicated).
	// Let's stick to safe defaults: MinOccurs=0 if we aren't strict.
	// Actually, standard inference often defaults to minOccurs=0 to be safe, or 1.
	// Let's refine:
	// If we tracked "PresenceCount" per parent instance it would be better.
	// Given the stream usage, `el.Count` is total occurrences of `el`.
	// The parent of `el` (let's say `p`) has `p.Count`.
	// If `el` is a child of `p`, `el.Count` could be >= `p.Count` if maxOccurs > 1.
	// If `el.Count` < `p.Count`, it MUST be optional (MinOccurs=0).

	// BUT, we don't have back-pointer to parent here easily during recursion without passing it.
	// Let's accept that limitation for this iteration and default MinOccurs=0 which is safe, or 1.
	// XML Schema default is MinOccurs=1.

	for _, child := range el.Children {
		// Heuristic: If total child count < parent count, it must be optional.
		if child.Count < el.Count {
			child.MinOccurs = 0
		} else {
			// It might be 1 or more. Default to 1.
		}
		finalizeStructure(child)
	}
}
