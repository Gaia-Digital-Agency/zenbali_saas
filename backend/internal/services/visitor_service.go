package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
)

type VisitorService struct {
	repos *repository.Repositories
}

func NewVisitorService(repos *repository.Repositories) *VisitorService {
	return &VisitorService{repos: repos}
}

func (s *VisitorService) TrackVisitor(ctx context.Context, ipAddress, userAgent string) error {
	// Get location from IP (using free IP-API service)
	country, city := s.getLocationFromIP(ipAddress)

	visitor := &models.Visitor{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Country:   country,
		City:      city,
	}

	return s.repos.Visitor.Create(ctx, visitor)
}

func (s *VisitorService) GetStats(ctx context.Context) (*models.VisitorStats, error) {
	return s.repos.Visitor.GetStats(ctx)
}

func (s *VisitorService) getLocationFromIP(ip string) (country, city string) {
	// Skip for localhost/private IPs
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "Local", "Local"
	}

	// Use free IP geolocation API
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://ip-api.com/json/" + ip + "?fields=country,city")
	if err != nil {
		return "Unknown", "Unknown"
	}
	defer resp.Body.Close()

	var result struct {
		Country string `json:"country"`
		City    string `json:"city"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "Unknown", "Unknown"
	}

	if result.Country == "" {
		result.Country = "Unknown"
	}
	if result.City == "" {
		result.City = "Unknown"
	}

	return result.Country, result.City
}
