package server

import (
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {

	mux := http.NewServeMux()
	v1 := http.NewServeMux()
	v1.HandleFunc("GET /", s.homeHandler)
	v1.HandleFunc("GET /users", s.GetUsersHandler)
	v1.HandleFunc("POST /users", s.CreateUserHandler)
	v1.HandleFunc("GET /users/{id}", s.GetUserHandler)
	v1.HandleFunc("PUT /users/{id}", s.UpdateUserHandler)
	v1.HandleFunc("DELETE /users/{id}", s.DeleteUserHandler)


	v1.HandleFunc("POST /register", s.RegisterUserHandler)  
	v1.HandleFunc("POST /login", s.LoginHandler)            
	
	mux.Handle("/v1/", http.StripPrefix("/v1", v1))
	
	return mux
}

