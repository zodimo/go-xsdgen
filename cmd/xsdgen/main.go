package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/zodimo/go-xsdgen/pkg/xsd"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	inputFlag := generateCmd.String("input", "", "Path to input XML file")
	styleFlag := generateCmd.String("style", "russian", "XSD generation style: russian, salami, venetian")

	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)
	valXmlFlag := validateCmd.String("xml", "", "Path to XML file")
	valXsdFlag := validateCmd.String("xsd", "", "Path to XSD file")

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		if *inputFlag == "" {
			generateCmd.PrintDefaults()
			os.Exit(1)
		}

		file, err := os.Open(*inputFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		schema, err := xsd.Analyze(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing XML: %v\n", err)
			os.Exit(1)
		}

		var style xsd.Style
		switch *styleFlag {
		case "russian":
			style = xsd.StyleRussianDoll
		case "salami":
			style = xsd.StyleSalamiSlice
		case "venetian":
			style = xsd.StyleVenetianBlind
		default:
			fmt.Fprintf(os.Stderr, "Unknown style: %s\n", *styleFlag)
			os.Exit(1)
		}

		err = xsd.Generate(os.Stdout, schema, style)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating XSD: %v\n", err)
			os.Exit(1)
		}

	case "validate":
		validateCmd.Parse(os.Args[2:])
		if *valXmlFlag == "" || *valXsdFlag == "" {
			validateCmd.PrintDefaults()
			os.Exit(1)
		}

		err := xsd.Validate(*valXmlFlag, *valXsdFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		fmt.Println("XML is valid against the XSD.")

	case "merge":
		mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
		styleFlag := mergeCmd.String("style", "russian", "XSD generation style: russian, salami, venetian")

		mergeCmd.Parse(os.Args[2:])
		inputFiles := mergeCmd.Args()

		if len(inputFiles) < 2 {
			fmt.Fprintln(os.Stderr, "Error: at least two input XML files are required for merging")
			fmt.Fprintln(os.Stderr, "Usage: xsdgen merge [flags] <xml-file1> <xml-file2> [more-xml-files...]")
			mergeCmd.PrintDefaults()
			os.Exit(1)
		}

		// Process first file
		f1, err := os.Open(inputFiles[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", inputFiles[0], err)
			os.Exit(1)
		}
		schema, err := xsd.Analyze(f1)
		f1.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", inputFiles[0], err)
			os.Exit(1)
		}
		if schema.Root != nil && schema.Root.Name == "schema" {
			fmt.Fprintf(os.Stderr, "Warning: Input file '%s' appears to be an XSD Schema (root element 'schema'). 'xsdgen merge' expects XML data files to infer the schema, not XSD files.\n", inputFiles[0])
		}

		// Merge subsequent files
		for i := 1; i < len(inputFiles); i++ {
			f, err := os.Open(inputFiles[i])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", inputFiles[i], err)
				os.Exit(1)
			}
			s2, err := xsd.Analyze(f)
			f.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error analyzing %s: %v\n", inputFiles[i], err)
				os.Exit(1)
			}
			if s2.Root != nil && s2.Root.Name == "schema" {
				fmt.Fprintf(os.Stderr, "Warning: Input file '%s' appears to be an XSD Schema (root element 'schema'). 'xsdgen merge' expects XML data files to infer the schema, not XSD files.\n", inputFiles[i])
			}

			merged, report, err := xsd.MergeSchemas(schema, s2)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error merging %s: %v\n", inputFiles[i], err)
				os.Exit(1)
			}

			// Print Report
			fmt.Fprintf(os.Stderr, "Merge Report for %s:\n", inputFiles[i])
			if len(report.FieldsAdded) > 0 {
				fmt.Fprintln(os.Stderr, "  Fields Added:")
				for _, f := range report.FieldsAdded {
					fmt.Fprintf(os.Stderr, "    - %s\n", f)
				}
			}
			if len(report.TypesGeneralized) > 0 {
				fmt.Fprintln(os.Stderr, "  Types Generalized:")
				for _, t := range report.TypesGeneralized {
					fmt.Fprintf(os.Stderr, "    - %s\n", t)
				}
			}
			if len(report.ConstraintsRelaxed) > 0 {
				fmt.Fprintln(os.Stderr, "  Constraints Relaxed:")
				for _, c := range report.ConstraintsRelaxed {
					fmt.Fprintf(os.Stderr, "    - %s\n", c)
				}
			}
			fmt.Fprintln(os.Stderr, "--------------------------------------------------")

			schema = merged
		}

		// Generate Output
		var style xsd.Style
		switch *styleFlag {
		case "russian":
			style = xsd.StyleRussianDoll
		case "salami":
			style = xsd.StyleSalamiSlice
		case "venetian":
			style = xsd.StyleVenetianBlind
		default:
			fmt.Fprintf(os.Stderr, "Unknown style: %s\n", *styleFlag)
			os.Exit(1)
		}

		err = xsd.Generate(os.Stdout, schema, style)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating XSD: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: xsdgen <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  generate  Generate XSD from XML")
	fmt.Println("  validate  Validate XML against XSD")
	fmt.Println("  merge     Merge multiple XML files into one XSD")
}
