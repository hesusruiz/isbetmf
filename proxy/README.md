# TMForum Reporting Package

The TMForum Reporting package provides functionality to connect to remote TMForum servers, retrieve objects of various types, validate them against requirements, and generate comprehensive reports.

## Features

- **Remote Server Connection**: Connect to any TMForum-compliant server with configurable base URL
- **Object Retrieval**: Retrieve objects of configurable types (productOffering, productSpecification, etc.)
- **Smart Pagination**: Automatic pagination support to retrieve all objects efficiently (100 per page by default)
- **Validation**: Validate objects for required fields and related party requirements
- **Comprehensive Reporting**: Generate detailed Markdown reports with statistics and error details
- **Configurable**: Support for configuration files, environment variables, and command-line options
- **Progress Tracking**: Optional progress reporting for long-running operations

## Architecture

The package is organized into several components:

- **Config**: Configuration management with environment variable support
- **Client**: HTTP client for connecting to TMForum servers with automatic path prefix resolution
- **Validator**: Object validation against requirements
- **Reporter**: Report generation in Markdown format
- **Reporter**: Main orchestrator that coordinates all components
- **Routes**: Automatic path prefix mapping for different resource types

## Installation

```bash
go get github.com/hesusruiz/isbetmf/reporting
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/hesusruiz/isbetmf/reporting"
)

func main() {
    // Create configuration
    config := reporting.DefaultConfig()
    config.BaseURL = "https://tmf.example.com"
    
    // Create proxy instance
    proxyInstance, err := reporting.NewProxy(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Run validation
    ctx := context.Background()
    if err := proxyInstance.Run(ctx); err != nil {
        log.Fatal(err)
    }
}
```

### Command Line Interface

A command-line interface is provided in `cmd/tmfproxy/`:

```bash
# Build the binary
go build -o tmfproxy cmd/tmfproxy/main.go

# Run with basic options
./tmfproxy -base-url "https://tmf.example.com"

# Run with progress tracking
./tmfproxy -base-url "https://tmf.example.com" -progress

# Run with custom object types
./tmfproxy -base-url "https://tmf.example.com" -object-types "productOffering,productSpecification"

# Show help
./tmfproxy -help
```

## Configuration

### Configuration File

Create a configuration file (YAML or JSON):

```yaml
# config.yaml
base_url: "https://tmf.example.com"
timeout: 30
object_types:
  - "productOffering"
  - "productSpecification"
validate_required_fields: true
validate_related_party: true
output_dir: "./reports"
report_file: "validation_report.md"
```

### Environment Variables

Set environment variables to override configuration:

```bash
export TMF_BASE_URL="https://tmf.example.com"
export TMF_TIMEOUT="60"
export TMF_OBJECT_TYPES="productOffering,productSpecification"
export TMF_OUTPUT_DIR="./custom_reports"
export TMF_REPORT_FILE="custom_report.md"
export TMF_PAGINATION_ENABLED="true"
export TMF_PAGE_SIZE="100"
export TMF_MAX_OBJECTS="10000"
```

### Programmatic Configuration

```go
config := &proxy.Config{
    BaseURL:               "https://tmf.example.com",
    Timeout:               30,
    ObjectTypes:           []string{"productOffering", "productSpecification"},
    ValidateRequiredFields: true,
    ValidateRelatedParty:  true,
    OutputDir:             "./reports",
    ReportFile:            "report.md",
}

// Load from environment variables
config.LoadConfigFromEnv()

// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal(err)
}
```

## Object Types

The following object types are supported by default:

- `productOffering`
- `productSpecification`
- `productOfferingPrice`
- `category`
- `individual`
- `organization`
- `productCatalog`
- `customer`
- `product`
- `service`

Each object type automatically maps to the correct TMForum API endpoint using predefined path prefixes. The path prefixes are automatically generated and maintained by a separate tool, ensuring compatibility with the latest TMForum specifications.

You can customize this list in your configuration, but all object types must exist in the routes map to be processed.

## URL Building

The proxy automatically constructs the correct URLs for each object type using predefined path prefixes. For example:

- `productOffering` → `/tmf-api/productCatalogManagement/v4/productOffering`
- `individual` → `/tmf-api/party/v4/individual`
- `category` → `/tmf-api/productCatalogManagement/v4/category`

The path prefixes are automatically generated and maintained by a separate tool, ensuring compatibility with the latest TMForum specifications. If an unknown object type is specified, the proxy will return an error.

## Pagination

The proxy automatically handles pagination to retrieve all objects from TMForum servers:

- **Default Page Size**: 100 objects per page (configurable)
- **Automatic Detection**: Stops when fewer objects than the page size are returned
- **Efficient Retrieval**: Uses `limit` and `offset` parameters in API calls
- **Safety Limits**: Configurable maximum objects per type (default: 10,000)
- **Configurable**: Can be disabled or customized via configuration

Example pagination flow:
1. Request: `?limit=100&offset=0` → Get first 100 objects
2. Request: `?limit=100&offset=100` → Get next 100 objects
3. Request: `?limit=100&offset=200` → Get next 100 objects
4. Continue until fewer than 100 objects are returned

## Validation Rules

### Required Fields

All objects must have the following fields:
- `id`: Unique identifier
- `href`: Resource reference
- `lastUpdate`: Last update timestamp
- `version`: Object version

### Related Party Requirements

Objects must include related party information with the following roles:
- `Seller`: The selling party
- `SellerOperator`: The operator responsible for selling

## Report Output

Reports are generated in Markdown format and include:

- **Summary Statistics**: Total objects, valid/invalid counts, errors, warnings
- **Statistics by Object Type**: Breakdown by each object type
- **Error Summary**: Count of each error type
- **Warning Summary**: Count of each warning type
- **Detailed Results**: Individual validation results for each object

### Report Structure

```
# TMForum Object Validation Report

## Summary Statistics
| Metric | Value |
|--------|-------|
| Total Objects | 150 |
| Valid Objects | 142 |
| Invalid Objects | 8 |
| Total Errors | 12 |
| Total Warnings | 25 |

## Statistics by Object Type
| Object Type | Count | Valid | Invalid | Errors | Warnings |
|-------------|-------|-------|---------|--------|----------|
| productOffering | 50 | 48 | 2 | 3 | 8 |

## Error Summary
| Error Code | Count |
|-------------|-------|
| MISSING_REQUIRED_FIELD | 8 |
| MISSING_RELATED_PARTY | 4 |

## Detailed Validation Results
### productOffering Objects
#### Object: PO-001
- **Type:** ProductOffering
- **Valid:** true
- **Timestamp:** 2024-01-15 10:30:00 UTC
```

## Error Codes

### Validation Errors

- `MISSING_REQUIRED_FIELD`: Required field is missing
- `MISSING_RELATED_PARTY`: Related party information is missing
- `UNKNOWN_TYPE`: Object type is not recognized

### Validation Warnings

- `MISSING_REQUIRED_ROLE`: Required related party role is missing
- `EMPTY_ROLE`: Related party role is empty
- `MISSING_PARTY_ID`: Related party ID is missing
- `MISSING_PARTY_HREF`: Related party href is missing

## Advanced Usage

### Progress Tracking

```go
progressChan := make(chan proxy.ProgressUpdate)

go func() {
    if err := proxyInstance.RunWithProgress(ctx, progressChan); err != nil {
        log.Printf("Error: %v", err)
    }
}()

for update := range progressChan {
    fmt.Printf("[%s] %s - %d%%\n", 
        update.Stage, update.Message, update.Progress)
}
```

### Custom Validation

```go
// Create custom validator
validator := proxy.NewValidator(config)

// Validate individual objects
result := validator.ValidateObject(obj)

// Validate multiple objects
results := validator.ValidateObjects(objects)
```

### Custom Reporting

```go
// Create custom reporter
reporter := proxy.NewReporter(config)

// Generate report
report, err := reporter.GenerateReport(results)
if err != nil {
    log.Fatal(err)
}

// Access statistics
fmt.Printf("Total objects: %d\n", report.Statistics.TotalObjects)
fmt.Printf("Valid objects: %d\n", report.Statistics.ValidObjects)
```

## Error Handling

The package provides comprehensive error handling:

- **Connection Errors**: Network connectivity issues
- **Validation Errors**: Object validation failures
- **Configuration Errors**: Invalid configuration settings
- **Report Generation Errors**: File system or formatting issues

All errors include context and can be wrapped for additional information.

## Testing

Run the test suite:

```bash
go test ./proxy/...
```

Run with coverage:

```bash
go test -cover ./proxy/...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This package is licensed under the same license as the parent project.

## Support

For issues and questions:

1. Check the documentation
2. Search existing issues
3. Create a new issue with detailed information
4. Include configuration and error logs
