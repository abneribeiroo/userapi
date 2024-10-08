package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"userApi/internal/database"
)

func (s *Server) homeHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Main Handler"
	s.db.Health()
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
	log.Printf("Home Handler called")
}

func (s *Server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var User database.User
	if err := json.NewDecoder(r.Body).Decode(&User); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdUser, err := s.db.CreateUser(User.Username, User.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)

}

func (s *Server) GetUsersHandler(w http.ResponseWriter, r *http.Request) {

	users, err := s.db.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	oneUser, err := s.db.GetUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(oneUser)
}

func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var User database.User

	err = json.NewDecoder(r.Body).Decode(&User)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if User.Username == "" || User.Email == "" {
		http.Error(w, "Name and email are required", http.StatusBadRequest)
		return
	}

	resultMsg, err := s.db.UpdateUser(id, User.Username, User.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": resultMsg})
}

func (s *Server) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	msg, err := s.db.DeleteUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

func (s *Server) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Printf("reveived request to delete a User by ID \n")
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	log.Printf("reveived request to delete a User by ID \n")
}
