package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"rotamore.com.br/api/database"
)

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status": "API Running",
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: arquivo .env não encontrado, usando variáveis do sistema")
	}

	pool, err := database.Connect()

	if err != nil {
		log.Fatalf("Falha ao conectar no banco: %v", err)
	}

	defer pool.Close()

	log.Println("Conexão com o banco de dados estabelecida com sucesso!")

	http.HandleFunc("/", statusHandler)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
