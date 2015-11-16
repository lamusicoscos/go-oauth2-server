package oauth

import (
	"log"
	"net/http"

	"github.com/RichardKnop/go-oauth2-server/json"
	"github.com/RichardKnop/go-oauth2-server/util"
)

func (s *Service) refreshTokenGrant(w http.ResponseWriter, r *http.Request, client *Client) {
	// Double check the grant type
	log.Print(r.Form)
	if r.Form.Get("grant_type") != "refresh_token" {
		json.Error(w, "Invalid grant type", http.StatusBadRequest)
		return
	}

	// Validate the refresh token
	theRefreshToken, err := s.ValidateRefreshToken(
		r.Form.Get("refresh_token"),
		client,
	)
	if err != nil {
		json.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the scope string
	scope, err := s.getScope(r.Form.Get("scope"))
	if err != nil {
		json.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Requested scope CANNOT include any scope not originally granted
	if !util.SpaceDelimitedStringNotGreater(scope, theRefreshToken.Scope) {
		json.Error(w, "Requested scope cannot be greater", http.StatusBadRequest)
		return
	}

	// Create a new access token
	accessToken, err := s.GrantAccessToken(
		theRefreshToken.Client,
		theRefreshToken.User,
		scope,
	)
	if err != nil {
		json.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Create or retrieve a refresh token
	refreshToken, err := s.GetOrCreateRefreshToken(
		theRefreshToken.Client,
		theRefreshToken.User,
		scope,
	)
	if err != nil {
		json.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Write the access token to a JSON response
	writeJSON(w, s.cnf.Oauth.AccessTokenLifetime, accessToken, refreshToken)
}
