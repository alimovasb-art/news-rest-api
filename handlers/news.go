package handlers

import (
	"encoding/json"
	"net/http"
	"news-restapi/models"
	"news-restapi/storage"
	"news-restapi/utils"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

func CreateNews(w http.ResponseWriter, r *http.Request) {
	authorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	var newNews models.News
	err := json.NewDecoder(r.Body).Decode(&newNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	newNews.AuthorID = authorID

	news, err := storage.LoadNews()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(newNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
		return
	}

	newNews.ID = len(news) + 1
	newNews.CreatedAt = time.Now()
	newNews.UpdatedAt = nil
	newNews.Views = 0

	news = append(news, newNews)
	err = storage.SaveNews(news)
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to save news", err.Error())
		return
	}

	utils.SendSuccess(w, http.StatusCreated, "News created successfully", newNews)
}

func GetNews(w http.ResponseWriter, r *http.Request) {
	authorIdStr := r.URL.Query().Get("author_id")
	title := r.URL.Query().Get("title")

	if title != "" {
		news, err := storage.LoadNewsJoinedUsers()
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "Invalid title format", err.Error())
			return
		}

		query := strings.ToLower(title)

		var result []models.NewsResponse

		for i := range news {
			if strings.Contains(strings.ToLower(news[i].Title), query) {
				result = append(result, news[i])
			}
		}

		if len(result) == 0 {
			utils.SendError(w, http.StatusNotFound, "No news found for this title", nil)
			return
		}

		utils.SendSuccess(w, http.StatusOK, "News by author found", result)
		return
	}
	//Get news by authors
	if authorIdStr != "" {
		authorID, err := strconv.Atoi(authorIdStr)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
			return
		}

		news, err := storage.LoadNewsJoinedUsers()
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
			return
		}

		var newsByAuthors []models.NewsResponse

		for i := range news {
			if news[i].AuthorID == authorID && news[i].DeletedAt == nil {
				newsByAuthors = append(newsByAuthors, news[i])
			}
		}

		if len(newsByAuthors) == 0 {
			utils.SendError(w, http.StatusNotFound, "No news found for this author", nil)
			return
		}

		utils.SendSuccess(w, http.StatusOK, "News by author found", newsByAuthors)
		return
	}

	news, err := storage.LoadNewsResponse()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}
	var activeNews []models.NewsResponse
	for _, news := range news {
		if news.DeletedAt == nil {
			activeNews = append(activeNews, news)
		}
	}

	page, limit := utils.GetPaginationParams(r)

	startIndex := (page - 1) * limit
	endIndex := startIndex + limit

	if startIndex > len(activeNews) {
		startIndex = len(activeNews)
	}
	if endIndex > len(activeNews) {
		endIndex = len(activeNews)
	}

	paginatedNews := activeNews[startIndex:endIndex]

	for i := range paginatedNews {
		user, err := storage.GetUserByID(paginatedNews[i].AuthorID)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "Author not found", nil)
			return
		}
		paginatedNews[i].Author = &models.Author{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		}
	}

	utils.SendSuccess(w, http.StatusOK, "News list retrived", paginatedNews)
}

func GetNewsByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	news, err := storage.LoadNewsResponse()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}
	for i := range news {
		if news[i].ID == id {
			if news[i].DeletedAt != nil {
				utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
				return
			}

			news[i].Views++
			err = storage.SaveNewsResponse(news)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save news", err.Error())
				return
			}

			user, err := storage.GetUserByID(news[i].AuthorID)
			if err != nil {
				utils.SendError(w, http.StatusNotFound, "Failed to find user", err.Error())
				return
			}

			saveAuthor := models.Author{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
			}

			news[i].Author = &saveAuthor

			utils.SendSuccess(w, http.StatusOK, "News fetched successfully", news[i])
			return
		}
	}
	utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
}

func UpdateNews(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	authorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	var updateNews models.News
	err = json.NewDecoder(r.Body).Decode(&updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
		return
	}

	news, err := storage.LoadNews()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load news", err.Error())
		return
	}

	for i := range news {
		if news[i].ID == id {
			if news[i].DeletedAt != nil {
				utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
				return
			}

			if authorID != news[i].AuthorID {
				utils.SendError(w, http.StatusBadRequest, "Only author can update news, you are not the author of this news!", nil)
				return
			}

			news[i].Title = updateNews.Title
			news[i].Description = updateNews.Description
			news[i].ShortDescription = updateNews.ShortDescription

			now := time.Now()
			news[i].UpdatedAt = &now

			err = storage.SaveNews(news)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save news", err.Error())
				return
			}

			utils.SendSuccess(w, http.StatusOK, "News updated successfully", news[i])
			return
		}
	}
	utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
}

func PatchNews(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	authorID, ok := r.Context().Value("author_id").(int)
	if !ok {
		utils.SendError(w, http.StatusInternalServerError, "Failed to get user from context", nil)
		return
	}

	var updateNews models.UpdateNews
	err = json.NewDecoder(r.Body).Decode(&updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	validate := validator.New()
	err = validate.Struct(updateNews)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "You must fill title(from 3 to 20 symbols), description(from 10 to 40 symbols) and short_description(from 20)", err.Error())
		return
	}

	news, err := storage.LoadNews()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load news", err.Error())
		return
	}

	for i := range news {
		if news[i].ID == id {
			if news[i].DeletedAt != nil {
				utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
				return
			}

			if authorID != news[i].AuthorID {
				utils.SendError(w, http.StatusBadRequest, "Only author can update news, you are not the author of this news!", nil)
				return
			}

			if updateNews.Title != "" {
				news[i].Title = updateNews.Title
			}
			if updateNews.ShortDescription != "" {
				news[i].ShortDescription = updateNews.ShortDescription
			}
			if updateNews.Description != "" {
				news[i].Description = updateNews.Description
			}

			now := time.Now()
			news[i].UpdatedAt = &now

			err = storage.SaveNews(news)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save news", err.Error())
				return
			}

			utils.SendSuccess(w, http.StatusOK, "News updated successfully", news[i])
			return
		}
	}
	utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
}

func DeleteNews(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.SendError(w, http.StatusBadRequest, "Invalid ID format", err.Error())
		return
	}

	news, err := storage.LoadNews()
	if err != nil {
		utils.SendError(w, http.StatusInternalServerError, "Failed to load database", err.Error())
		return
	}

	for i := range news {
		if news[i].ID == id {
			if news[i].DeletedAt != nil {
				utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
				return
			}

			now := time.Now()
			news[i].DeletedAt = &now

			err = storage.SaveNews(news)
			if err != nil {
				utils.SendError(w, http.StatusInternalServerError, "Failed to save news", err.Error())
				return
			}

			utils.SendSuccess(w, http.StatusOK, "News deleted successfully", nil)
			return
		}
	}
	utils.SendError(w, http.StatusNotFound, "There are no news with this id", nil)
}
