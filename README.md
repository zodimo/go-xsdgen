# XML Validation and Schema Generator

This tool allows you to generate an XSD (XML Schema Definition) file from an XML input. It analyzes the XML structure and infers data types, handling optional attributes and elements.

It supports three common XSD design patterns:
- **Russian Doll**: A single global element with nested definitions.
- **Salami Slice**: All elements are defined globally and referenced.
- **Venetian Blind**: All complex types are defined globally; elements use these types.

The tool also includes a validation command to verify an XML file against an XSD schema.

## Installation

Ensure you have Go installed (1.24+).

```bash
go install ./cmd/xsdgen
```

Ensure that `$GOPATH/bin` (or `$HOME/go/bin`) is in your `$PATH`.

Prerequisites for validation:
- `xmllint` (usually available on Linux/macOS via `libxml2-utils` or similar package).

## Usage

### Generate XSD

```bash
xsdgen generate -input <xml-file> -style <style> > output.xsd
```

Arguments:
- `-input`: Path to the input XML file (required).
- `-style`: Generation style. Options: `russian` (default), `salami`, `venetian`.

**Examples:**

```bash
# Generate using Russian Doll (default)
xsdgen generate -input data.xml > schema_russian.xsd

# Generate using Salami Slice
xsdgen generate -input data.xml -style salami > schema_salami.xsd

# Generate using Venetian Blind
xsdgen generate -input data.xml -style venetian > schema_venetian.xsd
```

### Validate XML

```bash
xsdgen validate -xml <xml-file> -xsd <xsd-file>
```

Arguments:
- `-xml`: Path to the XML file validation.
- `-xsd`: Path to the XSD schema file.

**Example:**

```bash
xsdgen validate -xml data.xml -xsd schema.xsd
```

## Features

- **Type Inference**: Automatically detects `int`, `decimal`, `boolean`, `date`, `dateTime`, and `string`.
- **structure Analysis**: Handles repeated elements (`maxOccurs="unbounded"`) and optional elements/attributes.
- **Multiple Styles**: Choose the XSD structure that fits your needs.