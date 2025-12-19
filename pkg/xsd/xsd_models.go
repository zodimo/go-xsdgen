package xsd

import "encoding/xml"

// XSDSchema is the root xs:schema element
type XSDSchema struct {
	XMLName            xml.Name          `xml:"xs:schema"`
	XmlnsXs            string            `xml:"xmlns:xs,attr"`
	ElementFormDefault string            `xml:"elementFormDefault,attr,omitempty"`
	Elements           []*XSDElement     `xml:"xs:element,omitempty"`
	ComplexTypes       []*XSDComplexType `xml:"xs:complexType,omitempty"`
	SimpleTypes        []*XSDSimpleType  `xml:"xs:simpleType,omitempty"`
}

type XSDElement struct {
	Name      string          `xml:"name,attr,omitempty"`
	Type      string          `xml:"type,attr,omitempty"`
	MinOccurs string          `xml:"minOccurs,attr,omitempty"`
	MaxOccurs string          `xml:"maxOccurs,attr,omitempty"`
	Nillable  string          `xml:"nillable,attr,omitempty"`
	Ref       string          `xml:"ref,attr,omitempty"`
	Complex   *XSDComplexType `xml:"xs:complexType,omitempty"`
}

type XSDComplexType struct {
	Name     string          `xml:"name,attr,omitempty"`
	Sequence *XSDSequence    `xml:"xs:sequence,omitempty"`
	All      *XSDAll         `xml:"xs:all,omitempty"`
	Attribs  []*XSDAttribute `xml:"xs:attribute,omitempty"`
}

type XSDSequence struct {
	Elements []*XSDElement `xml:"xs:element"`
}

type XSDAll struct {
	Elements []*XSDElement `xml:"xs:element"`
}

type XSDAttribute struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Use  string `xml:"use,attr,omitempty"` // required or optional
}

type XSDSimpleType struct {
	Name string `xml:"name,attr"`
	// Simplified... potentially restriction
}
