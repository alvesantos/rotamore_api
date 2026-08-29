package vehicle

import (
	"encoding/json"
	"net/http"
	"strings"

	"rotamore.com.br/api/internal/auth"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

type CreateVehicleRequest struct {
	Name  string `json:"name"`
	Brand string `json:"brand"`
	Plate string `json:"plate"`
	Color string `json:"color"`
	Year  int    `json:"year"`
}

type SetActiveRequest struct {
	ID string `json:"id"`
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

func (h *Handler) HandleVehicles(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Não autorizado"})
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		vehicles, err := h.repo.FindByUserID(r.Context(), userID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao buscar veículos"})
			return
		}
		if vehicles == nil {
			vehicles = []Vehicle{}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"vehicles": vehicles})

	case http.MethodPost:
		var req CreateVehicleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Corpo da requisição inválido"})
			return
		}

		if req.Name == "" || req.Brand == "" || req.Plate == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Nome/modelo, marca e placa são obrigatórios"})
			return
		}

		v := &Vehicle{
			UserID: userID,
			Name:   req.Name,
			Brand:  req.Brand,
			Plate:  req.Plate,
			Color:  req.Color,
			Year:   req.Year,
		}

		if err := h.repo.Create(r.Context(), v); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Veículo cadastrado com sucesso",
			"vehicle": v,
		})

	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "ID do veículo é obrigatório"})
			return
		}

		if err := h.repo.Delete(r.Context(), id, userID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao remover veículo"})
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "Veículo removido com sucesso"})

	default:
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) HandleSetActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromAuth(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Não autorizado"})
		return
	}

	var req SetActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "ID do veículo é obrigatório"})
		return
	}

	if err := h.repo.SetActive(r.Context(), req.ID, userID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erro ao definir veículo ativo"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Veículo ativo atualizado com sucesso"})
}
