package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/hesusruiz/isbetmf/internal/errl"
	pdp "github.com/hesusruiz/isbetmf/pdp"
	"github.com/hesusruiz/isbetmf/tmfserver/repository"
)

func takeDecision(
	ruleEngine *pdp.PDP,
	r *Request,
	tokenClaims map[string]any,
	tmfObject *repository.TMFObject,
) (err error) {

	// Some rules are hardcoded because they are always enforced
	// The rest is delegated to the policy engine

	// First we need to get a map representing the object
	objMap := make(map[string]any)
	err = json.Unmarshal(tmfObject.Content, &objMap)
	if err != nil {
		return errl.Error(err)
	}

	// The object must have both the seller and sellerOperator identities
	sellerDid, sellerOperatorDid, err := getSellerAndBuyerInfo(objMap)
	if err != nil || sellerDid == "" || sellerOperatorDid == "" {
		err = errl.Errorf("failed to get seller and buyer info: %w", err)
		return err
	}

	userDid := r.AuthUser.OrganizationIdentifier
	if !strings.HasPrefix(userDid, "did:elsi:") {
		userDid = "did:elsi:" + userDid
	}

	r.AuthUser.isOwner = (userDid == sellerDid) || (userDid == sellerOperatorDid)

	// Assemble all data in a single "input" argument, to the style of OPA.
	// We mutate the predeclared identifier, so the policy can access the data for this request.
	// We can also service possible callbacks from the rules engine.

	userArgument := pdp.StarTMFMap(r.AuthUser.ToMap())
	tmfObjectArgument := pdp.StarTMFMap(tmfObject.ToMap())
	requestArgument := pdp.StarTMFMap(r.ToMap())
	tokenArgument := pdp.StarTMFMap(tokenClaims)

	input := map[string]any{
		"request": requestArgument,
		"token":   tokenArgument,
		"tmf":     tmfObjectArgument,
		"user":    userArgument,
	}

	if slog.Default().Enabled(context.Background(), slog.LevelDebug) {
		b, err := json.MarshalIndent(input, "", "  ")
		if err == nil {
			fmt.Println("PDP input:", string(b))
		}
	}

	decision := true
	if ruleEngine != nil {
		decision, err = ruleEngine.TakeAuthnDecision(pdp.Authorize, input)

		// An error is considered a rejection, continue with the next candidate object
		if err != nil {
			return errl.Errorf("rules engine rejected request due to an error: %w", err)
		}
	}

	// The rules engine rejected the request, continue with the next candidate object
	if !decision {
		return errl.Errorf("PDP: request rejected due to policy")
	}

	// The rules engine accepted the request, add the object to the final list
	slog.Info("PDP: request authorised")
	return nil
}
