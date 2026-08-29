package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"rotamore.com.br/api/database"
	"rotamore.com.br/api/internal/auth"
	"rotamore.com.br/api/internal/user"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Client-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "API Running",
		"project": "Rota+",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: arquivo .env não encontrado, usando variáveis do sistema")
	}

	pool, err := database.Connect()
	if err != nil {
		log.Fatalf("❌ Erro fatal: Não foi possível conectar ao banco de dados PostgreSQL: %v", err)
	}
	defer pool.Close()

	log.Println("✓ Conexão com o banco de dados PostgreSQL estabelecida com sucesso!")
	userRepo := user.NewPostgresRepository(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := userRepo.SeedDefaultUsers(ctx); err != nil {
		log.Printf("Aviso ao rodar seed no PostgreSQL: %v", err)
	}
	cancel()

	authHandler := auth.NewHandler(userRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/", statusHandler)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/me", authHandler.Me)
	mux.HandleFunc("/api/auth/profile", authHandler.UpdateProfile)

	handlerWithCORS := corsMiddleware(mux)

	log.Println("🚀 Servidor Rota+ API rodando em http://localhost:8080")
	log.Println("📋 Rotas disponíveis:")
	log.Println("  - POST /api/auth/login")
	log.Println("  - GET  /api/auth/me")
	log.Println("  - PUT  /api/auth/profile")
	log.Println("  - GET  /")

	log.Fatal(http.ListenAndServe(":8080", handlerWithCORS))
}
