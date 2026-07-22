package storage

import (
	"encoding/json"
	"errors"
	"news-restapi/models"
	"os"
)

func LoadUsers() ([]models.User, error) {
	file, err := os.Open("users.json")
	if err != nil {
		return []models.User{}, err
	}

	defer file.Close()

	var users []models.User

	decode := json.NewDecoder(file)

	err = decode.Decode(&users)
	if err != nil {
		return []models.User{}, err
	}

	return users, nil
}

func SaveUsers(users []models.User) error {
	file, err := os.Create("users.json")
	if err != nil {
		return err
	}

	defer file.Close()

	encode := json.NewEncoder(file)
	encode.SetIndent("", "  ")

	err = encode.Encode(users)
	if err != nil {
		return err
	}

	return nil
}

func GetUserByID(id int) (*models.User, error) {
	users, err := LoadUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.ID == id && user.DeletedAt == nil {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}
