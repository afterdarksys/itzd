// Package client provides an HTTP client for the itz.agency API.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client talks to the itz.agency API.
type Client struct {
	base  string
	token string
	http  *http.Client
}

// New creates a new API client.
func New(apiEndpoint, token string) *Client {
	return &Client{
		base:  strings.TrimRight(apiEndpoint, "/"),
		token: token,
		http:  &http.Client{Timeout: 15 * time.Second},
	}
}

type HTTPError struct {
	Operation string
	Status    int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s failed with HTTP %d", e.Operation, e.Status)
}

func escapedPath(values ...string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = url.PathEscape(value)
	}
	return strings.Join(parts, "/")
}

func (c *Client) do(method, path string, body any) ([]byte, int, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.base+path, r)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	return data, resp.StatusCode, err
}

func (c *Client) get(path string) ([]byte, int, error) {
	return c.do("GET", path, nil)
}

func (c *Client) post(path string, body any) ([]byte, int, error) {
	return c.do("POST", path, body)
}

// ── Types ────────────────────────────────────────────────────────────────────

type ChallengeReq struct {
	Address string `json:"address"`
	Network string `json:"network"`
}

type ChallengeResp struct {
	Challenge string `json:"challenge"`
	Message   string `json:"message"`
	ExpiresAt int64  `json:"expires_at"`
}

type VerifyReq struct {
	Address   string `json:"address"`
	Signature string `json:"signature"`
	Challenge string `json:"challenge"`
}

type AuthResp struct {
	Success     bool   `json:"success"`
	AccessToken string `json:"access_token"`
	UserID      string `json:"user_id"`
	Address     string `json:"address"`
	Message     string `json:"message"`
}

type WalletRecord struct {
	Username      string `json:"username"`
	Address       string `json:"address"`
	Network       string `json:"network"`
	Namespace     string `json:"namespace"`
	Reference     string `json:"reference"`
	CAIP2         string `json:"caip2"`
	WalletAddress string `json:"wallet_address"`
	TXTRecord     string `json:"txt_record"`
	DNSName       string `json:"dns_name"`
	Verified      bool   `json:"verified"`
	DNSSynced     bool   `json:"dns_synced"`
}

type ResolveResp struct {
	Username string         `json:"username"`
	Address  string         `json:"address"`
	Wallets  []WalletRecord `json:"wallets"`
	Count    int            `json:"count"`
}

type ReverseResp struct {
	WalletAddress string         `json:"wallet_address"`
	Usernames     []WalletRecord `json:"usernames"`
	Count         int            `json:"count"`
}

type RegisterReq struct {
	Username      string `json:"username"`
	Network       string `json:"network"`
	WalletAddress string `json:"wallet_address"`
}

type UsernameResp struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Network       string `json:"network"`
	Namespace     string `json:"namespace"`
	FullAddress   string `json:"full_address"`
	DNSName       string `json:"dns_name"`
	WalletAddress string `json:"wallet_address"`
	Verified      bool   `json:"verified"`
	DNSSynced     bool   `json:"dns_synced"`
}

type NetworksResp struct {
	Networks map[string]string `json:"networks"`
}

// ── Methods ──────────────────────────────────────────────────────────────────

// Health checks if the API is reachable.
func (c *Client) Health() (map[string]any, error) {
	data, code, err := c.get("/health")
	if err != nil {
		return nil, err
	}
	if code < 200 || code >= 300 {
		return nil, &HTTPError{Operation: "health", Status: code}
	}
	var out map[string]any
	return out, json.Unmarshal(data, &out)
}

// Challenge requests an auth challenge for a wallet address.
func (c *Client) Challenge(address, network string) (*ChallengeResp, error) {
	data, code, err := c.post("/auth/challenge", ChallengeReq{Address: address, Network: network})
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("challenge failed (%d): %s", code, data)
	}
	var out ChallengeResp
	return &out, json.Unmarshal(data, &out)
}

// VerifyEthereum submits a signed challenge for Ethereum-compatible wallets.
func (c *Client) VerifyEthereum(address, challenge, signature string) (*AuthResp, error) {
	data, code, err := c.post("/auth/verify/ethereum", VerifyReq{
		Address:   address,
		Signature: signature,
		Challenge: challenge,
	})
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("verify failed (%d): %s", code, data)
	}
	var out AuthResp
	return &out, json.Unmarshal(data, &out)
}

// Resolve looks up all wallets for a username.
func (c *Client) Resolve(username string) (*ResolveResp, error) {
	data, code, err := c.get("/resolve/" + escapedPath(username))
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, fmt.Errorf("username %q not found", username)
	}
	if code != 200 {
		return nil, fmt.Errorf("resolve failed (%d): %s", code, data)
	}
	var out ResolveResp
	return &out, json.Unmarshal(data, &out)
}

// ResolveNetwork looks up a wallet for a username on a specific chain.
func (c *Client) ResolveNetwork(username, network string) (*WalletRecord, error) {
	data, code, err := c.get("/resolve/" + escapedPath(username, network))
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, fmt.Errorf("no %s wallet registered for %q", network, username)
	}
	if code != 200 {
		return nil, fmt.Errorf("resolve failed (%d): %s", code, data)
	}
	var out WalletRecord
	return &out, json.Unmarshal(data, &out)
}

// Whoami reverse-looks up a wallet address to find registered usernames.
func (c *Client) Whoami(walletAddress string) (*ReverseResp, error) {
	data, code, err := c.get("/resolve/wallet/" + escapedPath(walletAddress))
	if err != nil {
		return nil, err
	}
	if code == 404 {
		return nil, fmt.Errorf("no usernames found for %s", walletAddress)
	}
	if code != 200 {
		return nil, fmt.Errorf("lookup failed (%d): %s", code, data)
	}
	var out ReverseResp
	return &out, json.Unmarshal(data, &out)
}

// Register registers a username → wallet mapping.
func (c *Client) Register(username, network, walletAddress string) (*UsernameResp, error) {
	data, code, err := c.post("/username/register", RegisterReq{
		Username:      username,
		Network:       network,
		WalletAddress: walletAddress,
	})
	if err != nil {
		return nil, err
	}
	if code == 409 {
		return nil, fmt.Errorf("username %q already taken on %s", username, network)
	}
	if code == 402 {
		var errResp struct {
			Detail map[string]string `json:"detail"`
		}
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("plan limit: %s", errResp.Detail["message"])
	}
	if code != 201 {
		return nil, fmt.Errorf("register failed (%d): %s", code, data)
	}
	var out UsernameResp
	return &out, json.Unmarshal(data, &out)
}

// Check checks if a username is available.
func (c *Client) Check(username, network string) (map[string]any, error) {
	data, _, err := c.post("/username/check", map[string]string{
		"username": username,
		"network":  network,
	})
	if err != nil {
		return nil, err
	}
	var out map[string]any
	return out, json.Unmarshal(data, &out)
}

// MyUsernames returns all usernames owned by the authenticated user.
func (c *Client) MyUsernames() ([]UsernameResp, error) {
	data, code, err := c.get("/username/my-usernames")
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("failed (%d): %s", code, data)
	}
	var out []UsernameResp
	return out, json.Unmarshal(data, &out)
}
