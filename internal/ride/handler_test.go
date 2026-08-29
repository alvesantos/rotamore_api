package ride_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rotamore.com.br/api/internal/auth"
	"rotamore.com.br/api/internal/ride"
	"rotamore.com.br/api/internal/user"
)

func TestRideFlowAndIsolation(t *testing.T) {
	repo := ride.NewMemoryRepository()
	handler := ride.NewHandler(repo)

	driver1 := &user.User{ID: "d1", Name: "Driver 1", Type: user.TypeDriver}
	driver2 := &user.User{ID: "d2", Name: "Driver 2", Type: user.TypeDriver}

	t1, _ := auth.GenerateToken(driver1)
	t2, _ := auth.GenerateToken(driver2)

	// 1. Driver 1 creates a ride
	reqBody, _ := json.Marshal(map[string]interface{}{
		"customer_name":    "Maria Silva",
		"customer_phone":   "11999998888",
		"passengers_count": 2,
		"pickup":           "Aeroporto",
		"destination":      "Hotel Ponta Verde",
		"notes":            "Voo G3 1450",
		"ride_date":        "2026-09-10",
		"ride_time":        "14:30",
		"price":            120.0,
		"status":           "agendada",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/rides", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+t1)
	w := httptest.NewRecorder()
	handler.HandleRides(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("esperava 201, obteve %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Ride ride.Ride `json:"ride"`
	}
	_ = json.NewDecoder(w.Body).Decode(&res)

	// 2. Driver 2 lists rides -> 0 rides (isolated)
	req2 := httptest.NewRequest(http.MethodGet, "/api/rides", nil)
	req2.Header.Set("Authorization", "Bearer "+t2)
	w2 := httptest.NewRecorder()
	handler.HandleRides(w2, req2)

	var res2 struct {
		Rides []ride.Ride `json:"rides"`
	}
	_ = json.NewDecoder(w2.Body).Decode(&res2)
	if len(res2.Rides) != 0 {
		t.Errorf("esperava 0 corridas para driver 2, obteve %d", len(res2.Rides))
	}

	// 3. Driver 1 lists rides -> 1 ride
	req1 := httptest.NewRequest(http.MethodGet, "/api/rides", nil)
	req1.Header.Set("Authorization", "Bearer "+t1)
	w1 := httptest.NewRecorder()
	handler.HandleRides(w1, req1)

	var res1 struct {
		Rides []ride.Ride `json:"rides"`
	}
	_ = json.NewDecoder(w1.Body).Decode(&res1)
	if len(res1.Rides) != 1 {
		t.Errorf("esperava 1 corrida para driver 1, obteve %d", len(res1.Rides))
	}
}
