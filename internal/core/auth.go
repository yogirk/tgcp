package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rk/tgcp/internal/utils"
	"golang.org/x/oauth2/google"
)

// AuthState holds the authentication information
type AuthState struct {
	Authenticated bool
	UserEmail     string
	ProjectID     string
	Error         error
}

// Authenticate performs the ADC check and project detection
// If projectOverride is not empty, it takes precedence.
func Authenticate(ctx context.Context, projectOverride string) AuthState {
	utils.Log("Starting authentication...")

	state := AuthState{
		Authenticated: false,
	}

	// 1. Find Credentials
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		utils.Log("Authentication failed: %v", err)
		state.Error = fmt.Errorf("could not find default credentials: %w", err)
		return state
	}

	state.Authenticated = true

	// 2. Determine Project ID
	// Priority: Override Flag > Credentials JSON > Quota Project
	if projectOverride != "" {
		state.ProjectID = projectOverride
		utils.Log("Using project override: %s", state.ProjectID)
	} else if creds.ProjectID != "" {
		state.ProjectID = creds.ProjectID
		utils.Log("Found project ID in credentials: %s", state.ProjectID)
	} else {
		// Try to read quota project from internal fields if possible, or fallback manually
		// Since we can't easily access internal fields, we might check gcloud config as fallback
		// For now, let's leave it empty and let the UI handle "No project selected"
		utils.Log("No project ID found in credentials")
	}

	// 3. Try to determine User Email (best effort)
	// If credentials are from a JSON file (Service Account or Authorized User), we might parse it.
	state.UserEmail = "Unknown"

	// Check standard ADC location for more info
	// This is a heuristic to display the user email in the UI
	// Standard path: ~/.config/gcloud/application_default_credentials.json
	home, _ := os.UserHomeDir()
	adcPath := filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")

	if content, err := os.ReadFile(adcPath); err == nil {
		// Just a simple partial struct to grab the client_email/client_id if present
		var jsonCreds struct {
			ClientEmail string `json:"client_email"` // for service accounts
			ClientID    string `json:"client_id"`    // for user credentials, usually doesn't have email directly in simple way without token info
			Type        string `json:"type"`
			Account     string `json:"account"` // sometimes present in gcloud config
		}
		// Also check "account" field which might be there in some versions
		var raw map[string]interface{}
		if err := json.Unmarshal(content, &raw); err == nil {
			if val, ok := raw["client_email"].(string); ok {
				state.UserEmail = val
			} else if val, ok := raw["account"].(string); ok {
				state.UserEmail = val
			} else if val, ok := raw["quota_project_id"].(string); ok && state.ProjectID == "" {
				// Fallback project ID if not finding it elsewhere
				state.ProjectID = val
			}
		}

		if err := json.Unmarshal(content, &jsonCreds); err == nil {
			if jsonCreds.ClientEmail != "" {
				state.UserEmail = jsonCreds.ClientEmail
			} else if jsonCreds.Account != "" {
				state.UserEmail = jsonCreds.Account
			}
		}
	} else {
		// If GOOGLE_APPLICATION_CREDENTIALS is set, try to read that
		envCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
		if envCreds != "" {
			if content, err := os.ReadFile(envCreds); err == nil {
				var raw map[string]interface{}
				if err := json.Unmarshal(content, &raw); err == nil {
					if val, ok := raw["client_email"].(string); ok {
						state.UserEmail = val
					}
				}
			}
		}
	}

	utils.Log("Auth complete. User: %s, Project: %s", state.UserEmail, state.ProjectID)
	return state
}
