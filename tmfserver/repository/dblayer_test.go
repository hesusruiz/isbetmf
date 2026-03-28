package repository

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func argsToStrings(args []any) []string {
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = fmt.Sprint(a)
	}
	return out
}

func containsAllInOrder(got []string, want []string) bool {
	if len(want) == 0 {
		return true
	}
	j := 0
	for _, g := range got {
		if g == want[j] {
			j++
			if j == len(want) {
				return true
			}
		}
	}
	return false
}

func TestBuildSelectFromParms_NoParams(t *testing.T) {
	sql, args, _, _, _ := BuildSelectFromParms("", url.Values{})
	if !strings.Contains(sql, "FROM tmf_object") {
		t.Fatalf("expected SQL to contain FROM tmf_object, got: %s", sql)
	}
	if !strings.Contains(sql, "GROUP BY id") {
		t.Fatalf("expected SQL to contain GROUP BY id, got: %s", sql)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args for empty params, got: %v", args)
	}
}

func TestBuildSelectFromParms_ResourceFilter(t *testing.T) {
	sql, args, _, _, _ := BuildSelectFromParms("Product", url.Values{})
	if !strings.Contains(strings.ToLower(sql), "where") {
		t.Fatalf("expected SQL to contain WHERE when resource provided, got: %s", sql)
	}
	argStrs := argsToStrings(args)
	if len(argStrs) != 1 || argStrs[0] != "Product" {
		t.Fatalf("expected single arg 'Product', got: %v", argStrs)
	}
}

func TestBuildSelectFromParms_LimitOffset(t *testing.T) {
	v := url.Values{}
	v.Set("limit", "5")
	v.Set("offset", "10")
	sql, args, _, _, _ := BuildSelectFromParms("", v)

	if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
		t.Fatalf("expected SQL to contain LIMIT, got: %s", sql)
	}
	if !strings.Contains(strings.ToUpper(sql), "OFFSET") {
		t.Fatalf("expected SQL to contain OFFSET, got: %s", sql)
	}
	// args may include limit/offset or they may be inlined depending on builder; ensure no panic and SQL contains clauses
	_ = args
}

func TestBuildSelectFromParms_SellerMultipleValues(t *testing.T) {
	v := url.Values{}
	// multiple comma-separated values and multiple instances
	v.Add("seller", "a,b")
	v.Add("seller", "c")
	sql, args, _, _, _ := BuildSelectFromParms("", v)

	if !strings.Contains(sql, "seller") {
		t.Fatalf("expected SQL to reference seller, got: %s", sql)
	}
	argStrs := argsToStrings(args)
	want := []string{"a", "b", "c"}
	if !containsAllInOrder(argStrs, want) {
		t.Fatalf("expected args to contain %v in order, got: %v", want, argStrs)
	}
}

func TestBuildSelectFromParms_JSONFieldMultiValues(t *testing.T) {
	v := url.Values{}
	v.Add("status", "Active,Launched")
	sql, args, _, _, _ := BuildSelectFromParms("", v)

	if !strings.Contains(sql, "content->>'$.status'") {
		t.Fatalf("expected SQL to reference JSON path for status, got: %s", sql)
	}
	argStrs := argsToStrings(args)
	want := []string{"Active", "Launched"}
	if !containsAllInOrder(argStrs, want) {
		t.Fatalf("expected args to contain %v in order, got: %v", want, argStrs)
	}
}

func TestBuildSelectFromParms_TopLevelField(t *testing.T) {
	v := url.Values{}
	v.Set("lifecycleStatus", "Launched")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "content->>'$.lifecycleStatus'") {
		t.Fatalf("expected SQL to reference JSON path for lifecycleStatus, got: %s", sql)
	}
	if len(args) < 2 || args[1] != "Launched" {
		t.Fatalf("expected args to contain 'Launched', got: %v", args)
	}
}

func TestBuildSelectFromParms_MultiValueTopLevelField(t *testing.T) {
	v := url.Values{}
	v.Set("lifecycleStatus", "Launched,Active")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "content->>'$.lifecycleStatus'") {
		t.Fatalf("expected SQL to reference JSON path for lifecycleStatus, got: %s", sql)
	}
	want := []string{"Launched", "Active"}
	argStrs := argsToStrings(args)
	if !containsAllInOrder(argStrs, want) {
		t.Fatalf("expected args to contain %v in order, got: %v", want, argStrs)
	}
}

func TestBuildSelectFromParms_NestedField(t *testing.T) {
	v := url.Values{}
	// Simulate filtering by productSpecification.id
	v.Set("productSpecification.id", "urn:ngsi-ld:product-specification:19f7f34a-1777-4553-b47b-6ad466d8a0ea")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "json_each(tmf_object.content, '$.productSpecification')") {
		t.Fatalf("expected SQL to reference JSON path for productSpecification.id, got: %s", sql)
	}
	if len(args) < 2 || args[1] != "urn:ngsi-ld:product-specification:19f7f34a-1777-4553-b47b-6ad466d8a0ea" {
		t.Fatalf("expected args to contain productSpecification.id value, got: %v", args)
	}
}

func TestBuildSelectFromParms_ArrayOfObjectsField(t *testing.T) {
	v := url.Values{}
	// Simulate filtering by category.id (array of objects)
	v.Set("category.id", "urn:ngsi-ld:category:31a1d8a8-56e8-49c3-aabb-6b0306bc0316")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "json_each(tmf_object.content, '$.category')") {
		t.Fatalf("expected SQL to reference JSON path for category.id, got: %s", sql)
	}
	if len(args) < 2 || args[1] != "urn:ngsi-ld:category:31a1d8a8-56e8-49c3-aabb-6b0306bc0316" {
		t.Fatalf("expected args to contain category.id value, got: %v", args)
	}
}

func TestBuildSelectFromParms_ArrayOfObjectsMultiValue(t *testing.T) {
	v := url.Values{}
	// Simulate filtering by relatedParty.role with multiple values
	v.Set("relatedParty.role", "Seller,SellerOperator")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "content->>'$.relatedParty.role'") {
		t.Fatalf("expected SQL to reference JSON path for relatedParty.role, got: %s", sql)
	}
	want := []string{"Seller", "SellerOperator"}
	argStrs := argsToStrings(args)
	if !containsAllInOrder(argStrs, want) {
		t.Fatalf("expected args to contain %v in order, got: %v", want, argStrs)
	}
}

func TestBuildSelectFromParms_MultipleFilters(t *testing.T) {
	v := url.Values{}
	v.Set("lifecycleStatus", "Launched")
	v.Set("name", "Product Offer Example")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "content->>'$.lifecycleStatus'") || !strings.Contains(sql, "content->>'$.name'") {
		t.Fatalf("expected SQL to reference both lifecycleStatus and name, got: %s", sql)
	}
	want := []string{"Launched", "Product Offer Example"}
	argStrs := argsToStrings(args)
	if !containsAllInOrder(argStrs, want) {
		t.Fatalf("expected args to contain %v in order, got: %v", want, argStrs)
	}
}

func TestBuildSelectFromParms_LimitOffsetAndType(t *testing.T) {
	v := url.Values{}
	v.Set("limit", "2")
	v.Set("offset", "1")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(strings.ToUpper(sql), "LIMIT") {
		t.Fatalf("SQL contains LIMIT, got: %s", sql)
	}
	if !strings.Contains(strings.ToUpper(sql), "OFFSET") {
		t.Fatalf("SQL contains OFFSET, got: %s", sql)
	}
	if !strings.Contains(sql, "WHERE") {
		t.Fatalf("expected SQL to contain WHERE for type, got: %s", sql)
	}
	if len(args) != 1 || args[0] != "ProductOffering" {
		t.Fatalf("expected args to contain only ProductOffering, got: %v", args)
	}
}

func TestBuildSelectFromParms_SellerShortcut(t *testing.T) {
	v := url.Values{}
	v.Set("seller", "did:elsi:VATES-B60645900,did:elsi:VATES-11111111K")
	sql, args, _, _, _ := BuildSelectFromParms("ProductOffering", v)
	if !strings.Contains(sql, "seller") {
		t.Fatalf("expected SQL to reference seller, got: %s", sql)
	}
	want := []string{"did:elsi:VATES-B60645900", "did:elsi:VATES-11111111K"}
	argStrs := argsToStrings(args)
	if !containsAllInOrder(argStrs, want) {
		t.Fatalf("expected args to contain %v in order, got: %v", want, argStrs)
	}
}
