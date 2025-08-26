package service

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hesusruiz/isbetmf/tmfserver/notifications"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	// In-memory SQLite DB
	sqldb, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sqlite open: %v", err)
	}
	db := sqlx.NewDb(sqldb, "sqlite3")

	_, err = db.Exec(repository.CreateTMFTableSQL)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Create service struct directly (no external verifier)
	s := &Service{db: db}
	// Wire notifications manager to a fake delivery by default
	s.notif = notifications.NewManager(notifications.NewMemoryStore(), &fakeDelivery{})
	return s
}

// newReq creates a fresh Request without AccessToken to ensure AllowFakeClaims path is used
func newReq(method, action, api, resource, id string, body []byte, qp url.Values) *Request {
	return &Request{
		Method:       method,
		Action:       action,
		APIfamily:    api,
		ResourceName: resource,
		ID:           id,
		Body:         body,
		QueryParams:  qp,
	}
}

// fakeDelivery records delivered payloads for assertions
type fakeDelivery struct{ deliveries []any }

func (f *fakeDelivery) Deliver(_ *notifications.Subscription, payload any) error {
	f.deliveries = append(f.deliveries, payload)
	return nil
}

func TestCreateAndDeleteHubSubscription(t *testing.T) {
	s := newTestService(t)

	body := map[string]any{
		"callback":   "http://localhost:9991/listener/test",
		"eventTypes": []string{"ProductOfferingCreateEvent"},
		"headers":    map[string]any{"x-auth-token": "abc123"},
	}
	b, _ := json.Marshal(body)
	req := newReq("POST", "CREATE", "TMF620", "", "", b, nil)

	resp := s.CreateHubSubscription(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	respMap, _ := resp.Body.(map[string]any)
	id, _ := respMap["id"].(string)
	if id == "" {
		t.Fatalf("expected id in response")
	}

	// Delete
	delReq := newReq("DELETE", "DELETE", "TMF620", "", id, nil, nil)
	delResp := s.DeleteHubSubscription(delReq)
	if delResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", delResp.StatusCode)
	}
}

func TestCreateGenericObjectPublishesEvent(t *testing.T) {
	s := newTestService(t)

	// Replace notifications manager with one that uses fake delivery
	memStore := notifications.NewMemoryStore()
	fdel := &fakeDelivery{}
	s.notif = notifications.NewManager(memStore, fdel)

	// Create a subscription to receive create events
	sub := &notifications.Subscription{
		ID:         "sub1",
		APIFamily:  "TMF620",
		Callback:   "http://localhost:9991/listener/ProductOfferingCreateEvent",
		EventTypes: []string{"ProductOfferingCreateEvent"},
		Headers:    map[string]string{"x-auth-token": "abc123"},
	}
	if _, err := s.notif.CreateSubscription("TMF620", sub); err != nil {
		t.Fatalf("create sub: %v", err)
	}

	resourceName := "productOffering"
	obj := map[string]any{
		"@type":   resourceName,
		"version": "1.0",
	}
	b, _ := json.Marshal(obj)
	req := newReq("POST", "CREATE", "TMF620", resourceName, "", b, nil)

	resp := s.CreateGenericObject(req)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// Wait briefly for goroutine delivery
	time.Sleep(200 * time.Millisecond)

	if len(fdel.deliveries) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(fdel.deliveries))
	}
	// Basic payload shape assertions
	payload, ok := fdel.deliveries[0].(map[string]any)
	if !ok {
		t.Fatalf("payload not a map")
	}
	if payload["eventType"] != "ProductOfferingCreateEvent" {
		t.Fatalf("unexpected eventType: %v", payload["eventType"])
	}
}

func TestCRUDAndListGenericObject(t *testing.T) {
	s := newTestService(t)

	// Create
	resourceName := "productOffering"
	createObj := map[string]any{
		"@type": resourceName,
	}
	bCreate, _ := json.Marshal(createObj)
	cReq := newReq("POST", "CREATE", "TMF620", resourceName, "", bCreate, nil)
	cResp := s.CreateGenericObject(cReq)
	if cResp.StatusCode != http.StatusCreated {
		t.Fatalf("create expected 201, got %d", cResp.StatusCode)
	}
	bodyMap := cResp.Body.(map[string]any)
	id, _ := bodyMap["id"].(string)
	if id == "" {
		t.Fatalf("no id returned")
	}

	// Get
	gReq := newReq("GET", "READ", "TMF620", resourceName, id, nil, nil)
	gResp := s.GetGenericObject(gReq)
	if gResp.StatusCode != http.StatusOK {
		t.Fatalf("get expected 200, got %d", gResp.StatusCode)
	}

	// Update (must include version greater than existing)
	upd := map[string]any{
		"@type":   resourceName,
		"id":      id,
		"version": "1.1",
	}
	bUpd, _ := json.Marshal(upd)
	uReq := newReq("PATCH", "UPDATE", "TMF620", resourceName, id, bUpd, nil)
	uResp := s.UpdateGenericObject(uReq)
	if uResp.StatusCode != http.StatusOK {
		t.Fatalf("update expected 200, got %d", uResp.StatusCode)
	}
	updated := uResp.Body.(map[string]any)
	if updated["version"].(string) != "1.1" {
		t.Fatalf("expected version 1.1, got %v", updated["version"])
	}

	// List (all)
	lReq := newReq("GET", "LIST", "TMF620", resourceName, "", nil, url.Values{})
	lResp := s.ListGenericObjects(lReq)
	if lResp.StatusCode != http.StatusOK {
		t.Fatalf("list expected 200, got %d", lResp.StatusCode)
	}
	if lResp.Headers["X-Total-Count"] == "" {
		t.Fatalf("missing X-Total-Count header")
	}

	// List with fields=none (should reduce fields per item)
	lReqQP := newReq("GET", "LIST", "TMF620", resourceName, "", nil, url.Values{"fields": []string{"none"}})
	lResp2 := s.ListGenericObjects(lReqQP)
	if lResp2.StatusCode != http.StatusOK {
		t.Fatalf("list expected 200, got %d", lResp2.StatusCode)
	}
	items, ok := lResp2.Body.([]map[string]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected list of items")
	}
	// Expect minimal keys present
	item := items[0]
	if item["id"] == nil || item["href"] == nil || item["version"] == nil || item["lastUpdate"] == nil || item["@type"] == nil {
		t.Fatalf("fields=none did not include minimal fields")
	}

	// Delete
	dReq := newReq("DELETE", "DELETE", "TMF620", resourceName, id, nil, nil)
	dResp := s.DeleteGenericObject(dReq)
	if dResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete expected 204, got %d", dResp.StatusCode)
	}

	// Get after delete -> 404
	gReq = newReq("GET", "READ", "TMF620", resourceName, id, nil, nil)
	gResp2 := s.GetGenericObject(gReq)
	if gResp2.StatusCode != http.StatusNotFound {
		t.Fatalf("get after delete expected 404, got %d", gResp2.StatusCode)
	}
}
