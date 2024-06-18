package service

import (
	"context"
	"xo-packs/model"
	"xo-packs/repository"
)

type CategoryService interface {
	GetAllCategories(context.Context) ([]*model.Category, error)
	GetCategories(context.Context, string) ([]*model.Category, error)
	GetCategoryLiterals(context.Context) ([]string, error)
}

type CategorySvcImpl struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &CategorySvcImpl{categoryRepo: repo}
}

func (categoryService CategorySvcImpl) GetAllCategories(c context.Context) ([]*model.Category, error) {
	return categoryService.categoryRepo.GetAllCategories(c)
}

func (categoryService CategorySvcImpl) GetCategories(c context.Context, urlPath string) ([]*model.Category, error) {
	return categoryService.categoryRepo.GetCategories(c, urlPath)
}

func (categoryService CategorySvcImpl) GetCategoryLiterals(c context.Context) ([]string, error) {
	return categoryService.categoryRepo.GetCategoryLiterals(c)
}
