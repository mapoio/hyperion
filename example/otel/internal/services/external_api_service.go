package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mapoio/hyperion"
)

// ExternalAPIService demonstrates HTTP client tracing with public APIs
type ExternalAPIService struct {
	httpClient      *http.Client
	apiCallCounter  hyperion.Counter
	apiCallDuration hyperion.Histogram
}

// NewExternalAPIService creates a new ExternalAPIService instance
func NewExternalAPIService(
	httpClient *http.Client,
	meter hyperion.Meter,
) *ExternalAPIService {
	return &ExternalAPIService{
		httpClient: httpClient,
		apiCallCounter: meter.Counter("external_api.calls.total",
			hyperion.WithMetricDescription("Total external API calls"),
			hyperion.WithMetricUnit("1"),
		),
		apiCallDuration: meter.Histogram("external_api.call.duration",
			hyperion.WithMetricDescription("External API call duration in milliseconds"),
			hyperion.WithMetricUnit("ms"),
		),
	}
}

// GetRandomUser fetches a random user from JSONPlaceholder API
func (s *ExternalAPIService) GetRandomUser(hctx hyperion.Context, userID int) (user map[string]any, err error) {
	startTime := time.Now()
	apiName := "jsonplaceholder"
	endpoint := "users"

	hctx, end := hctx.UseIntercept("ExternalAPIService", "GetRandomUser")
	defer end(&err)

	defer func() {
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.apiCallCounter.Add(hctx, 1,
			hyperion.String("api", apiName),
			hyperion.String("endpoint", endpoint),
			hyperion.String("status", status),
		)
		s.apiCallDuration.Record(hctx, duration,
			hyperion.String("api", apiName),
			hyperion.String("endpoint", endpoint),
			hyperion.String("status", status),
		)
	}()

	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/users/%d", userID)
	hctx.Logger().Info("calling external API", "url", url)

	req, err := http.NewRequestWithContext(hctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		hctx.Logger().Error("API call failed", "error", err)
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hctx.Logger().Warn("API returned non-200 status", "status", resp.StatusCode)
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	hctx.Logger().Info("external API call successful",
		"api", apiName,
		"user_name", user["name"],
	)

	return user, nil
}

// GetRandomPost fetches a random post from JSONPlaceholder API
func (s *ExternalAPIService) GetRandomPost(hctx hyperion.Context, postID int) (post map[string]any, err error) {
	startTime := time.Now()
	apiName := "jsonplaceholder"
	endpoint := "posts"

	hctx, end := hctx.UseIntercept("ExternalAPIService", "GetRandomPost")
	defer end(&err)

	defer func() {
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.apiCallCounter.Add(hctx, 1,
			hyperion.String("api", apiName),
			hyperion.String("endpoint", endpoint),
			hyperion.String("status", status),
		)
		s.apiCallDuration.Record(hctx, duration,
			hyperion.String("api", apiName),
			hyperion.String("endpoint", endpoint),
			hyperion.String("status", status),
		)
	}()

	url := fmt.Sprintf("https://jsonplaceholder.typicode.com/posts/%d", postID)
	hctx.Logger().Info("calling external API", "url", url)

	req, err := http.NewRequestWithContext(hctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		hctx.Logger().Error("API call failed", "error", err)
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hctx.Logger().Warn("API returned non-200 status", "status", resp.StatusCode)
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if err := json.Unmarshal(body, &post); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	hctx.Logger().Info("external API call successful",
		"api", apiName,
		"post_title", post["title"],
	)

	return post, nil
}

// GetIPInfo fetches IP geolocation info from ipapi.co
func (s *ExternalAPIService) GetIPInfo(hctx hyperion.Context) (info map[string]any, err error) {
	startTime := time.Now()
	apiName := "ipapi"
	endpoint := "json"

	hctx, end := hctx.UseIntercept("ExternalAPIService", "GetIPInfo")
	defer end(&err)

	defer func() {
		duration := float64(time.Since(startTime).Milliseconds())
		status := "success"
		if err != nil {
			status = "error"
		}

		s.apiCallCounter.Add(hctx, 1,
			hyperion.String("api", apiName),
			hyperion.String("endpoint", endpoint),
			hyperion.String("status", status),
		)
		s.apiCallDuration.Record(hctx, duration,
			hyperion.String("api", apiName),
			hyperion.String("endpoint", endpoint),
			hyperion.String("status", status),
		)
	}()

	url := "https://ipapi.co/json/"
	hctx.Logger().Info("calling external API", "url", url)

	req, err := http.NewRequestWithContext(hctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		hctx.Logger().Error("API call failed", "error", err)
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hctx.Logger().Warn("API returned non-200 status", "status", resp.StatusCode)
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	hctx.Logger().Info("external API call successful",
		"api", apiName,
		"city", info["city"],
		"country", info["country_name"],
	)

	return info, nil
}
