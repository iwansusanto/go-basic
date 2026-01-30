package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"kasir-api/models"
	"kasir-api/services"
	"kasir-api/utils"
)

type CategoryHandler struct {
	Service *services.CategoryService
}

func NewCategoryHandler(service *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{Service: service}
}

// @Router       /category/{id} [get]
func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	// Parse ID dari URL path
	// URL: /api/category/123 -> ID = 123
	idStr := strings.TrimPrefix(r.URL.Path, "/api/category/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Response{
			Status:  "failed",
			Message: "Invalid Category ID",
		})
		return
	}

	category, err := h.Service.GetByID(id)
	if err == sql.ErrNoRows {
		utils.WriteJSON(w, http.StatusNotFound, utils.Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{
			Status:  "failed",
			Message: "Failed to fetch category: " + err.Error(),
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Response{
		Status:  "success",
		Message: "Category retrieved successfully",
		Data:    category,
	})
}

// @Router       /category/{id} [delete]
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	// get id
	idStr := strings.TrimPrefix(r.URL.Path, "/api/category/")

	// ganti id int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Response{
			Status:  "failed",
			Message: "Invalid Category ID",
		})
		return
	}

	err = h.Service.Delete(id)
	if err == sql.ErrNoRows {
		utils.WriteJSON(w, http.StatusNotFound, utils.Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{
			Status:  "failed",
			Message: "Failed to delete category: " + err.Error(),
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Response{
		Status:  "success",
		Message: "Category deleted successfully",
	})
}

// @Router       /category/{id} [put]
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	// get id dari request
	idStr := strings.TrimPrefix(r.URL.Path, "/api/category/")

	// ganti int
	id, err := strconv.Atoi(idStr)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Response{
			Status:  "failed",
			Message: "Invalid Category ID",
		})
		return
	}

	// get data dari request
	var updateCategory models.Category
	err = json.NewDecoder(r.Body).Decode(&updateCategory)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Response{
			Status:  "failed",
			Message: "Invalid request body",
		})
		return
	}

	// Fetch existing category first
	existingCategory, err := h.Service.GetByID(id)
	if err == sql.ErrNoRows {
		utils.WriteJSON(w, http.StatusNotFound, utils.Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{
			Status:  "failed",
			Message: "Failed to fetch category: " + err.Error(),
		})
		return
	}

	// Merge: only update fields that are provided in request body
	if updateCategory.Name != "" {
		existingCategory.Name = updateCategory.Name
	}
	if updateCategory.Description != "" {
		existingCategory.Description = updateCategory.Description
	}

	// Update category di database
	updatedCategory, err := h.Service.Update(existingCategory)
	if err == sql.ErrNoRows {
		utils.WriteJSON(w, http.StatusNotFound, utils.Response{
			Status:  "failed",
			Message: "Category not found",
		})
		return
	}

	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{
			Status:  "failed",
			Message: "Failed to update category: " + err.Error(),
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Response{
		Status:  "success",
		Message: "Category updated successfully",
		Data:    updatedCategory,
	})
}

// @Router       /category [get]
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Service.GetAll()
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{
			Status:  "failed",
			Message: "Failed to fetch categories: " + err.Error(),
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, utils.Response{
		Status:  "success",
		Message: "Categories retrieved successfully",
		Data:    categories,
	})
}

// @Router       /category [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	// baca data dari request
	var categoryBaru models.Category
	err := json.NewDecoder(r.Body).Decode(&categoryBaru)
	if err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.Response{
			Status:  "failed",
			Message: "Invalid request body",
		})
		return
	}

	category, err := h.Service.Create(categoryBaru)
	if err != nil {
		utils.WriteJSON(w, http.StatusInternalServerError, utils.Response{
			Status:  "failed",
			Message: "Failed to save category: " + err.Error(),
		})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, utils.Response{
		Status:  "success",
		Message: "Category created successfully",
		Data:    category,
	})
}
