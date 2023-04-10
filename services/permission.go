package services

import (
	"rbac-service/errors"
	"rbac-service/models"
	"rbac-service/tables"
	"rbac-service/utils"

	"github.com/gofiber/fiber/v2"
)

const (
	CreatePermissionRouteName  = "permission.create_permission"
	GetPermissionListRouteName = "permission.get_permission_list"
	UpdatePermissionRouteName  = "permission.update_permission"
	GetPermissionByIdRouteName = "permission.get_permission_by_id"
)

func (s *Service) NewPermissionService() Servicer {
	s.App.Route("/permission", func(router fiber.Router) {
		router.Use(s.RequiredSignin)
		router.Post("/create", s.RequiredPermission, s.CreatePermission).Name(CreatePermissionRouteName)
		router.Get("/list", s.RequiredPermission, s.GetPermissionList).Name(GetPermissionListRouteName)
		router.Put("/:id", s.RequiredPermission, s.UpdatePermission).Name(UpdatePermissionRouteName)
		router.Get("/:id", s.RequiredPermission, s.GetPermissionById).Name(GetPermissionByIdRouteName)
	})
	return s
}

// CreatePermission 创建权限
// 只能创建当前租户下的权限
func (s *Service) CreatePermission(c *fiber.Ctx) error {
	form := &models.CreatePermissionForm{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	permissionExists, err := s.Dao.GetPermissionByName(form.Name, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if permissionExists != nil {
		return errors.ParameterError(c, errors.Message("name", "权限名称已存在"))
	}
	permissionExists, err = s.Dao.GetPermissionByCode(form.Code, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if permissionExists != nil {
		return errors.ParameterError(c, errors.Message("code", "权限码已存在"))
	}
	permission := &tables.Permission{Name: form.Name, Code: form.Code, Description: form.Description, TenantId: tenantId}
	permission.Init()
	if err := s.Dao.CreatePermission(permission); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

func (s *Service) UpdatePermission(c *fiber.Ctx) error {
	permissionId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	permissionExists, err := s.Dao.GetPermissionById(permissionId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if permissionExists == nil {
		return errors.ParameterError(c, "权限不存在")
	}
	form := models.UpdatePermissionForm{}
	if err := c.BodyParser(&form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	updateMap := map[string]any{}
	if name, ok := form["name"]; ok && name != permissionExists.Name {
		permissionExists, err := s.Dao.GetPermissionByName(name.(string), tenantId)
		if err != nil {
			return errors.DatabaseError(c)
		}
		if permissionExists != nil {
			return errors.ParameterError(c, errors.Message("name", "权限名称已存在"))
		}
		updateMap["name"] = name
	}
	if code, ok := form["code"]; ok && code != permissionExists.Code {
		permissionExists, err := s.Dao.GetPermissionByCode(code.(string), tenantId)
		if err != nil {
			return errors.DatabaseError(c)
		}
		if permissionExists != nil {
			return errors.ParameterError(c, errors.Message("code", "权限码已存在"))
		}
		updateMap["code"] = code
	}
	if description, ok := form["description"]; ok && description != permissionExists.Description {
		updateMap["description"] = description
	}
	if err := s.Dao.UpdatePermission(permissionExists, updateMap); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

func (s *Service) GetPermissionById(c *fiber.Ctx) error {
	permissionId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	permissionExists, err := s.Dao.GetPermissionById(permissionId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if permissionExists == nil {
		return errors.ParameterError(c, "权限不存在")
	}
	return errors.SucceededWithData(c, map[string]any{
		"id":   permissionExists.Id,
		"name": permissionExists.Name,
		"code": permissionExists.Code,
	})
}

func (s *Service) GetPermissionList(c *fiber.Ctx) error {
	query := &models.PermissionQuery{}
	if err := c.QueryParser(query); err != nil {
		return errors.QueryParserError(c, err.Error())
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	offset, limit := utils.GetOffsetLimit(query.Page, query.Limit)
	permissionList, err := s.Dao.GetPermissionList(
		tenantId,
		offset,
		limit,
	)
	if err != nil {
		return errors.DatabaseError(c)
	}
	count, err := s.Dao.GetPermissionCount(tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	return errors.SucceededWithData(c, map[string]any{
		"list":  new(models.PermissionList).FormTable(permissionList),
		"total": count,
	})
}
