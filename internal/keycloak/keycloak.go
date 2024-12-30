package keycloak

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func CreateKeycloakTenant(tenantName, adminUsername, adminPassword string) error {
	// Get Keycloak Admin URL and credentials from environment variables
	keycloakURL := os.Getenv("KEYCLOAK_URL")
	clientID := os.Getenv("KEYCLOAK_CLIENT_ID")
	clientSecret := os.Getenv("KEYCLOAK_CLIENT_SECRET")

	// Validate required environment variables
	if keycloakURL == "" {
		return fmt.Errorf("KEYCLOAK_URL environment variable is not set")
	}
	if clientID == "" {
		return fmt.Errorf("KEYCLOAK_CLIENT_ID environment variable is not set")
	}
	if clientSecret == "" {
		return fmt.Errorf("KEYCLOAK_CLIENT_SECRET environment variable is not set")
	}

	// Authenticate with Keycloak to get an access token
	token, err := getKeycloakAccessToken(keycloakURL, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("failed to authenticate with Keycloak: %v", err)
	}

	// Create a new realm
	realm := map[string]interface{}{
		"id":      tenantName,
		"realm":   tenantName,
		"enabled": true,
		"users": []map[string]interface{}{
			{
				"username":    adminUsername,
				"enabled":     true,
				"credentials": []map[string]string{{"type": "password", "value": adminPassword}},
			},
		},
	}

	body, err := json.Marshal(realm)
	if err != nil {
		return fmt.Errorf("failed to marshal realm data: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/admin/realms", keycloakURL), strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create realm, status: %s, response: %s", resp.Status, string(respBody))
	}

	return nil
}

func getKeycloakAccessToken(keycloakURL, clientID, clientSecret string) (string, error) {
	data := fmt.Sprintf("client_id=%s&client_secret=%s&grant_type=client_credentials", clientID, clientSecret)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", keycloakURL), strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send token request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to fetch access token, status: %s, response: %s", resp.Status, string(respBody))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	return tokenResponse.AccessToken, nil
}
