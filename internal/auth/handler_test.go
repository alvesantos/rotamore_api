package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rotamore.com.br/api/internal/auth"
	"rotamore.com.br/api/internal/user"
)

func TestAuthFlow(t *testing.T) {
	repo := user.NewMemoryRepository()
	handler := auth.NewHandler(repo)

	// 1. Test Admin Login
	adminBody, _ := json.Marshal(map[string]string{
		"identifier": "rogab@admin.com",
		"password":   "r0g4b@2026!",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(adminBody))
	w := httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("esperava status 200 para login do admin, obteve %d: %s", w.Code, w.Body.String())
	}

	var adminResp auth.LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&adminResp); err != nil {
		t.Fatalf("erro ao decodificar resposta do admin: %v", err)
	}

	if adminResp.User.Type != user.TypeAdmin {
		t.Errorf("esperava tipo 'admin', obteve '%s'", adminResp.User.Type)
	}
	if adminResp.Token == "" {
		t.Error("token não retornado")
	}

	// 2. Test Driver Login
	driverBody, _ := json.Marshal(map[string]string{
		"identifier": "ricberns@gmail.com",
		"password":   "1254101254@Abc",
	})
	reqDriver := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(driverBody))
	wDriver := httptest.NewRecorder()
	handler.Login(wDriver, reqDriver)

	if wDriver.Code != http.StatusOK {
		t.Fatalf("esperava status 200 para login do motorista, obteve %d: %s", wDriver.Code, wDriver.Body.String())
	}

	var driverResp auth.LoginResponse
	if err := json.NewDecoder(wDriver.Body).Decode(&driverResp); err != nil {
		t.Fatalf("erro ao decodificar resposta do motorista: %v", err)
	}

	if driverResp.User.Type != user.TypeDriver {
		t.Errorf("esperava tipo 'driver', obteve '%s'", driverResp.User.Type)
	}

	// 3. Test Invalid Password
	invalidBody, _ := json.Marshal(map[string]string{
		"identifier": "rogab@admin.com",
		"password":   "senha_incorreta",
	})
	reqInvalid := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(invalidBody))
	wInvalid := httptest.NewRecorder()
	handler.Login(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusUnauthorized {
		t.Errorf("esperava status 401 para senha incorreta, obteve %d", wInvalid.Code)
	}

	// 4. Test Me endpoint
	reqMe := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+adminResp.Token)
	wMe := httptest.NewRecorder()
	handler.Me(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("esperava status 200 para endpoint /me, obteve %d: %s", wMe.Code, wMe.Body.String())
	}
}
