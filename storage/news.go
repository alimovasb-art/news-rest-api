package storage

import (
	"encoding/json"
	"errors"
	"news-restapi/models"
	"os"
)

func LoadNews() ([]models.News, error) {
	file, err := os.Open("news.json")
	if err != nil {
		return []models.News{}, err
	}

	defer file.Close()

	var news []models.News

	decode := json.NewDecoder(file)

	err = decode.Decode(&news)
	if err != nil {
		return []models.News{}, err
	}

	return news, nil
}

func SaveNews(news []models.News) error {
	file, err := os.Create("news.json")
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)

	encoder.SetIndent("", "  ")

	err = encoder.Encode(news)
	if err != nil {
		return err
	}

	return nil
}

func LoadNewsResponse() ([]models.NewsResponse, error) {
	file, err := os.Open("news.json")
	if err != nil {
		return []models.NewsResponse{}, err
	}

	defer file.Close()

	var news []models.NewsResponse

	decode := json.NewDecoder(file)

	err = decode.Decode(&news)
	if err != nil {
		return []models.NewsResponse{}, err
	}

	return news, nil
}

func SaveNewsResponse(news []models.NewsResponse) error {
	file, err := os.Create("news.json")
	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(file)

	encoder.SetIndent("", "  ")

	err = encoder.Encode(news)
	if err != nil {
		return err
	}

	return nil
}

func LoadNewsJoinedUsers() ([]models.NewsResponse, error) {
	newsList, err := LoadNews()
	if err != nil {
		return nil, err
	}

	users, err := LoadUsers()
	if err != nil {
		return nil, errors.New("failed to load users")
	}

	usersMap := make(map[int]models.Author)
	for _, u := range users {
		usersMap[u.ID] = models.Author{
			ID:        u.ID,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Email:     u.Email,
		}
	}

	var ListOfJoinedNews []models.NewsResponse

	for _, item := range newsList {

		if item.DeletedAt != nil {
			continue
		}

		author, ok := usersMap[item.AuthorID]
		if !ok {
			continue
		}

		safeAuthor := author

		response := models.NewsResponse{
			ID:               item.ID,
			Title:            item.Title,
			ShortDescription: item.ShortDescription,
			Description:      item.Description,
			Views:            item.Views,
			AuthorID:         item.AuthorID,
			Author:           &safeAuthor,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
			DeletedAt:        item.DeletedAt,
		}

		ListOfJoinedNews = append(ListOfJoinedNews, response)
	}

	return ListOfJoinedNews, nil
}
