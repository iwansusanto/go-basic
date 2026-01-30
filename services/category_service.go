package services

import (
	"kasir-api/models"
	"kasir-api/repositories"
)

type CategoryService struct {
	Repo *repositories.CategoryRepository
}

func NewCategoryService(repo *repositories.CategoryRepository) *CategoryService {
	return &CategoryService{Repo: repo}
}

func (s *CategoryService) GetAll() ([]models.Category, error) {
	return s.Repo.GetAll()
}

func (s *CategoryService) GetByID(id int) (models.Category, error) {
	return s.Repo.GetByID(id)
}

func (s *CategoryService) Create(category models.Category) (models.Category, error) {
	return s.Repo.Create(category)
}

func (s *CategoryService) Update(category models.Category) (models.Category, error) {
	return s.Repo.Update(category)
}

func (s *CategoryService) Delete(id int) error {
	return s.Repo.Delete(id)
}
