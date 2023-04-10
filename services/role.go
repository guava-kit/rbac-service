package services

import (
	"rbac-service/errors"
	"rbac-service/models"
	"rbac-service/tables"
	"rbac-service/utils"

	"github.com/gofiber/fiber/v2"
)

const (
	CreateRoleRouteName           = "role.create_role"
	CreateRolePermissionRouteName = "role.add_permission"
	GetRoleListRouteName          = "role.get_role_list"
	UpdateRoleRouteName           = "role.update_role"
	GetRoleByIdRouteName          = "role.get_role_by_id"
)

func (s *Service) NewRoleService() Servicer {
	s.App.Route("/role", func(router fiber.Router) {
		router.Use(s.RequiredSignin)
		router.Post("/create", s.RequiredPermission, s.CreateRole).Name(CreateRoleRouteName)
		router.Post("/add_permission", s.RequiredPermission, s.CreateRolePermission).Name(CreateRolePermissionRouteName)
		router.Get("/list", s.RequiredPermission, s.GetRoleList).Name(GetRoleListRouteName)
		router.Put("/:id", s.RequiredPermission, s.UpdateRole).Name(UpdateRoleRouteName)
		router.Get("/:id", s.RequiredPermission, s.GetRoleById).Name(GetRoleByIdRouteName)
	})
	return s
}

// CreateRole 创建角色
// 只能创建当前租户下的角色
func (s *Service) CreateRole(c *fiber.Ctx) error {
	form := &models.CreateRoleForm{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	roleExists, err := s.Dao.GetRoleByName(form.Name, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if roleExists != nil {
		return errors.ParameterError(c, errors.Message("name", "角色已存在"))
	}
	role := &tables.Role{Name: form.Name, TenantId: tenantId}
	role.Init()
	if err := s.Dao.CreateRole(role); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// CreateRolePermission 给角色分配权限
// 可以分配当前租户的子级的权限
func (s *Service) CreateRolePermission(c *fiber.Ctx) error {
	form := &models.CreateRolePermission{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	roleExists, err := s.Dao.GetRoleById(form.RoleId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if roleExists == nil {
		return errors.ParameterError(c, errors.Message("role_id", "角色不存在"))
	}
	permissionExists, err := s.Dao.GetPermissionOnTenantTreeById(form.PermissionId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if permissionExists == nil {
		return errors.ParameterError(c, errors.Message("permission_id", "权限不存在"))
	}
	rolePermissionExists, err := s.Dao.GetRolePermissionByUnique(form.RoleId, form.PermissionId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if rolePermissionExists != nil {
		return errors.ParameterError(c, errors.Message("permission_id", "角色已有此权限"))
	}
	rolePermission := &tables.RolePermission{RoleId: form.RoleId, PermissionId: form.PermissionId, TenantId: tenantId}
	rolePermission.Init()
	if err := s.Dao.CreateRolePermission(rolePermission); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// UpdateRole 更新角色
func (s *Service) UpdateRole(c *fiber.Ctx) error {
	roleId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	roleExists, err := s.Dao.GetRoleById(roleId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if roleExists == nil {
		return errors.ParameterError(c, "角色不存在")
	}
	form := models.UpdateRoleForm{}
	if err := c.BodyParser(&form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	updateMap := map[string]any{}
	if name, ok := form["name"]; ok && name != roleExists.Name {
		roleExists, err := s.Dao.GetRoleByName(name.(string), tenantId)
		if err != nil {
			return errors.DatabaseError(c)
		}
		if roleExists != nil {
			return errors.ParameterError(c, errors.Message("name", "角色名称已存在"))
		}
		updateMap["name"] = name
	}
	if err := s.Dao.UpdateRole(roleExists, updateMap); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

func (s *Service) GetRoleById(c *fiber.Ctx) error {
	roleId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	roleExists, err := s.Dao.GetRoleById(roleId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if roleExists == nil {
		return errors.ParameterError(c, "角色不存在")
	}
	return errors.SucceededWithData(c, map[string]any{
		"id":   roleExists.Id,
		"name": roleExists.Name,
	})
}

func (s *Service) GetRoleList(c *fiber.Ctx) error {
	query := &models.RoleQuery{}
	if err := c.QueryParser(query); err != nil {
		return errors.QueryParserError(c, err.Error())
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	offset, limit := utils.GetOffsetLimit(query.Page, query.Limit)
	roleList, err := s.Dao.GetRoleList(
		tenantId,
		offset,
		limit,
	)
	if err != nil {
		return errors.DatabaseError(c)
	}
	count, err := s.Dao.GetRoleCount(tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	return errors.SucceededWithData(c, map[string]any{
		"list":  new(models.RoleList).FormTable(roleList),
		"total": count,
	})
}
