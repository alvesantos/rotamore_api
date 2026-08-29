package auth_test

import (
	"bytes"
	"context"
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

	// 1. Test Driver Login with client_type = "driver" -> SUCCESS
	driverBody, _ := json.Marshal(map[string]string{
		"identifier":  "ricberns@gmail.com",
		"password":    "1254101254@Abc",
		"client_type": "driver",
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

	// 2. Test Inactive Driver Login -> FORBIDDEN (403)
	driverResp.User.Status = "inactive"
	_ = repo.Update(context.Background(), driverResp.User)

	reqInactive := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(driverBody))
	wInactive := httptest.NewRecorder()
	handler.Login(wInactive, reqInactive)

	if wInactive.Code != http.StatusForbidden {
		t.Fatalf("esperava status 403 para motorista inativo, obteve %d: %s", wInactive.Code, wInactive.Body.String())
	}

	// Restore active
	driverResp.User.Status = "active"
	_ = repo.Update(context.Background(), driverResp.User)

	// 3. Test Admin Login trying to access Driver portal (client_type = "driver") -> FORBIDDEN (403)
	adminBody, _ := json.Marshal(map[string]string{
		"identifier":  "rogab@admin.com",
		"password":    "r0g4b@2026!",
		"client_type": "driver",
	})
	reqAdmin := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(adminBody))
	wAdmin := httptest.NewRecorder()
	handler.Login(wAdmin, reqAdmin)

	if wAdmin.Code != http.StatusForbidden {
		t.Fatalf("esperava status 403 para login de admin no app de motorista, obteve %d: %s", wAdmin.Code, wAdmin.Body.String())
	}

	// 4. Test Invalid Password
	invalidBody, _ := json.Marshal(map[string]string{
		"identifier": "ricberns@gmail.com",
		"password":   "senha_errada",
	})
	reqInvalid := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(invalidBody))
	wInvalid := httptest.NewRecorder()
	handler.Login(wInvalid, reqInvalid)

	if wInvalid.Code != http.StatusUnauthorized {
		t.Errorf("esperava status 401 para senha incorreta, obteve %d", wInvalid.Code)
	}

	// 5. Test Me endpoint
	reqMe := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+driverResp.Token)
	wMe := httptest.NewRecorder()
	handler.Me(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("esperava status 200 para endpoint /me, obteve %d: %s", wMe.Code, wMe.Body.String())
	}

	// 6. Test UpdateProfile endpoint
	updateBody, _ := json.Marshal(map[string]string{
		"name":     "Ricardo Atualizado",
		"lastname": "Berns Novo",
		"phone":    "11977777777",
	})
	reqUpdate := httptest.NewRequest(http.MethodPut, "/api/auth/profile", bytes.NewReader(updateBody))
	reqUpdate.Header.Set("Authorization", "Bearer "+driverResp.Token)
	wUpdate := httptest.NewRecorder()
	handler.UpdateProfile(wUpdate, reqUpdate)

	if wUpdate.Code != http.StatusOK {
		t.Fatalf("esperava status 200 para update profile, obteve %d: %s", wUpdate.Code, wUpdate.Body.String())
	}
}
