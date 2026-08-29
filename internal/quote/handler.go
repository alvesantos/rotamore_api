package quote

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"rotamore.com.br/api/internal/auth"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

type CreateQuoteRequest struct {
	Category    string   `json:"category"` // 'transfer' | 'passeio'
	Pickup      string   `json:"pickup"`
	Destination string   `json:"destination"`
	Stops       []string `json:"stops"`
	Price       float64  `json:"price"`
}

func getUserIDFromAuth(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", http.ErrNoCookie
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", http.ErrNoCookie
	}
	claims, err := auth.ValidateToken(parts[1])
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func (h *Handler) HandleQuotes(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Não autorizado"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		limitStr := r.URL.Query().Get("limit")
		limit := 10
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}

		quotes, err := h.repo.FindByUserID(r.Context(), userID, limit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao buscar orçamentos"})
			return
		}
		if quotes == nil {
			quotes = []Quote{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"quotes": quotes})

	case http.MethodPost:
		var req CreateQuoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Corpo da requisição inválido"})
			return
		}

		category := strings.ToLower(strings.TrimSpace(req.Category))
		if category != "passeio" {
			category = "transfer"
		}

		destination := strings.TrimSpace(req.Destination)
		if category == "passeio" && destination == "" && len(req.Stops) > 0 {
			destination = strings.Join(req.Stops, " ➔ ")
		}

		if req.Pickup == "" || destination == "" || req.Price <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Embarque, destino e valor válido são obrigatórios"})
			return
		}

		if req.Stops == nil {
			req.Stops = []string{}
		}

		q := &Quote{
			UserID:      userID,
			Category:    category,
			Pickup:      strings.TrimSpace(req.Pickup),
			Destination: destination,
			Stops:       req.Stops,
			Price:       req.Price,
		}

		if err := h.repo.Create(r.Context(), q); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Orçamento salvo com sucesso",
			"quote":   q,
		})

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ID do orçamento é obrigatório"})
			return
		}

		if err := h.repo.Delete(r.Context(), id, userID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao remover orçamento"})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Orçamento removido com sucesso"})

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}
