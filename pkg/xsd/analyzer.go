package xsd

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"time"
)

type stackItem struct {
	Element     *Element
	HasCharData bool
}

// Analyze parses the XML from the reader and returns a Schema definition
func Analyze(r io.Reader) (*Schema, error) {
	decoder := xml.NewDecoder(r)
	var root *Element
	var stack []*stackItem

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
				parentItem := stack[len(stack)-1]
				parent := parentItem.Element
				parent.Complex = true

				// Check if child already exists
				if existing, ok := parent.Children[name]; ok {
					current = existing
					current.Count++

					// If seen again in same parent instance, it's repeating
					if current.LastSeenInstance == parent.Count {
						current.MaxOccurs = -1
					}
					current.LastSeenInstance = parent.Count

				} else {
					current = NewElement(name)
					parent.Children[name] = current
					parent.ChildrenOrder = append(parent.ChildrenOrder, name)
					current.Count = 1
					current.LastSeenInstance = parent.Count

					// If this is a new child definition but parent has appeared before,
					// it means this child was missing in previous instances.
					if parent.Count > 1 {
						current.MinOccurs = 0
					}
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

			stack = append(stack, &stackItem{Element: current})

		case xml.EndElement:
			// When finishing a parent, check which children were NOT seen
			if len(stack) > 0 {
				item := stack[len(stack)-1]
				ended := item.Element

				// Identify empty elements (no CharData seen)
				// If an element is empty, it might be a String in disguise (or just optional content).
				// If strictly empty <tag/>, for XSD type inference, treating it as String helps
				// if other instances have text. If it was Int, and now Empty, it becomes String (Int + String = String).
				if !item.HasCharData {
					// Treat empty content as equivalent to empty string, which forces String type if combined with non-string,
					// or remains String if it was unknown.
					// Note: inferType("") returns TypeString usually, but let's be explicit.
					// If we specifically want empty to equate to "potential string", we merge with TypeString.
					if ended.SimpleType == TypeUnknown {
						ended.SimpleType = TypeString
					} else {
						// e.g. was Int, now empty -> Int + String = String
						ended.SimpleType = mergeTypes(ended.SimpleType, TypeString)
					}
				}

				// Check missing children in this instance
				for _, child := range ended.Children {
					if child.LastSeenInstance != ended.Count {
						child.MinOccurs = 0
					}
				}
			}
			stack = stack[:len(stack)-1]

		case xml.CharData:
			val := strings.TrimSpace(string(se))
			if len(stack) > 0 {
				item := stack[len(stack)-1]
				// Only consider it data if it's not just whitespace
				// (XML non-ignorable whitespace is tricky, but TrimSpace is standard for data-oriented XML)
				if len(val) > 0 {
					item.HasCharData = true
					current := item.Element
					inferred := inferType(val)
					if current.SimpleType == TypeUnknown {
						current.SimpleType = inferred
					} else {
						current.SimpleType = mergeTypes(current.SimpleType, inferred)
					}
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
	if n, err := strconv.ParseInt(val, 10, 64); err == nil {
		// Check range for xs:int (32-bit signed)
		// Min: -2147483648, Max: 2147483647
		// XSD int is implementation dependent but typically 32-bit.
		// Safe bet: if it fits in 32-bit, call it int. Else long.
		const minInt32 = -2147483648
		const maxInt32 = 2147483647
		if n >= minInt32 && n <= maxInt32 {
			return TypeInt
		}
		return TypeLong
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		return TypeFloat
	}
	if strings.ToLower(val) == "true" || strings.ToLower(val) == "false" {
		return TypeBool
	}
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

	// Int/Long Hierarchy
	if (t1 == TypeInt && t2 == TypeLong) || (t1 == TypeLong && t2 == TypeInt) {
		return TypeLong
	}

	// Numeric -> Decimal
	// Int/Long + Float -> Float (Decimal)
	isNumeric1 := (t1 == TypeInt || t1 == TypeLong || t1 == TypeFloat)
	isNumeric2 := (t2 == TypeInt || t2 == TypeLong || t2 == TypeFloat)
	if isNumeric1 && isNumeric2 {
		return TypeFloat
	}

	// Fallback to string for mismatched types (e.g. Bool + Int, Int + String)
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
		finalizeStructure(child)
	}
}
