package githubapp

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
  "sync"


	"github.com/golang-jwt/jwt/v5"
)
type App struct {
	AppID string
	Key   *rsa.PrivateKey
}

var (
	instance *App
	once     sync.Once
	initErr  error
)

func Get() (*App, error) {
	once.Do(func() {
		appID := strings.TrimSpace(os.Getenv("GITHUB_APP_ID"))
		keyPath := strings.TrimSpace(os.Getenv("GITHUB_APP_PRIVATE_KEY_PATH"))

		if appID == "" || keyPath == "" {
			initErr = fmt.Errorf("missing GITHUB_APP_ID or GITHUB_APP_PRIVATE_KEY_PATH")
			return
		}

		b, err := os.ReadFile(keyPath)
		if err != nil {
			initErr = fmt.Errorf("read github app key: %w", err)
			return
		}

		key, err := jwt.ParseRSAPrivateKeyFromPEM(b)
		if err != nil {
			initErr = fmt.Errorf("parse github app private key: %w", err)
			return
		}

		instance = &App{
			AppID: appID,
			Key:   key,
		}
	})

	return instance, initErr
}

func appJWT(app_id string, key *rsa.PrivateKey) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iat": now.Add(-1 * time.Minute).Unix(),
		"exp": now.Add(9 * time.Minute).Unix(),
		"iss": app_id,
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return tok.SignedString(key)
}

func InstallationTokenForRepo(ctx context.Context, owner, repo string) (string, error) {
  app, err := Get()
	if err != nil {
		return "", err
	}

	j, err := appJWT(app.AppID, app.Key)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://api.github.com/repos/%s/%s/installation", owner, repo),
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+j)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("installation lookup failed (%d): %s", resp.StatusCode, body)
	}

	var inst struct{ ID int64 `json:"id"` }
	if err := json.NewDecoder(resp.Body).Decode(&inst); err != nil {
		return "", err
	}

	req, _ = http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", inst.ID),
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+j)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 && resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token create failed (%d): %s", resp.StatusCode, body)
	}

	var tok struct{ Token string `json:"token"` }
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", err
	}
	if tok.Token == "" {
		return "", fmt.Errorf("empty installation token returned")
	}
	return tok.Token, nil
}

