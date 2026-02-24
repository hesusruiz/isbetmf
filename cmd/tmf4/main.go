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

	docmodel, err := document.BuildV2Model()
	if err != nil {
		panic(err)
	}

	// get a count of the number of paths and schemas.
	pathItems := docmodel.Model.Paths.PathItems

	// print the number of paths and schemas in the document
	len := orderedmap.Len(pathItems)
	fmt.Println(len)

	for _, pathItem := range pathItems.FromNewest() {
		ops := pathItem.GetOperations()
		for k, op := range ops.FromNewest() {
			fmt.Println(k, op.OperationId)
		}
	}

}
