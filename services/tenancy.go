package services

import (
	"rbac-service/errors"
	"rbac-service/models"
	"rbac-service/tables"
	"rbac-service/utils"

	"github.com/gofiber/fiber/v2"
)

const (
	CreateTenantRouteName  = "tenant.create_tenant"
	GetTenantListRouteName = "tenant.get_tenant_list"
	UpdateTenantRouteName  = "tenant.update_tenant"
	GetTenantByIdRouteName = "tenant.get_tenant_by_id"
)

func (s *Service) NewTenantService() Servicer {
	s.App.Route("/tenant", func(router fiber.Router) {
		router.Use(s.RequiredSignin)
		router.Post("/create", s.RequiredPermission, s.CreateTenant).Name(CreateTenantRouteName)
		router.Get("/list", s.RequiredPermission, s.GetTenantList).Name(GetTenantListRouteName)
		router.Put("/:id", s.RequiredPermission, s.UpdateTenant).Name(UpdateTenantRouteName)
		router.Get("/:id", s.RequiredPermission, s.GetTenantById).Name(GetTenantByIdRouteName)
	})
	return s
}

// CreateTenant 创建租户
// 判断创建的租户是否有父级id，如果有需要判断父级id是否属于上下文tenantId的子级
// 判断租户名称是否已经存在
func (s *Service) CreateTenant(c *fiber.Ctx) error {
	form := &models.CreateTenantForm{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	user := c.UserContext().Value(ContextKey("user")).(*tables.User)
	if form.ParentId != nil {
		tenantId, ok := c.UserContext().Value(ContextKey("tenantId")).(string)
		if !ok {
			return errors.ParameterError(c, "请求头需要携带tenant")
		}
		tenantTreeExists, err := s.Dao.GetTenantTree(tenantId, *form.ParentId)
		if err != nil {
			return errors.DatabaseError(c)
		}
		if tenantTreeExists == nil {
			return errors.ParameterError(c, errors.Message("parent_id", "父级租户不存在"))
		}
	}
	tenantExists, err := s.Dao.GetTenantByNameWithParent(form.Name, form.ParentId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if tenantExists != nil {
		return errors.ParameterError(c, errors.Message("name", "租户名称已存在"))
	}
	tenant := &tables.Tenant{Name: form.Name, ParentId: form.ParentId, Owner: user.Id}
	tenant.Init()
	if err := s.Dao.CreateTenant(tenant); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// UpdateTenant 更新租户
func (s *Service) UpdateTenant(c *fiber.Ctx) error {
	updateTenantId := c.Params("id")
	tenantId := c.UserContext().Value(ContextKey("tenantId")).(string)
	originTenant, err := s.Dao.GetTenantJoinTenantTreeOnDescendant(tenantId, updateTenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if originTenant == nil {
		return errors.ParameterError(c, "租户不存在")
	}
	form := models.UpdateTenantForm{}
	if err := c.BodyParser(&form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	updateMap := map[string]any{}
	if err := s.CompareName(c, form, originTenant, updateMap); err != nil {
		return err
	}
	if err := s.CompareParentId(c, form, tenantId, originTenant, updateMap); err != nil {
		return err
	}
	if len(updateMap) > 0 {
		if err := s.Dao.UpdateTenant(originTenant, updateMap); err != nil {
			return errors.DatabaseError(c)
		}
	}
	return errors.Succeeded(c)
}

func (s *Service) CompareName(c *fiber.Ctx, form models.UpdateTenantForm, originTenant *tables.Tenant, updateMap models.UpdateMap) error {
	if name, ok := form["name"]; ok && name != originTenant.Name {
		tenantExists, err := s.Dao.GetTenantByNameWithParent(name.(string), originTenant.ParentId)
		if err != nil {
			return errors.DatabaseError(c)
		}
		if tenantExists != nil {
			return errors.ParameterError(c, errors.Message("name", "租户名称已存在"))
		}
		updateMap["name"] = name
	}
	return nil
}

func (s *Service) CompareParentId(c *fiber.Ctx, form models.UpdateTenantForm, contextTenantId string, originTenant *tables.Tenant, updateMap models.UpdateMap) error {
	if parentId, ok := form["parent_id"]; ok && parentId != originTenant.ParentId {
		if parentId != nil {
			tenantTreeExists, err := s.Dao.GetTenantTree(contextTenantId, parentId.(string))
			if err != nil {
				return errors.DatabaseError(c)
			}
			if tenantTreeExists == nil {
				return errors.ParameterError(c, errors.Message("parent_id", "父级租户不存在"))
			}
		}
		var tenantExists *tables.Tenant
		var err error
		parentIdString, isString := parentId.(string)
		if name, ok := form["name"]; ok && name != originTenant.Name {
			if isString {
				tenantExists, err = s.Dao.GetTenantByNameWithParent(name.(string), &parentIdString)
			} else {
				tenantExists, err = s.Dao.GetTenantByNameWithParent(name.(string), nil)
			}
		} else {
			if isString {
				tenantExists, err = s.Dao.GetTenantByNameWithParent(originTenant.Name, &parentIdString)
			} else {
				tenantExists, err = s.Dao.GetTenantByNameWithParent(originTenant.Name, nil)
			}
		}
		if err != nil {
			return errors.DatabaseError(c)
		}
		if tenantExists != nil {
			return errors.ParameterError(c, errors.Message("name", "租户名称已存在"))
		}
		updateMap["parent_id"] = parentId
	}
	return nil
}

// GetTenantById 通过id获取租户
func (s *Service) GetTenantById(c *fiber.Ctx) error {
	tenantId := c.Params("id")
	user := c.UserContext().Value(ContextKey("user")).(*tables.User)
	tenantExists, err := s.Dao.GetTenantByIdWithOwner(tenantId, user.Id)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if tenantExists == nil {
		return errors.ParameterError(c, "租户不存在")
	}
	return errors.SucceededWithData(c, map[string]any{
		"id":   tenantExists.Id,
		"name": tenantExists.Name,
	})
}

// GetTenantList 获取租户列表
func (s *Service) GetTenantList(c *fiber.Ctx) error {
	query := &models.TenantQuery{}
	if err := c.QueryParser(query); err != nil {
		return errors.QueryParserError(c, err.Error())
	}
	user := c.UserContext().Value(ContextKey("user")).(*tables.User)
	offset, limit := utils.GetOffsetLimit(query.Page, query.Limit)
	tenantList, err := s.Dao.GetTenantListByOwner(
		user.Id,
		offset,
		limit,
	)
	if err != nil {
		return errors.DatabaseError(c)
	}
	count, err := s.Dao.GetTenantCountByOwner(user.Id)
	if err != nil {
		return errors.DatabaseError(c)
	}
	return errors.SucceededWithData(c, map[string]any{
		"list":  new(models.TenantList).FormTable(tenantList),
		"total": count,
	})
}
