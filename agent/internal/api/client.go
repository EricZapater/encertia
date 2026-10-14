package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(baseURL string) *Client {
	cleanURL := strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL: cleanURL,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type VerifyDeviceRequest struct {
	HardwareUUIDHash string `json:"hardwareUuidHash"`
}

type VerifyDeviceResponse struct {
	BootstrapToken   string `json:"bootstrapToken"`
	TeacherEmail     string `json:"teacherEmail"`
	DeviceName       string `json:"deviceName"`
	ExpiresInSeconds int    `json:"expiresInSeconds"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Tokens struct {
		AccessToken string `json:"accessToken"`
	} `json:"tokens"`
}

type RegisterDeviceRequest struct {
	HardwareUUIDHash string `json:"hardwareUuidHash"`
	Name             string `json:"name"`
	TeacherEmail     string `json:"teacherEmail"`
}

type DeviceSummary struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	TeacherID    string  `json:"teacherId"`
	TeacherName  string  `json:"teacherName"`
	TeacherEmail string  `json:"teacherEmail"`
	IsActive     bool    `json:"isActive"`
	CreatedAt    string  `json:"createdAt"`
}

type APIErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// VerifyDevice calls POST /devices/verify with the hardware UUID hash.
func (c *Client) VerifyDevice(hardwareHash string) (*VerifyDeviceResponse, error) {
	reqBody := VerifyDeviceRequest{
		HardwareUUIDHash: hardwareHash,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/devices/verify", c.BaseURL)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error de connexió amb el servidor Encertia (%s): %w", url, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIErrorResponse
		if json.Unmarshal(respBytes, &apiErr) == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("servidor (%d): %s (codi: %s)", resp.StatusCode, apiErr.Error.Message, apiErr.Error.Code)
		}
		return nil, fmt.Errorf("ordinador no autoritzat (codi HTTP %d)", resp.StatusCode)
	}

	var verifyResp VerifyDeviceResponse
	if err := json.Unmarshal(respBytes, &verifyResp); err != nil {
		return nil, fmt.Errorf("error processant la resposta del servidor: %w", err)
	}

	return &verifyResp, nil
}

// LoginAdmin authenticates the administrator and returns the access token.
func (c *Client) LoginAdmin(email, password string) (string, error) {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/auth/login", c.BaseURL)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("error connectant amb el servidor: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIErrorResponse
		if json.Unmarshal(respBytes, &apiErr) == nil && apiErr.Error.Message != "" {
			return "", fmt.Errorf("error del servidor (HTTP %d, %s): %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
		}
		rawBody := strings.TrimSpace(string(respBytes))
		if rawBody != "" {
			return "", fmt.Errorf("error del servidor (HTTP %d): %s", resp.StatusCode, rawBody)
		}
		return "", fmt.Errorf("credencials incorrectes (HTTP %d)", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(respBytes, &loginResp); err != nil {
		return "", err
	}

	if loginResp.Tokens.AccessToken == "" {
		return "", errors.New("el servidor no ha retornat cap token d'accés vàlid")
	}

	return loginResp.Tokens.AccessToken, nil
}

// RegisterDevice registers this physical machine for a designated teacher.
func (c *Client) RegisterDevice(adminToken, hardwareHash, deviceName, teacherEmail string) (*DeviceSummary, error) {
	reqBody := RegisterDeviceRequest{
		HardwareUUIDHash: hardwareHash,
		Name:             deviceName,
		TeacherEmail:     teacherEmail,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/admin/devices", c.BaseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error connectant amb el servidor: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var apiErr APIErrorResponse
		if json.Unmarshal(respBytes, &apiErr) == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("error del servidor (HTTP %d, %s): %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
		}
		rawBody := strings.TrimSpace(string(respBytes))
		if rawBody != "" {
			return nil, fmt.Errorf("error donant d'alta l'equip (HTTP %d): %s", resp.StatusCode, rawBody)
		}
		return nil, fmt.Errorf("error donant d'alta l'equip (HTTP %d)", resp.StatusCode)
	}

	var summary DeviceSummary
	if err := json.Unmarshal(respBytes, &summary); err != nil {
		return nil, fmt.Errorf("error processant la confirmació de l'equip: %w", err)
	}

	return &summary, nil
}
