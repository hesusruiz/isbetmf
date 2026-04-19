package service

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"log/slog"

	"github.com/hesusruiz/isbetmf/config"
	"github.com/hesusruiz/isbetmf/internal/errl"
	pdp "github.com/hesusruiz/isbetmf/pdp"
	"github.com/hesusruiz/isbetmf/tmfserver/notifications"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
	"github.com/hesusruiz/isbetmf/types"
)

// The DOME implementation has some non-conformances with the TMF specifications, and this is to bypass them.
var DOMEHacks = true

// Request represents a generic HTTP request. Handlers must convert to this representation.
// In this way, we support easily any HTTP framework (currently Fiber), but also other
// future channels like JSON-RPC or even non-HTTP channels like GRPC.
type Request struct {
	Method        string
	Action        HttpAction
	APIfamily     string
	APIVersion    string
	ResourceName  string
	ID            string
	QueryParams   url.Values
	Body          []byte
	AuthUser      types.AuthUser
	HealthRequest bool
}

func (r *Request) ToMap() map[string]any {
	return map[string]any{
		"method":   r.Method,
		"action":   r.Action,
		"api":      r.APIfamily,
		"version":  r.APIVersion,
		"resource": r.ResourceName,
		"id":       r.ID,
	}
}

type HttpAction string

const (
	READ   HttpAction = "READ"
	CREATE HttpAction = "CREATE"
	UPDATE HttpAction = "UPDATE"
	DELETE HttpAction = "DELETE"
	LIST   HttpAction = "LIST"
)

// These are more friendly names for the writers of policy rules and can be used interchangeably
var HttpActions = map[string]HttpAction{
	"GET":    READ,
	"POST":   CREATE,
	"PATCH":  UPDATE,
	"DELETE": DELETE,
	"LIST":   LIST,
}

// Response represents a generic HTTP response.
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       any
}

type ServerOperatorInfo struct {
	OrganizationIdentifier string
	Did                    string
	Name                   string
	Country                string
	EmailAddress           string
}

// Service is the service for the API.
type Service struct {

	// The environment where we are running
	environment config.Environment

	// The admin token used to authenticate the superadmin. Handle as a secret.
	adminToken string

	// Pluggable storage backend
	storage TMFStorage

	// The rules engine implemented using Starlark
	ruleEngine *pdp.PDP

	// The url of theVerifier server which signs the Access Tokens,
	// and the PDP retrieves the JWKS from it to verify the signatures.
	verifierServer string

	// The OpenID configuration to use the Verifier Server
	oid *OpenIDConfig

	// Notifications manager
	notif *notifications.Manager

	// TMF Client for proxying requests
	tmfClient *TMFClient

	// Fressness for local objects when proxy enabled
	fressness time.Duration

	// Flag to enable/disable proxy functionality.
	// When not enabled, the service is a standard TMF Server, local only.
	// When enabled, the service is a proxy to a remote TMF Server.
	proxyEnabled bool

	// The paging service to help process remote TMForum objects when retrieving large quantities.
	paging *ClientWithPaging

	// Information about us (the server operator)
	ServerOperatorOrganizationIdentifier string
	ServerOperatorDid                    string
	ServerOperatorName                   string
	ServerOperatorCountry                string
	ServerEmailAddress                   string

	AdditionalTrustedparties []ServerOperatorInfo

	// The domain of the remote TMForum API server when we act as proxy
	RemoteTMFServer string

	// The power required by a caller to be considered LEAR
	LEARPower types.OnePower

	// The powers required by a caller to be able to create, update and delete a product
	ProductCreatePower types.OnePower
	ProductUpdatePower types.OnePower
	ProductDeletePower types.OnePower

	// The features of the environment
	Features config.Features
}

// NewTMFService creates a new service.
func NewTMFService(cnf *config.Config, storage TMFStorage, ruleEngine *pdp.PDP) (*Service, error) {
	svc := &Service{}

	svc.environment = cnf.Environment
	svc.adminToken = cnf.AdminToken
	svc.storage = storage
	svc.ruleEngine = ruleEngine
	svc.verifierServer = cnf.VerifierServer
	svc.Features = cnf.Features

	// Information about us (the server operator)
	svc.ServerOperatorOrganizationIdentifier = cnf.ServerOperatorOrganizationIdentifier
	svc.ServerOperatorDid = cnf.ServerOperatorDid
	svc.ServerOperatorName = cnf.ServerOperatorName
	svc.ServerOperatorCountry = cnf.ServerOperatorCountry
	svc.ServerEmailAddress = cnf.ServerEmailAddress
	svc.LEARPower = cnf.LEARPower
	svc.ProductCreatePower = cnf.ProductCreatePower
	svc.ProductUpdatePower = cnf.ProductUpdatePower
	svc.ProductDeletePower = cnf.ProductDeletePower

	// The remote TMF server when we act as a proxy to it
	if cnf.ProxyEnabled {
		cnf.RemoteTMFServer = strings.TrimRight(cnf.RemoteTMFServer, "/")
		if cnf.RemoteTMFServer == "" {
			return nil, errl.Errorf("remote TMF server not set")
		}
		svc.RemoteTMFServer = cnf.RemoteTMFServer
		svc.proxyEnabled = cnf.ProxyEnabled

		tmfClientConfig := &TMFClientConfig{
			BaseURL: svc.RemoteTMFServer,
			Timeout: 120,
		}

		svc.tmfClient = NewClient(tmfClientConfig)

		svc.fressness = cnf.ClonePeriod
		if svc.fressness == 0 {
			svc.fressness = config.DefaultClonePeriod
		}
	}

	// Create the server operator identity, in case it is not yet in the database
	org := &repository.Organization{
		CommonName:             svc.ServerOperatorName,
		Country:                svc.ServerOperatorCountry,
		EmailAddress:           svc.ServerEmailAddress,
		Organization:           svc.ServerOperatorName,
		OrganizationIdentifier: svc.ServerOperatorOrganizationIdentifier,
	}
	obj, _ := repository.TMFRecordFromOrganizationAndToken(org, nil)

	if err := svc.UpsertObject(obj); err != nil {
		if errors.Is(err, &ErrObjectExists{}) {
			slog.Debug("server operator organization already exists", "organizationIdentifier", svc.ServerOperatorOrganizationIdentifier)
		} else {
			err = errl.Errorf("error creating server operator organization: %w", err)
			panic(err)
		}
	} else {
		slog.Info("server operator organization created", "obj_id", obj.ID)
	}

	// Retrieve the OpenId configuration of the Verifier server
	oid, err := NewOpenIDConfig(svc.verifierServer)
	if err != nil {
		return nil, errl.Errorf("failed to retrieve OpenID configuration: %w", err)
	}
	svc.oid = oid

	// Create the paging service
	pagingConfig := DefaultPagingConfig()
	pagingConfig.PageSize = 10
	svc.paging = NewClientWithPaging(pagingConfig)

	// Initialize notifications with in-memory store and HTTP delivery
	store := notifications.NewMemoryStore()
	deliver := notifications.NewHTTPDelivery(5 * time.Second)
	svc.notif = notifications.NewManager(store, deliver)

	return svc, nil
}

func (s *Service) IsDOME() bool {
	return s.environment == config.DOME_PRO || s.environment == config.DOME_PRE || s.environment == config.DOME_DEV || s.environment == config.LOCAL
}

func (s *Service) IsISBE() bool {
	return s.environment == config.ISBE_PRE || s.environment == config.ISBE_DEV
}

// ToKebabCase converts a camelCase string to kebab-case.
// For example: "productOffering" becomes "product-offering".
func ToKebabCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(s, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}

// LifecycleStatusMandatory defines which TMF resources require the lifecycleStatus field as mandatory.
// The map key is the resource name, and the value is the default lifecycleStatus value to set if missing.
// The keys are in lowercase to facilitate case-independent lookup.
var LifecycleStatusMandatory = map[string]string{
	// TMF620
	"catalog":              "In Study",
	"category":             "In Study",
	"productOffering":      "In Study",
	"productOfferingPrice": "In Study",
	"productSpecification": "In Study",

	// TMF633
	"serviceCandidate":     "In Study",
	"serviceCatalog":       "In Study",
	"serviceCategory":      "In Study",
	"serviceSpecification": "In Study",

	// TMF634
	"LogicalResourceSpecification":  "In Study",
	"PhysicalResourceSpecification": "In Study",
	"ResourceCandidate":             "In Study",
	"ResourceCatalog":               "In Study",
	"ResourceCategory":              "In Study",
	"ResourceFunctionSpecification": "In Study",
	"ResourceSpecification":         "In Study",

	// TMF635
	"UsageSpecification": "In Study",

	// TMF651
	"AgreementSpecification": "In Study",
}
