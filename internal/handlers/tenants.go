package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/vorticist/logger"
	"io"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"vortex.studio/account/internal/keycloak"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type TenantRequest struct {
	AdminUsername string `json:"admin-username"`
	BusinessName  string `json:"business-name"`
}

func CreateTenantHandler(w http.ResponseWriter, r *http.Request) {
	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract password from header
	passwordEncoded := r.Header.Get("X-Password")
	if passwordEncoded == "" {
		http.Error(w, "Missing X-Password header", http.StatusBadRequest)
		return
	}

	// Decode the Base64 password
	password, err := base64.StdEncoding.DecodeString(passwordEncoded)
	if err != nil {
		http.Error(w, "Invalid Base64 password", http.StatusBadRequest)
		return
	}

	// Decode Kubernetes secret
	secretPassword, err := getK8sSecretPassword(r, "default", "tenant-admin-secret", "admin-password")
	if err != nil {
		http.Error(w, "Error retrieving secret: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate password
	if string(password) != secretPassword {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Parse body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var tenantRequest TenantRequest
	if err := json.Unmarshal(body, &tenantRequest); err != nil {
		logger.Errorf("Error unmarshalling JSON: %v", err)
		logger.Errorf("Request body: %s", string(body))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if tenantRequest.AdminUsername == "" || tenantRequest.BusinessName == "" {
		logger.Errorf("Missing required fields: %v", tenantRequest)
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Simulate tenant creation (e.g., Kubernetes namespace, resources, etc.)
	log.Printf("Creating tenant for business: %s with admin: %s", tenantRequest.BusinessName, tenantRequest.AdminUsername)

	// Create a new Keycloak realm for the tenant
	adminPassword := "tenant-admin-password" // TODO: Generate a new admin password per tenant
	if err := keycloak.CreateKeycloakTenant(tenantRequest.BusinessName, tenantRequest.AdminUsername, adminPassword); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create tenant: %v", err), http.StatusInternalServerError)
		return
	}
	// Respond to the client
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Tenant '%s' created successfully", tenantRequest.BusinessName)))
}

func getK8sSecretPassword(r *http.Request, namespace, secretName, key string) (string, error) {
	var config *rest.Config
	var err error

	// Check if running in a Kubernetes cluster
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT") != "" {
		// Load in-cluster configuration
		config, err = rest.InClusterConfig()
		if err != nil {
			return "", fmt.Errorf("failed to load in-cluster config: %w", err)
		}
	} else {
		// Load local kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return "", fmt.Errorf("failed to load kubeconfig: %w", err)
		}
	}

	// Create Kubernetes client
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Retrieve the secret
	secret, err := clientset.CoreV1().Secrets(namespace).Get(r.Context(), secretName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve secret: %w", err)
	}

	// Decode the password
	encodedPassword, exists := secret.Data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' not found in secret", key)
	}

	return string(encodedPassword), nil
}
