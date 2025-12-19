package xsd

// SimpleType represents the inferred XSD data type
type SimpleType string

const (
	TypeString   SimpleType = "string"
	TypeInt      SimpleType = "int"
	TypeLong     SimpleType = "long"
	TypeFloat    SimpleType = "decimal"
	TypeBool     SimpleType = "boolean"
	TypeDateTime SimpleType = "dateTime"
	TypeDate     SimpleType = "date"
	TypeUnknown  SimpleType = "unknown"
)

// Attribute represents an XML attribute
type Attribute struct {
	Name     string
	Type     SimpleType
	Required bool
	Count    int // How many times this attribute was seen (to determine if required)
}

// Element represents an XML element
type Element struct {
	Name             string
	SimpleType       SimpleType
	Complex          bool
	Attributes       map[string]*Attribute
	Children         map[string]*Element
	ChildrenOrder    []string // To maintain order of children
	MinOccurs        int
	MaxOccurs        int // -1 for unbounded
	Count            int // How many times this element was seen in the current context
	LastSeenInstance int // ID of the parent instance where this child was last seen
}

// Schema represents the root of the analyzed XML structure
type Schema struct {
	Root *Element
}

func NewElement(name string) *Element {
	return &Element{
		Name:       name,
		SimpleType: TypeUnknown,
		Attributes: make(map[string]*Attribute),
		Children:   make(map[string]*Element),
		MinOccurs:  1,
		MaxOccurs:  1,
	}
}
