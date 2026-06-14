package config

import (
	_ "embed"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/hesusruiz/isbetmf/internal/errl"
)

type UpstreamEntry struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

type UpstreamEntries map[string]UpstreamEntry

//go:embed proxy.yaml
var internalUpstreamYAMLContent []byte

var internalUpstreamEntries UpstreamEntries

//   catalog:
//     host: tm-forum-api-product-catalog
//     port: 8080
//     path: /tmf-api/productCatalogManagement/v4
//   inventory:
//     host: tm-forum-api-product-inventory
//     port: 8080
//     path: /tmf-api/productInventory/v4
//   ordering:
//     host: tm-forum-api-product-ordering-management
//     port: 8080
//     path: /tmf-api/productOrderingManagement/v4
//   billing:
//     host: tm-forum-api-account
//     port: 8080
//     path: /tmf-api/accountManagement/v4
//   usage:
//     host: tm-forum-api-usage-management
//     port: 8080
//     path: /tmf-api/usageManagement/v4
//   party:
//     host: tm-forum-api-party-catalog
//     port: 8080
//     path: /tmf-api/party/v4
//   customer:
//     host: tm-forum-api-customer-management
//     port: 8080
//     path: /tmf-api/customerManagement/v4
//   resources:
//     host: tm-forum-api-resource-catalog
//     port: 8080
//     path: /tmf-api/resourceCatalog/v4
//   services:
//     host: tm-forum-api-service-catalog
//     port: 8080
//     path: /tmf-api/serviceCatalogManagement/v4
//   resourceInventory:
//     host: tm-forum-api-resource-inventory
//     port: 8080
//     path: /tmf-api/resourceInventoryManagement/v4
//   serviceInventory:
//     host: tm-forum-api-service-inventory
//     port: 8080
//     path: /tmf-api/serviceInventory/v4

// InternalUpstreamURL returns a url of the form: http://hostname:port/path
func InternalUpstreamURL(resourceName string) (string, error) {

	// 	Parse the proxy yaml content only once at first time
	if len(internalUpstreamEntries) == 0 {
		err := yaml.Unmarshal(internalUpstreamYAMLContent, &internalUpstreamEntries)
		if err != nil {
			return "", errl.Errorf("error parsing proxy yaml content: %s", err)
		}
	}

	// Get the entry for the resource name
	entry, ok := internalUpstreamEntries[resourceName]
	if !ok {
		return "", errl.Errorf("unknown resource type: %s", resourceName)
	}

	// Build the URL, which is like: http://hostname:port/path
	url := fmt.Sprintf("http://%s:%d%s", entry.Host, entry.Port, entry.Path)
	return url, nil
}

// InternalOrigin returns a url of the form: "http://hostname:port"
// for the TMF server running as a sidecar.
func InternalOrigin(resourceName string) (string, error) {

	// 	Parse the proxy yaml content only once at first time
	if len(internalUpstreamEntries) == 0 {
		err := yaml.Unmarshal(internalUpstreamYAMLContent, &internalUpstreamEntries)
		if err != nil {
			return "", errl.Errorf("error parsing proxy yaml content: %s", err)
		}
	}

	// Get the entry for the resource name
	entry, ok := internalUpstreamEntries[resourceName]
	if !ok {
		return "", errl.Errorf("unknown resource type: %s", resourceName)
	}

	// the url is like: http://hostname:port
	url := fmt.Sprintf("http://%s:%d", entry.Host, entry.Port)
	return url, nil

}
