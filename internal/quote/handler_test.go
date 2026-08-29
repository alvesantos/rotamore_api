package quote_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"rotamore.com.br/api/internal/auth"
	"rotamore.com.br/api/internal/quote"
	"rotamore.com.br/api/internal/user"
)

func TestQuoteIsolationAndCRUD(t *testing.T) {
	repo := quote.NewMemoryRepository()
	handler := quote.NewHandler(repo)

	// Create 2 test drivers
	driver1 := &user.User{ID: "driver-1-uuid", Name: "Driver 1", Type: user.TypeDriver}
	driver2 := &user.User{ID: "driver-2-uuid", Name: "Driver 2", Type: user.TypeDriver}

	token1, _ := auth.GenerateToken(driver1)
	token2, _ := auth.GenerateToken(driver2)

	// 1. Driver 1 creates a quote
	quoteBody, _ := json.Marshal(map[string]interface{}{
		"pickup":      "Ponta Verde",
		"destination": "Aeroporto",
		"price":       85.0,
	})
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/quotes", bytes.NewReader(quoteBody))
	reqCreate.Header.Set("Authorization", "Bearer "+token1)
	wCreate := httptest.NewRecorder()
	handler.HandleQuotes(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("esperava 201, obteve %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var createdResp struct {
		Quote quote.Quote `json:"quote"`
	}
	_ = json.NewDecoder(wCreate.Body).Decode(&createdResp)

	// 2. Driver 2 lists quotes -> should be EMPTY (strict user isolation)
	reqList2 := httptest.NewRequest(http.MethodGet, "/api/quotes", nil)
	reqList2.Header.Set("Authorization", "Bearer "+token2)
	wList2 := httptest.NewRecorder()
	handler.HandleQuotes(wList2, reqList2)

	var listResp2 struct {
		Quotes []quote.Quote `json:"quotes"`
	}
	_ = json.NewDecoder(wList2.Body).Decode(&listResp2)
	if len(listResp2.Quotes) != 0 {
		t.Fatalf("esperava 0 orçamentos para o driver 2, obteve %d", len(listResp2.Quotes))
	}

	// 3. Driver 1 lists quotes -> should have 1 quote
	reqList1 := httptest.NewRequest(http.MethodGet, "/api/quotes", nil)
	reqList1.Header.Set("Authorization", "Bearer "+token1)
	wList1 := httptest.NewRecorder()
	handler.HandleQuotes(wList1, reqList1)

	var listResp1 struct {
		Quotes []quote.Quote `json:"quotes"`
	}
	_ = json.NewDecoder(wList1.Body).Decode(&listResp1)
	if len(listResp1.Quotes) != 1 {
		t.Fatalf("esperava 1 orçamento para o driver 1, obteve %d", len(listResp1.Quotes))
	}

	// 4. Driver 1 deletes the quote
	reqDel := httptest.NewRequest(http.MethodDelete, "/api/quotes?id="+createdResp.Quote.ID, nil)
	reqDel.Header.Set("Authorization", "Bearer "+token1)
	wDel := httptest.NewRecorder()
	handler.HandleQuotes(wDel, reqDel)

	if wDel.Code != http.StatusOK {
		t.Fatalf("esperava status 200 na deleção, obteve %d: %s", wDel.Code, wDel.Body.String())
	}
}
