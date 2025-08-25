// Copyright 2023 Jesus Ruiz. All rights reserved.
// Use of this source code is governed by an Apache 2.0
// license that can be found in the LICENSE file.

package service

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/goccy/go-json"
	"github.com/hesusruiz/isbetmf/internal/errl"
	"gitlab.com/greyxor/slogor"
)

type OpenIDConfig struct {
	Issuer                                    string   `json:"issuer,omitempty"`
	AuthorizationEndpoint                     string   `json:"authorization_endpoint,omitempty"`
	DeviceAuthorizationEndpoint               string   `json:"device_authorization_endpoint,omitempty"`
	TokenEndpoint                             string   `json:"token_endpoint,omitempty"`
	TokenEndpointAuthMethodsSupported         []string `json:"token_endpoint_auth_methods_supported,omitempty"`
	JwksUri                                   string   `json:"jwks_uri,omitempty"`
	UserinfoEndpoint                          string   `json:"userinfo_endpoint,omitempty"`
	EndSessionEndpoint                        string   `json:"end_session_endpoint,omitempty"`
	ResponseTypesSupported                    []string `json:"response_types_supported,omitempty"`
	GrantTypesSupported                       []string `json:"grant_types_supported,omitempty"`
	RevocationEndpoint                        string   `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported    []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	IntrospectionEndpoint                     string   `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	CodeChallengeMethodsSupported             []string `json:"code_challenge_methods_supported,omitempty"`
	TlsClientCertificateBoundAccessTokens     bool     `json:"tls_client_certificate_bound_access_tokens,omitempty"`
	SubjectTypesSupported                     []string `json:"subject_types_supported,omitempty"`
	IdTokenSigningAlgValuesSupported          []string `json:"id_token_signing_alg_values_supported,omitempty"`
	ScopesSupported                           []string `json:"scopes_supported,omitempty"`

	cachedJWK   *jose.JSONWebKey
	lastRefresh time.Time
	freshness   time.Duration
}

func NewOpenIDConfig(verifierServer string) (*OpenIDConfig, error) {

	verifierWellKnownURL := verifierServer + "/.well-known/openid-configuration"

	res, err := http.Get(verifierWellKnownURL)
	if err != nil {
		return nil, errl.Errorf("failed to retrieve OpenID configuration: %w", err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		return nil, errl.Errorf("response failed with status: %d", res.StatusCode)
	}
	if err != nil {
		return nil, errl.Errorf("reading response body: %w", err)
	}

	oid := &OpenIDConfig{}
	err = json.Unmarshal(body, oid)
	if err != nil {
		return nil, errl.Errorf("unmarshalling OpenID configuration: %w", err)
	}

	if oid.JwksUri == "" {
		return nil, errl.Errorf("no JwksUri")
	}

	slog.Debug("JWKS URI", "uri", oid.JwksUri)

	// Set default refresh period
	oid.freshness = time.Hour

	// Load for key, to detect possible errors
	_, err = oid.VerificationJWK()
	if err != nil {
		return nil, errl.Error(err)
	}

	return oid, nil
}

func (oid *OpenIDConfig) VerificationJWK() (*jose.JSONWebKey, error) {

	if oid.JwksUri == "" {
		return nil, fmt.Errorf("no JwksUri")
	}

	// Check if we have a valid cached key
	if oid.cachedJWK != nil && time.Since(oid.lastRefresh) < oid.freshness {
		slog.Debug("returning cached JWK")
		return oid.cachedJWK, nil
	}

	res, err := http.Get(oid.JwksUri)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode > 299 {
		err := fmt.Errorf("response failed with status: %d", res.StatusCode)
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var jwks = &jose.JSONWebKeySet{}
	err = json.Unmarshal(body, jwks)
	if err != nil {
		slog.Error("unmarshalling JWKS", slogor.Err(err))
		return nil, err
	}

	if len(jwks.Keys) == 0 {
		err := fmt.Errorf("no JWK keys returned")
		return nil, err
	}

	slog.Debug("retrieved JWK")
	oid.cachedJWK = &jwks.Keys[0]
	oid.lastRefresh = time.Now()

	return &jwks.Keys[0], nil

}
