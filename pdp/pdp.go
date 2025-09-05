// Copyright 2023-2025 Jesus Ruiz. All rights reserved.
// Use of this source code is governed by an Apache 2.0
// license that can be found in the LICENSE file.

package pdp

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hesusruiz/isbetmf/internal/errl"
	"github.com/hesusruiz/isbetmf/internal/filecache"

	st "go.starlark.net/starlark"
)

type Config struct {

	// PolicyFileName is the name of the file where the policies are stored.
	// It can specify a local file or a remote URL.
	PolicyFileName string

	// The http Client to retrieve the policies from a remote server.
	// If nil we use our own http.Client with a timeout of 10 seconds and no redirects.
	httpClient *http.Client

	// Debug mode, more logs and less caching
	Debug bool
}

// Validate checks if the Config is valid
func (c *Config) Validate() error {
	if c == nil {
		return errl.Errorf("config cannot be nil")
	}
	if c.PolicyFileName == "" {
		return errl.Errorf("PolicyFileName is required")
	}
	return nil
}

// PDP implements a simple Policy Decision Point in Starlark, for use in front of TMForum APIs.
//
// There can be several instances simultaneously, and each instance is safe for concurrent
// use by different goroutines.
type PDP struct {

	// // The configuration of the PDP, which includes the file with the policies and other parameters.
	// config *Config

	// The name of the file where the policy rules reside.
	scriptname string

	debug bool

	// The file cache to read the policy and other files. Modifications to the original file
	// are picked up automatically according to a freshness policy.
	// fileCache    sync.Map
	fileCache *filecache.SimpleFileCache

	// The pool of instances of the policy execution engines, to minimize startup
	// and teardown overheads.
	// Every goroutine uses its own instance from the pool, so they are goroutine safe.
	// If the file with the policies change, the associated Starlark thread is updated,
	// Goroutines using old versions will run until completion, and new ones will pick
	// the new version of the policies
	threadPool sync.Pool

	// The http Client to retrieve the policies from a remote server if configured to do so.
	httpClient *http.Client
}

// NewPDP creates a new PDP instance.
func NewPDP(
	config *Config,
) (*PDP, error) {

	if err := config.Validate(); err != nil {
		return nil, errl.Errorf("invalid config: %w", err)
	}

	m := &PDP{}
	m.scriptname = config.PolicyFileName

	// Create the file cache and initialize it with the policy file.
	m.fileCache = filecache.NewSimpleFileCache(nil)
	m.fileCache.Get(config.PolicyFileName)

	// Create the pool of parsed and compiled Starlark policy rules.
	m.threadPool = sync.Pool{
		New: func() any {
			return m.bufferedParseAndCompileFile(m.scriptname)
		},
	}

	m.debug = config.Debug

	if config.httpClient != nil {
		// Use the supplied http.Client if provided.
		m.httpClient = config.httpClient
	} else {
		// We use an http.Client with a timeout of 10 seconds and no redirects.
		m.httpClient = &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return errors.New("no redirects allowed")
			},
		}
	}

	return m, nil
}

// bufferedParseAndCompileFile reads a file with Starlark code and compiles it
func (m *PDP) bufferedParseAndCompileFile(scriptname string) *threadEntry {
	te := m.createThreadEntry(scriptname)
	if te == nil {
		return nil
	}

	entry, err := m.fileCache.Get(scriptname)
	if err != nil {
		return nil
	}

	te.scriptHash = entry.FileHash
	src := entry.Content

	if err := m.compileStarlarkScript(te, string(src)); err != nil {
		return nil
	}

	if err := m.validateCompiledScript(te); err != nil {
		return nil
	}

	return te
}

// reset checks if the thread entry needs to be recompiled
func (m *PDP) reset(te *threadEntry) error {
	entry, err := m.fileCache.Get(te.scriptname)
	if err != nil {
		return errl.Errorf("error reading script file %s: %w", te.scriptname, err)
	}

	// If hashes are the same, we do not need to recompile the file.
	if entry.FileHash == te.scriptHash {
		return nil
	}

	// The file has changed, so we recompile it.
	src := entry.Content

	if err := m.compileStarlarkScript(te, string(src)); err != nil {
		return errl.Errorf("error compiling Starlark program: %w", err)
	}

	if err := m.validateCompiledScript(te); err != nil {
		return errl.Errorf("error getting authorize function: %w", err)
	}

	return nil
}

// Authorize evaluates authorization policies against the provided input data.
// It returns true if the request is authorized, false otherwise.
func (m *PDP) Authorize(input StarTMFMap) (bool, error) {
	return m.evaluateDecision(Authorize, input)
}

// evaluateDecision is the internal function that handles both authentication and authorization decisions
func (m *PDP) evaluateDecision(decision Decision, input StarTMFMap) (bool, error) {
	if !decision.IsValid() {
		return false, errl.Errorf("invalid decision type: %v", decision)
	}

	if input == nil {
		return false, errl.Errorf("input cannot be nil")
	}

	// Get a Starlark Thread from the pool to evaluate the policies.
	ent := m.threadPool.Get()
	if ent == nil {
		return false, errl.Errorf("getting a thread entry from pool")
	}
	defer m.threadPool.Put(ent)

	te := ent.(*threadEntry)
	if te == nil {
		return false, errl.Errorf("invalid entry type in the pool")
	}

	// Check if the thread is still valid. If not, we need to recompile the file.
	err := m.reset(te)
	if err != nil {
		return false, err
	}

	// We mutate the predeclared identifier, so the policy can access the data for this request.
	// We can also service possible callbacks from the rules engine.
	te.predeclared["input"] = input

	// Build the arguments to the StarLark function, which is empty.
	var args st.Tuple

	// Call the corresponding function in the Starlark Thread
	var result st.Value
	if decision == Authenticate {
		// For now, we only support authorization
		return false, errl.Errorf("authentication not yet implemented")
	} else {
		// Call the 'authorize' function
		result, err = st.Call(te.thread, te.authorizeFunction, args, nil)
	}

	if err != nil {
		fmt.Printf("rules ERROR: %s\n", err.(*st.EvalError).Backtrace())
		return false, errl.Errorf("error calling function: %w", err)
	}

	// Check that the value returned is of the correct type (boolean)
	resultType := result.Type()
	if resultType != "bool" {
		err := errl.Errorf("function returned wrong type: %v", resultType)
		return false, err
	}

	// Return the value as a Go boolean
	return bool(result.(st.Bool).Truth()), nil
}

func (m *PDP) GetFile(filename string) (*filecache.FileEntry, error) {

	entry, err := m.fileCache.MustExist(filename)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (m *PDP) PutFile(filename string, content []byte) error {
	return m.fileCache.Set(filename, content, 0)
}
