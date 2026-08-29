package ride

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

type CreateRideRequest struct {
	VehicleID       *string `json:"vehicle_id,omitempty"`
	CustomerName    string  `json:"customer_name"`
	CustomerPhone   string  `json:"customer_phone"`
	PassengersCount int     `json:"passengers_count"`
	Pickup          string  `json:"pickup"`
	Destination     string  `json:"destination"`
	Notes           string  `json:"notes"`
	RideDate        string  `json:"ride_date"`
	RideTime        string  `json:"ride_time"`
	Price           float64 `json:"price"`
	Status          string  `json:"status"`
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

func (h *Handler) HandleRides(w http.ResponseWriter, r *http.Request) {
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
		yearStr := r.URL.Query().Get("year")

		limit := 0
		year := 0
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
		if y, err := strconv.Atoi(yearStr); err == nil && y > 0 {
			year = y
		}

		rides, err := h.repo.FindByUserID(r.Context(), userID, limit, year)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao buscar corridas"})
			return
		}
		if rides == nil {
			rides = []Ride{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"rides": rides})

	case http.MethodPost:
		var req CreateRideRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Corpo da requisição inválido"})
			return
		}

		if req.CustomerName == "" || req.CustomerPhone == "" || req.Pickup == "" || req.Destination == "" || req.RideDate == "" || req.RideTime == "" || req.Price <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Nome do cliente, telefone, trajeto, data, horário e valor são obrigatórios"})
			return
		}

		if req.PassengersCount <= 0 {
			req.PassengersCount = 1
		}
		if req.Status == "" {
			req.Status = "agendada"
		}

		item := &Ride{
			UserID:          userID,
			VehicleID:       req.VehicleID,
			CustomerName:    strings.TrimSpace(req.CustomerName),
			CustomerPhone:   strings.TrimSpace(req.CustomerPhone),
			PassengersCount: req.PassengersCount,
			Pickup:          strings.TrimSpace(req.Pickup),
			Destination:     strings.TrimSpace(req.Destination),
			Notes:           strings.TrimSpace(req.Notes),
			RideDate:        strings.TrimSpace(req.RideDate),
			RideTime:        strings.TrimSpace(req.RideTime),
			Price:           req.Price,
			Status:          req.Status,
		}

		if err := h.repo.Create(r.Context(), item); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Corrida salva com sucesso",
			"ride":    item,
		})

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ID da corrida é obrigatório"})
			return
		}

		if err := h.repo.Delete(r.Context(), id, userID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao remover corrida"})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Corrida removida com sucesso"})

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleRideDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Não autorizado"})
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID da corrida é obrigatório"})
		return
	}

	item, err := h.repo.FindByID(r.Context(), id, userID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Corrida não encontrada"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"ride": item})
}
