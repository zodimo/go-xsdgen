# Supported XSD Features

This document tracks the support status of XML Schema Definition (XSD) features in `go-xsdgen`. The feature list is derived from `docs/XMLSchema.xsd`.

**Legend:**
*   `[x]`: Feature is currently implemented in `go-xsdgen` logic.
*   `[ ]`: Feature is not yet implemented.
*   **(Inferable)**: This feature can be reliably inferred from analyzing partial XML instance data (fundamental types and structures).
*   *Features representing constraints (e.g., `choice`, `int`, `pattern`) are NOT marked inferable, as a subset of data cannot prove the absence of other values.*

## Feature Checklist

### Schema Directives
- [ ] `include`
- [ ] `import`
- [ ] `redefine`

### Top-Level Definitions
- [x] `schema` (Root element) **(Inferable)**
- [x] `element` (Global definition) **(Inferable)**
- [x] `attribute` (Global definition - *Inferred as local attributes*) **(Inferable)**
- [x] `complexType` (Global definition - *Inferred as anonymous or named*) **(Inferable)**
- [x] `simpleType` (Global definition - *Inferred as anonymous or named*) **(Inferable)**
- [ ] `group` (Named model group)
- [ ] `attributeGroup` (Named attribute group)
- [ ] `notation`

### Content Models & Particles
- [x] `element` (Local declaration) **(Inferable)**
- [ ] `group` (Group reference)
- [ ] `all`
- [ ] `choice`
- [x] `sequence` **(Inferable)** - *Default assumption for element ordering*
- [ ] `any`
- [ ] `anyAttribute`

### Complex Type Definitions
- [x] `complexType` **(Inferable)**
    - [x] `simpleContent` **(Inferable)**
    - [x] `complexContent` **(Inferable)**
- [ ] `extension` (Complex type derivation)
- [ ] `restriction` (Complex type derivation)
- [ ] `abstract` attributes
- [ ] `block` / `final` attributes
- [ ] `mixed` content processing

### Simple Type Definitions
- [x] `simpleType` **(Inferable)**
- [ ] `restriction` (Simple type derivation)
- [ ] `list`
- [ ] `union`

### Built-in Simple Types
- [ ] `anyType`
- [x] `string` **(Inferable)** - *The only safe assumption for data*
- [x] `boolean` **(Inferable)**
- [x] `decimal` **(Inferable)** - *General numeric type*
- [ ] `float`
- [ ] `double`
- [ ] `duration`
- [x] `dateTime` **(Inferable)**
- [ ] `time`
- [x] `date` **(Inferable)**
- [ ] `gYearMonth`
- [ ] `gYear`
- [ ] `gMonthDay`
- [ ] `gDay`
- [ ] `gMonth`
- [ ] `hexBinary`
- [ ] `base64Binary`
- [ ] `anyURI`
- [ ] `QName`
- [ ] `NOTATION`
- [ ] `normalizedString`
- [ ] `token`
- [ ] `language`
- [ ] `IDREFS`
- [ ] `ENTITIES`
- [ ] `NMTOKEN`
- [ ] `NMTOKENS`
- [ ] `Name`
- [ ] `NCName`
- [ ] `ID`
- [ ] `IDREF`
- [ ] `ENTITY`
- [x] `integer` **(Inferable)** - *General unbounded integer*
- [ ] `nonPositiveInteger`
- [ ] `negativeInteger`
- [ ] `long`
- [ ] `int`
- [ ] `short`
- [ ] `byte`
- [ ] `nonNegativeInteger`
- [ ] `unsignedLong`
- [ ] `unsignedInt`
- [ ] `unsignedShort`
- [ ] `unsignedByte`
- [ ] `positiveInteger`

### Facets (Constraining Facets)
- [ ] `length`
- [ ] `minLength`
- [ ] `maxLength`
- [ ] `pattern`
- [ ] `enumeration`
- [ ] `whiteSpace`
- [ ] `maxInclusive`
- [ ] `maxExclusive`
- [ ] `minInclusive`
- [ ] `minExclusive`
- [ ] `totalDigits`
- [ ] `fractionDigits`

### Identity Constraints
- [ ] `unique`
- [ ] `key`
- [ ] `keyref`
- [ ] `selector`
- [ ] `field`

### Documentation & Annotation
- [ ] `annotation`
- [ ] `appinfo`
- [ ] `documentation`

### Schema Attributes
- [ ] `targetNamespace` **(Inferable)**
- [ ] `elementFormDefault`
- [ ] `attributeFormDefault`
- [ ] `blockDefault`
- [ ] `finalDefault`
- [ ] `version`
