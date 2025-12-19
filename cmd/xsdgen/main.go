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

		err = xsd.Generate(schema, style)
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
}
