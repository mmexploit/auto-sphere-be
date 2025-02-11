package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Register routes
	// mux.HandleFunc("/", s.HelloWorldHandler)

	mux.HandleFunc("GET /users/{id}", s.userGetOne)
	mux.HandleFunc("POST /users", s.userCreate)
	mux.HandleFunc("DELETE /users/{id}", s.userDelete)
	mux.HandleFunc("PATCH /users/{id}", s.userPatch)
	mux.HandleFunc("GET /users", s.getUsers)
	mux.HandleFunc("POST /users/login", s.login)
	mux.HandleFunc("POST /users/token/refresh", s.refreshToken)

	mux.HandleFunc("GET /shops/{id}", s.shopGetOne)
	mux.HandleFunc("POST /shops", s.shopCreate)
	mux.HandleFunc("DELETE /shops/{id}", s.shopDelete)
	mux.HandleFunc("PATCH /shops/approval/{id}", s.updateAppoval)
	mux.HandleFunc("PATCH /shops/{id}", s.shopPatch)
	mux.HandleFunc("GET /shops", s.getShops)

	mux.HandleFunc("GET /categories/{id}", s.catGetOne)
	mux.HandleFunc("POST /categories", s.catCreate)
	mux.HandleFunc("DELETE /categories/{id}", s.catDelete)
	mux.HandleFunc("PUT /categories/{id}", s.catPut)
	mux.HandleFunc("GET /categories", s.catGetAll)

	mux.HandleFunc("GET /values/{id}", s.catMemberGetOne)
	mux.HandleFunc("POST /values", s.catMemberCreate)
	mux.HandleFunc("DELETE /values/{id}", s.catMemberDelete)
	mux.HandleFunc("PATCH /values/{id}", s.catMemberPut)
	// mux.HandleFunc("GET /values", s.catMemberGetAll)
	mux.HandleFunc("/values", s.RoleMiddleware(s.catMemberGetAll, "ADMIN"))

	mux.HandleFunc("POST /shopCategories", s.scCreate)

	mux.HandleFunc("/health", s.healthHandler)

	// Wrap the mux with CORS middleware
	return s.corsMiddleware(mux)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Replace "*" with specific origins if needed
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are required

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed with the next handler
		next.ServeHTTP(w, r)
	})
}

// func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
// 	resp := map[string]string{"message": "Hello World"}
// 	jsonResp, err := json.Marshal(resp)
// 	if err != nil {
// 		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	if _, err := w.Write(jsonResp); err != nil {
// 		log.Printf("Failed to write response: %v", err)
// 	}
// }

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := json.Marshal(s.db.Health())
	if err != nil {
		http.Error(w, "Failed to marshal health check response", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}
