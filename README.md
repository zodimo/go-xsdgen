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

### Merge Schemas

Merge multiple XML files into a single XSD.

```bash
xsdgen merge -style <style> <file1> <file2> ...
```

Arguments:
- `-style`: Generation style (russian, salami, venetian). Default: `russian`.
- `files`: Two or more XML files to merge.

**Example:**

```bash
xsdgen merge -style venetian data1.xml data2.xml > merged_schema.xsd
```

## Library Usage

The core logic is available in the `pkg/xsd` package, allowing programmatic access to schema analysis and merging.

### Schema Merging

You can merge two inferred schemas to create a generalized schema that validates both inputs. This is useful when learning from multiple XML samples.

```go
import "github.com/zodimo/go-xsdgen/pkg/xsd"

// ... inside your code ...
s1, _ := xsd.Analyze(reader1)
s2, _ := xsd.Analyze(reader2)

merged, report, err := xsd.MergeSchemas(s1, s2)
if err != nil {
    log.Fatal(err)
}

// report.FieldsAdded contains fields present in s2 but not s1
// report.TypesGeneralized contains types that were widened (e.g. int -> string)
```

## Features

- **Type Inference**: Automatically detects `int`, `decimal`, `boolean`, `date`, `dateTime`, and `string`.
- **Structure Analysis**: Handles repeated elements (`maxOccurs="unbounded"`) and optional elements/attributes.
- **Multiple Styles**: Choose the XSD structure that fits your needs.
- **Schema Merging**: Combine multiple inferred schemas into a single general definition with detailed change reporting.