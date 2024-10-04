package middleware

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var validationUrl string
var excludedPaths []string

var services map[string]string

func LoadServiceMappings(configServices map[string]string) {
	services = configServices
}

func LoadValidationConfig(url string, pathsToExclude []string) {
	validationUrl = url
	excludedPaths = pathsToExclude
}

func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		for _, path := range excludedPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				// If the path is excluded, no token validation
				var targetUrl = resolveUrl(nil, r)
				r.Header.Add("targetUrl", targetUrl)
				next.ServeHTTP(w, r)
				return
			}
		}

		var companyId string = r.Header.Get("companyid")
		var token string = r.Header.Get("token")

		tokenInfo, err := requestTokenValidation(token, companyId)
		if err != nil {
			log.Printf("[!] Invalid Token for path [%s]\n", r.URL.Path)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Header.Del("token")
		var userIdStr string = strconv.FormatUint(*tokenInfo.UserId, 10)
		r.Header.Add("userid", userIdStr)

		var targetUrl = resolveUrl(tokenInfo.TenantInfo, r)
		r.Header.Add("targetUrl", targetUrl)

		next.ServeHTTP(w, r)
	})
}

func requestTokenValidation(token, companyId string) (*TokenTenantInfo, error) {
	client := http.Client{Timeout: 5 * time.Second}

	request, err := http.NewRequest("GET", validationUrl, nil)
	if err != nil {
		log.Printf("[!] Could not create validation token request\n")
		return nil, err
	}

	request.Header.Set("token", token)
	request.Header.Set("companyid", companyId)
	response, err := client.Do(request)
	if err != nil {
		log.Printf("[!] Could not connect to validation server\n")
		return nil, err
	}

	defer response.Body.Close()

	var validationResponse TokenTenantInfo
	if err := json.NewDecoder(response.Body).Decode(&validationResponse); err != nil {
		log.Printf("[!] Could not Parse validation response\n")
		return nil, err
	}

	return &validationResponse, nil
}

func resolveUrl(tenantInfo *Tenant, r *http.Request) string {
	var port uint32 = 1
	var trimmedPath string
	if strings.Contains(r.URL.Path, "/api/bmrm") {
		trimmedPath = strings.TrimPrefix(r.URL.Path, "/api/bmrm")
		port = tenantInfo.BmrmPort
	} else if strings.Contains(r.URL.Path, "/api/biz") {
		trimmedPath = strings.TrimPrefix(r.URL.Path, "/api/biz")
		port = tenantInfo.SgBizPort
	} else if strings.Contains(r.URL.Path, "/api/tally") {
		trimmedPath = strings.TrimPrefix(r.URL.Path, "/api/tally")
		port = tenantInfo.TallySyncPort
	}
	if port == 1 && strings.Contains(r.URL.Path, "/api/portal") {
		trimmedPath = strings.TrimPrefix(r.URL.Path, "/api/portal")
		baseUrl, exists := services["/api/portal"]
		if exists {
			return fmt.Sprintf("%s%s", baseUrl, trimmedPath)
		}
	}
	return fmt.Sprintf("%s:%v%s", tenantInfo.Host, port, trimmedPath)
}

type Tenant struct {
	ID            uint64
	CompanyGuid   string
	CompanyName   string
	Host          string
	BmrmPort      uint32
	SgBizPort     uint32
	TallySyncPort uint32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TokenTenantInfo struct {
	TenantInfo *Tenant
	UserId     *uint64
	Success    bool
	Message    string
}
