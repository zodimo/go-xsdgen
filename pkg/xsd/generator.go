package xsd

import (
	"encoding/xml"
	"fmt"
	"io"
)

type Style string

const (
	StyleRussianDoll   Style = "russian"
	StyleSalamiSlice   Style = "salami"
	StyleVenetianBlind Style = "venetian"
)

func Generate(w io.Writer, schema *Schema, style Style) error {
	xsd := &XSDSchema{
		XmlnsXs:            "http://www.w3.org/2001/XMLSchema",
		ElementFormDefault: "qualified",
	}

	switch style {
	case StyleRussianDoll:
		generateRussianDoll(schema, xsd)
	case StyleSalamiSlice:
		generateSalamiSlice(schema, xsd)
	case StyleVenetianBlind:
		generateVenetianBlind(schema, xsd)
	default:
		return fmt.Errorf("unknown style: %s", style)
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	fmt.Fprintln(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	return enc.Encode(xsd)
}

func mapType(t SimpleType) string {
	switch t {
	case TypeInteger:
		return "xs:integer"
	case TypeFloat:
		return "xs:decimal"
	case TypeBool:
		return "xs:boolean"
	case TypeDate:
		return "xs:date"
	case TypeDateTime:
		return "xs:dateTime"
	default:
		return "xs:string"
	}
}

// Russian Doll: Nested definitions
func generateRussianDoll(s *Schema, xsd *XSDSchema) {
	if s.Root == nil {
		return
	}
	xsd.Elements = append(xsd.Elements, buildElementRD(s.Root))
}

func buildElementRD(el *Element) *XSDElement {
	xEl := &XSDElement{
		Name:      el.Name,
		MinOccurs: fmtOccurs(el.MinOccurs),
		MaxOccurs: fmtOccurs(el.MaxOccurs),
	}
	if el.Nullable {
		xEl.Nillable = "true"
	}

	if !el.Complex && len(el.Attributes) == 0 {
		xEl.Type = mapType(el.SimpleType)
	} else {
		ct := &XSDComplexType{}

		if el.Compositor == CompositorAll {
			ct.All = &XSDAll{}
			for _, childName := range el.ChildrenOrder {
				child := el.Children[childName]
				ct.All.Elements = append(ct.All.Elements, buildElementRD(child))
			}
		} else {
			ct.Sequence = &XSDSequence{}
			for _, childName := range el.ChildrenOrder {
				child := el.Children[childName]
				ct.Sequence.Elements = append(ct.Sequence.Elements, buildElementRD(child))
			}
		}

		// Attributes
		for _, attr := range el.Attributes {
			use := "optional"
			if attr.Required {
				use = "required"
			}
			ct.Attribs = append(ct.Attribs, &XSDAttribute{
				Name: attr.Name,
				Type: mapType(attr.Type),
				Use:  use,
			})
		}
		xEl.Complex = ct
	}
	return xEl
}

// Salami Slice: Global elements, refs
func generateSalamiSlice(s *Schema, xsd *XSDSchema) {
	if s.Root == nil {
		return
	}
	// We need to collect all unique elements flatly.
	// But names might clash if nested structure is different.
	// Valid Salami Slice usually assumes unique names or qualified names.
	// For simplicity, we assume element names are unique identifiers or we suffix them.
	// Better: standard approach is assuming unique elements globally or making them so.
	// I'll traverse and lift all definitions to global.

	globals := make(map[string]*XSDElement)

	// First pass: define global elements
	// In my implementation buildGlobalSalami recursively visits children and adds them to globals.
	// So we just need to start at the root.

	buildGlobalSalami(s.Root, globals, xsd)
}

func buildGlobalSalami(el *Element, globals map[string]*XSDElement, xsd *XSDSchema) {
	if _, exists := globals[el.Name]; exists {
		// Already defined.
		// Note: If distinct structure sharing same name, valid Salami requires them to be same type or name collision resolution.
		// We ignore collision/structural-mismatch for this simple version and assume same name = same structure.
		return
	}

	xEl := &XSDElement{Name: el.Name}
	if el.Nullable {
		xEl.Nillable = "true"
	}
	globals[el.Name] = xEl

	if !el.Complex && len(el.Attributes) == 0 {
		xEl.Type = mapType(el.SimpleType)
	} else {
		ct := &XSDComplexType{Sequence: &XSDSequence{}}
		for _, childName := range el.ChildrenOrder {
			child := el.Children[childName]

			// Ensure child is defined globally
			buildGlobalSalami(child, globals, xsd)

			// Ref it
			ref := &XSDElement{
				Ref:       child.Name,
				MinOccurs: fmtOccurs(child.MinOccurs),
				MaxOccurs: fmtOccurs(child.MaxOccurs),
			}
			ct.Sequence.Elements = append(ct.Sequence.Elements, ref)
		}
		for _, attr := range el.Attributes {
			use := "optional"
			if attr.Required {
				use = "required"
			}
			ct.Attribs = append(ct.Attribs, &XSDAttribute{
				Name: attr.Name,
				Type: mapType(attr.Type),
				Use:  use,
			})
		}
		xEl.Complex = ct
	}
	xsd.Elements = append(xsd.Elements, xEl)
}

// Venetian Blind: Global types, local elements
func generateVenetianBlind(s *Schema, xsd *XSDSchema) {
	if s.Root == nil {
		return
	}

	// Build global complex types.
	// Root element is defined globally but uses a global type.

	types := make(map[string]*XSDComplexType)
	rootType := buildTypeVenetian(s.Root, types, xsd)

	// Add Root Element
	xsd.Elements = append(xsd.Elements, &XSDElement{
		Name: s.Root.Name,
		Type: rootType, // Name of the complex type
	})
}

func buildTypeVenetian(el *Element, types map[string]*XSDComplexType, xsd *XSDSchema) string {
	typeName := el.Name + "Type"

	if !el.Complex && len(el.Attributes) == 0 {
		// It's a simple type, no complex type definition needed unless we want to wrap it.
		// Use standard simple type mapping for element, so return empty string -> indicating usages should use standard type?
		// But wait, Venetian puts everything in ComplexTypes usually if structure.
		// If it is simple, we don't define a ComplexType.
		return mapType(el.SimpleType)
	}

	if _, exists := types[typeName]; exists {
		return typeName
	}

	ct := &XSDComplexType{
		Name:     typeName,
		Sequence: &XSDSequence{},
	}
	types[typeName] = ct

	for _, childName := range el.ChildrenOrder {
		child := el.Children[childName]
		childType := buildTypeVenetian(child, types, xsd)

		elem := &XSDElement{
			Name:      child.Name,
			MinOccurs: fmtOccurs(child.MinOccurs),
			MaxOccurs: fmtOccurs(child.MaxOccurs),
		}
		if child.Nullable {
			elem.Nillable = "true"
		}

		// If it's a standard simple type (starts with xs:), use Type=
		// If it's a custom type, use Type=
		elem.Type = childType

		ct.Sequence.Elements = append(ct.Sequence.Elements, elem)
	}

	for _, attr := range el.Attributes {
		use := "optional"
		if attr.Required {
			use = "required"
		}
		ct.Attribs = append(ct.Attribs, &XSDAttribute{
			Name: attr.Name,
			Type: mapType(attr.Type),
			Use:  use,
		})
	}

	xsd.ComplexTypes = append(xsd.ComplexTypes, ct)
	return typeName
}

func fmtOccurs(n int) string {
	if n == 1 {
		return "" // Default
	}
	if n == -1 {
		return "unbounded"
	}
	return fmt.Sprintf("%d", n)
}
