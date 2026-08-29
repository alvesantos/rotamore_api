package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"rotamore.com.br/api/internal/user"
)

type Handler struct {
	repo user.Repository
}

func NewHandler(repo user.Repository) *Handler {
	return &Handler{repo: repo}
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
	ClientType string `json:"client_type"` // e.g. "driver"
}

type LoginResponse struct {
	Token string     `json:"token"`
	User  *user.User `json:"user"`
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	LastName string `json:"lastname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Document string `json:"document"`
	State    string `json:"state"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Formato de requisição inválido"})
		return
	}

	identifier := req.Identifier
	if identifier == "" {
		if req.Email != "" {
			identifier = req.Email
		} else if req.Phone != "" {
			identifier = req.Phone
		}
	}

	if identifier == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Identificador (e-mail/celular) e senha são obrigatórios"})
		return
	}

	u, err := h.repo.FindByEmailOrPhone(r.Context(), identifier)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "E-mail/celular ou senha incorretos"})
		return
	}

	if !user.CheckPassword(u.PasswordHash, req.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "E-mail/celular ou senha incorretos"})
		return
	}

	// Check if user is active
	if strings.ToLower(u.Status) == "inactive" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Sua conta está inativa. Entre em contato com o suporte ou administrador.",
		})
		return
	}

	// Check if login is restricted to driver
	clientType := req.ClientType
	if clientType == "" {
		clientType = r.Header.Get("X-Client-Type")
	}

	if clientType == "driver" && u.Type != user.TypeDriver {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Acesso exclusivo para motoristas parceiros. Administradores devem acessar o Backoffice.",
		})
		return
	}

	token, err := GenerateToken(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Erro ao gerar token de autenticação"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
		User:  u,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Token de autorização não informado"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Formato do header Authorization inválido (esperado: Bearer <token>)"})
		return
	}

	claims, err := ValidateToken(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Token inválido ou expirado"})
		return
	}

	u, err := h.repo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Usuário não encontrado"})
		return
	}

	clientType := r.Header.Get("X-Client-Type")
	if clientType == "driver" && u.Type != user.TypeDriver {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: "Acesso exclusivo para motoristas parceiros. Administradores devem acessar o Backoffice.",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": u,
	})
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch && r.Method != http.MethodPost {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Token de autorização não informado"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Formato de token inválido"})
		return
	}

	claims, err := ValidateToken(parts[1])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Token inválido ou expirado"})
		return
	}

	u, err := h.repo.FindByID(r.Context(), claims.UserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Usuário não encontrado"})
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Corpo da requisição inválido"})
		return
	}

	if req.Name != "" {
		u.Name = strings.TrimSpace(req.Name)
	}
	if req.LastName != "" {
		u.LastName = strings.TrimSpace(req.LastName)
	}
	if req.Phone != "" {
		u.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Email != "" {
		u.Email = strings.TrimSpace(req.Email)
	}
	if req.Document != "" {
		u.Document = strings.TrimSpace(req.Document)
	}
	if req.State != "" {
		u.State = strings.TrimSpace(req.State)
	}

	if err := h.repo.Update(r.Context(), u); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Erro ao atualizar dados do usuário"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Perfil atualizado com sucesso",
		"user":    u,
	})
}
