package handlers

import (
	"encoding/json"
	"net/http"
	"news-restapi/models"
	"news-restapi/storage"
	"news-restapi/utils"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	users, err := storage.LoadUsers()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}

	var newUser models.User
	err = json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(newUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Please fill all fields", err.Error())
		return
	}

	newUser.ID = len(users) + 1
	newUser.CreatedAt = time.Now()
	newUser.UpdatedAt = nil
	newUser.DeletedAt = nil

	users = append(users, newUser)
	err = storage.SaveUsers(users)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to save user", err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, "User created successfully", newUser)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := storage.LoadUsers()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}

	var activeUsers []models.User
	for _, users := range users {
		if users.DeletedAt == nil {
			activeUsers = append(activeUsers, users)
		}
	}

	page, limit := utils.GetPaginationParams(r)

	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex > len(activeUsers) {
		startIndex = len(activeUsers)
	}
	if endIndex > len(activeUsers) {
		endIndex = len(activeUsers)
	}

	utils.SendSuccess(w, http.StatusOK, "Users loaded successfully", activeUsers[startIndex:endIndex])
}

func GetUsersByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	users, err := storage.LoadUsers()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}

	for i := range users {
		if users[i].ID == id && users[i].DeletedAt == nil {
			utils.SendSuccess(w, http.StatusOK, "Users loaded succesfully", users[i])
			return
		}
	}
	utils.SendError(w, http.StatusNotFound, "There are no users with this id", nil)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid id type", err.Error())
		return
	}

	var updatedUser models.UpdateUser
	err = json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(updatedUser)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Please fill all fields", err.Error())
		return
	}

	users, err := storage.LoadUsers()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to lode database", err.Error())
		return
	}

	for i := range users {
		if users[i].ID == id && users[i].DeletedAt == nil {
			if updatedUser.FirstName != "" {
				users[i].FirstName = updatedUser.FirstName
			}
			if updatedUser.LastName != "" {
				users[i].LastName = updatedUser.LastName
			}
			if updatedUser.Email != "" {
				users[i].Email = updatedUser.Email
			}
			if updatedUser.Password != "" {
				users[i].Password = updatedUser.Password
			}

			now := time.Now()
			users[i].UpdatedAt = &now

			err := storage.SaveUsers(users)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save updated user", err.Error())
				return
			}

			utils.SendSuccess(w, http.StatusOK, "User updated successfully", users[i])
			return
		}
	}

	utils.SendError(w, http.StatusBadRequest, "There are no user with this id", nil)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID type", err.Error())
		return
	}

	users, err := storage.LoadUsers()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load databse", err.Error())
		return
	}

	for i := range users {
		if users[i].ID == id && users[i].DeletedAt == nil {
			now := time.Now()
			users[i].DeletedAt = &now

			err = storage.SaveUsers(users)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failde to save to database", err.Error())
				return
			}

			utils.SendSuccess(w, http.StatusOK, "User deleted successfully", nil)
			return
		}
	}

	utils.SendError(w, http.StatusBadRequest, "There are no users with this id", nil)
}
