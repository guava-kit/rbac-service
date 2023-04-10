package services

import (
	"rbac-service/errors"
	"rbac-service/models"
	"rbac-service/tables"
	"rbac-service/utils"

	"github.com/gofiber/fiber/v2"
)

const (
	CreateUserRouteName       = "user.create_user"
	CreateUserRoleRouteName   = "user.add_role"
	CreateUserTenantRouteName = "user.add_tenant"
	UpdateUserRouteName       = "user.update_user"
	GetUserListRouteName      = "user.get_user_list"
	GetUserByIdRouteName      = "user.get_user_by_id"
)

func (s *Service) NewUserService() Servicer {
	s.App.Route("/user", func(router fiber.Router) {
		router.Use(s.RequiredSignin)
		router.Post("/create", s.RequiredPermission, s.CreateUser).Name(CreateUserRouteName)
		router.Post("/add_role", s.RequiredPermission, s.CreateUserRole).Name(CreateUserRoleRouteName)
		router.Post("/add_tenant", s.RequiredPermission, s.CreateUserTenant).Name(CreateUserTenantRouteName)
		router.Put("/:id", s.RequiredPermission, s.UpdateUser).Name(UpdateUserRouteName)
		router.Get("/list", s.RequiredPermission, s.GetUserList).Name(GetUserListRouteName)
		router.Get("/:id", s.RequiredPermission, s.GetUserById).Name(GetUserByIdRouteName)
	})
	return s
}

// CreateUser 创建用户
// 手动创建用户，可以指定所在租户
func (s *Service) CreateUser(c *fiber.Ctx) error {
	form := &models.CreateUserForm{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	tenantExists, err := s.Dao.GetTenantTree(tenantId, form.TenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if tenantExists == nil {
		return errors.ParameterError(c, errors.Message("tenant_id", "租户不存在"))
	}
	userExists, err := s.Dao.GetUserByUsername(form.Username)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userExists != nil {
		return errors.ParameterError(c, errors.Message("username", "用户名已存在"))
	}
	user := &tables.User{Username: form.Username, Password: utils.GeneratePasswordHash(form.Password)}
	user.Init()
	userTenant := &tables.UserTenant{UserId: user.Id, TenantId: tenantId}
	userTenant.Init()
	if err := s.Dao.CreateUserAndJoinTenant(user, userTenant); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// CreateUserRole 给用户分配角色
// 用户只能分配当前租户下的角色
func (s *Service) CreateUserRole(c *fiber.Ctx) error {
	form := &models.CreateUserRoleForm{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	userTenantEists, err := s.Dao.GetUserTenantByUnique(form.UserId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userTenantEists == nil {
		return errors.ParameterError(c, errors.Message("user_id", "用户不存在"))
	}
	roleExists, err := s.Dao.GetRoleById(form.RoleId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if roleExists == nil {
		return errors.ParameterError(c, errors.Message("role_id", "角色不存在"))
	}
	userRoleExists, err := s.Dao.GetUserRoleByUnique(form.UserId, form.RoleId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userRoleExists != nil {
		return errors.ParameterError(c, errors.Message("role_id", "用户已有此角色"))
	}
	userRole := &tables.UserRole{UserId: form.UserId, RoleId: form.RoleId, TenantId: tenantId}
	userRole.Init()
	if err := s.Dao.CreateUserRole(userRole); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// CreateUserTenant 把用户分配给租户
// 可以把用户加入到不同的租户，加入的租户必须是当前租户的子级
func (s *Service) CreateUserTenant(c *fiber.Ctx) error {
	form := &models.CreateUserTenantForm{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	tenantTreeExists, err := s.Dao.GetTenantTree(tenantId, form.TenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if tenantTreeExists == nil {
		return errors.ParameterError(c, errors.Message("tenant_id", "租户不存在"))
	}
	userTenantEists, err := s.Dao.GetUserOnTenantTreeById(form.UserId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userTenantEists == nil {
		return errors.ParameterError(c, errors.Message("user_id", "用户不存在"))
	}
	userTenantExists, err := s.Dao.GetUserTenantByUnique(form.UserId, form.TenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userTenantExists != nil {
		return errors.ParameterError(c, errors.Message("tenant_id", "租户已有此用户"))
	}
	userTenant := &tables.UserTenant{UserId: form.UserId, TenantId: form.TenantId}
	userTenant.Init()
	if err := s.Dao.CreateUserTenant(userTenant); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// UpdateUser 更新用户信息
func (s *Service) UpdateUser(c *fiber.Ctx) error {
	userId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	userExists, err := s.Dao.GetUserOnTenantTreeById(userId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userExists == nil {
		return errors.ParameterError(c, "用户不存在")
	}
	form := models.UpdateUserForm{}
	if err := c.BodyParser(&form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	updateMap := map[string]any{}
	if username, ok := form["username"]; ok && username != userExists.Username {
		userExists, err := s.Dao.GetUserByUsername(username.(string))
		if err != nil {
			return errors.DatabaseError(c)
		}
		if userExists != nil {
			return errors.ParameterError(c, errors.Message("username", "用户名已存在"))
		}
		updateMap["username"] = username
	}
	if password, ok := form["password"]; ok && !utils.ComparePasswordHash(password.(string), userExists.Username) {
		updateMap["password"] = utils.GeneratePasswordHash(password.(string))
	}
	if err := s.Dao.UpdateUser(userExists, updateMap); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// GetUserById 通过id获取单个用户信息
func (s *Service) GetUserById(c *fiber.Ctx) error {
	userId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	userExists, err := s.Dao.GetUserOnTenantById(userId, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userExists == nil {
		return errors.ParameterError(c, "用户不存在")
	}
	return errors.SucceededWithData(c, map[string]any{
		"id":       userExists.Id,
		"username": userExists.Username,
	})
}

// GetUserList 通过不同条件获取用户列表
func (s *Service) GetUserList(c *fiber.Ctx) error {
	query := &models.UserQuery{}
	if err := c.QueryParser(query); err != nil {
		return errors.QueryParserError(c, err.Error())
	}
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	offset, limit := utils.GetOffsetLimit(query.Page, query.Limit)
	userList, err := s.Dao.GetUserListByTenant(tenantId, offset, limit)
	if err != nil {
		return errors.DatabaseError(c)
	}
	count, err := s.Dao.GetUserCountByTenant(tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	return errors.SucceededWithData(c, map[string]any{
		"list":  new(models.UserList).FormTable(userList),
		"total": count,
	})
}
