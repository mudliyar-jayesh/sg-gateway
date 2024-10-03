package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var validationUrl string
var excludedPaths []string

func LoadValidationConfig(url string, pathsToExclude []string) {
	validationUrl = url
	excludedPaths = pathsToExclude
}

func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		for _, path := range excludedPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				// If the path is excluded, no token validation
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
