package main

import (
	_ "embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/orderedmap"
)

// This is a simple tool to process the Swagger files in the "swagger" directory
// and extract the mapping of last path part to management system and the routes.
// It assumes the Swagger files are in the format used by the TMForum APIs.
// It will print the mapping and the routes to the standard output in JSON format.

var (
	managementToUpstream = map[string]string{}
	resourceToManagement = map[string]string{}
	resourceToFullPath   = map[string]string{}
)

func main() {

	// Visit all the JSON files in the "swagger" directory
	swaggerDir := "./www/oapiv4/oapiv4"

	// Read the directory entries
	dirEntries, err := os.ReadDir(swaggerDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading directory %s: %v\n", swaggerDir, err)
		os.Exit(1)
	}

	// Process each file in the directory
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() {
			// Process the file
			filePath := path.Join(swaggerDir, dirEntry.Name())
			if !strings.HasSuffix(filePath, ".json") {
				// Skip non-JSON files
				continue
			}
			processOneFile(filePath)
		}
	}

	if true {
		os.Exit(0)
	}

}

func processOneFile(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	document, err := libopenapi.NewDocument(content)
	if err != nil {
		panic(err)
	}

	v2Model, err := document.BuildV2Model()
	if err != nil {
		panic(err)
	}

	index := v2Model.Index
	fmt.Printf("There are %d paths and %d schemas in the document\n",
		len(index.GetAllPaths()), len(index.GetAllSchemas()))

	definitions := v2Model.Model.Definitions

	dd := definitions.Definitions
	pp := dd.GetPair("ProductOffering_Create")
	fmt.Println("=====>", pp.Key)

	for key, reference := range definitions.Definitions.FromNewest() {

		if key == "ProductOffering_Create" || key == "ProductOffering" || key == "ProductOffering_Update" {

			fmt.Println(key)
			schema := reference.Schema()
			fmt.Println("  Required:", schema.Required)

			properties := schema.Properties
			for key, _ := range properties.FromNewest() {
				fmt.Println("  ", key)

			}

		}
	}

	if true {
		os.Exit(0)
	}

	// get a count of the number of paths and schemas.
	pathItems := v2Model.Model.Paths.PathItems

	// print the number of paths and schemas in the document
	pathCount := orderedmap.Len(pathItems)
	fmt.Println()
	fmt.Println(filePath, pathCount)

	for _, pathItem := range pathItems.FromNewest() {
		ops := pathItem.GetOperations()
		if ops != nil {
			for op := range ops.ValuesFromNewest() {
				if op.OperationId != "createProductOffering" {
					continue
				}
				fmt.Println("  ", op.OperationId)
				opParameters := op.Parameters
				if opParameters == nil {
					fmt.Println("    No parameters")
					continue
				} else {
					fmt.Println("    Parameters:", len(opParameters))
				}

				for _, opParameter := range opParameters {
					if opParameter == nil {
						fmt.Println("    No parameter")
						continue
					}
					fmt.Println("    ", opParameter.Name, "in:", opParameter.In, "required:", *opParameter.Required)
					bodySchemaPtr := opParameter.Schema.Schema()
					if bodySchemaPtr == nil {
						fmt.Println("      No schema")
						continue
					}
					bodySchemaRef := opParameter.Schema.GetReference()
					fmt.Println("      Ref", bodySchemaRef)

					// We are here typically because of a body is required

					if len(bodySchemaPtr.Type) > 0 {
						fmt.Println("      Type", bodySchemaPtr.Type[0], "Title:", bodySchemaPtr.Title)
					}

					fmt.Println("      Required:", bodySchemaPtr.Required)
					fmt.Println("      Title:", bodySchemaPtr.Title)
					fmt.Println("      Description:", bodySchemaPtr.Description)

					bodyProperties := bodySchemaPtr.Properties
					if bodyProperties != nil {
						for key, bodyPropProxy := range bodyProperties.FromOldest() {
							bodyPropSchema := bodyPropProxy.Schema()
							if bodyPropSchema != nil {
								fmt.Printf("         %s\n", key)

							}
						}
					}

					// b, err := json.MarshalIndent(schemaPtr, "", "  ")
					// if err != nil {
					// 	fmt.Println("      Error marshalling schema:", err)
					// 	continue
					// }
					// fmt.Println("      ", string(b))

				}

			}

		}
	}

}
