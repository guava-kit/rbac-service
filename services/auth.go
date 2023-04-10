package services

import (
	"context"
	"rbac-service/errors"
	"rbac-service/models"
	"rbac-service/tables"
	"rbac-service/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	SignupRouteName        = "auth.signup"
	SigninRouteName        = "auth.signin"
	ResetPasswordRouteName = "auth.resetPassword"
)

func (s *Service) NewAuthService() Servicer {
	s.App.Route("/auth", func(router fiber.Router) {
		router.Post("/signup", s.Signup).Name(SignupRouteName)
		router.Post("/signin", s.Signin).Name(SigninRouteName)
		router.Use(s.RequiredSignin).Post("/reset_password", s.ResetPassword).Name(ResetPasswordRouteName)
	})
	return s
}

// Signup 用户注册，可使用邮箱或者手机号码注册
func (s *Service) Signup(c *fiber.Ctx) error {
	signupType := c.Query("type")
	switch signupType {
	case "email":
		return s.SignupWithEmail(c)
	case "phone":
		return s.SignupWithPhone(c)
	}
	return errors.ParameterError(c, "注册类型错误")
}

// SignupWithEmail 使用邮箱注册
func (s *Service) SignupWithEmail(c *fiber.Ctx) error {
	form := &models.SignupWithEmail{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	usernameExists, err := s.Dao.GetUserByUsername(form.Username)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if usernameExists != nil {
		return errors.ParameterError(c, errors.Message("username", "用户名已存在"))
	}
	emailExists, err := s.Dao.GetUserByEmail(form.Email)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if emailExists != nil {
		return errors.ParameterError(c, errors.Message("email", "邮箱已注册"))
	}
	user := &tables.User{
		Username: form.Username,
		Email:    form.Email,
		Password: utils.GeneratePasswordHash(form.Password),
	}
	user.Init()
	if err := s.Dao.CreateUser(user); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// SignupWithPhone 使用手机号码注册
func (s *Service) SignupWithPhone(c *fiber.Ctx) error {
	form := &models.SignupWithPhone{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	usernameExists, err := s.Dao.GetUserByUsername(form.Username)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if usernameExists != nil {
		return errors.ParameterError(c, errors.Message("username", "用户名已存在"))
	}
	phoneExists, err := s.Dao.GetUserByPhone(form.Phone)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if phoneExists != nil {
		return errors.ParameterError(c, errors.Message("phone", "手机号码已注册"))
	}
	user := &tables.User{
		Username: form.Username,
		Phone:    form.Phone,
		Password: utils.GeneratePasswordHash(form.Password),
	}
	user.Init()
	if err := s.Dao.CreateUser(user); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// Signin 用户登录，可使用用户名，邮箱或者手机号码登录
func (s *Service) Signin(c *fiber.Ctx) error {
	signinType := c.Query("type")
	switch signinType {
	case "username":
		return s.SigninWithUsername(c)
	case "email":
		return s.SigninWithEmail(c)
	case "phone":
		return s.SigninWithPhone(c)
	}
	return errors.ParameterError(c, "登录类型错误")
}

// SigninWithUsername 使用用户名登录
func (s *Service) SigninWithUsername(c *fiber.Ctx) error {
	form := &models.SigninWithUsername{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	user, err := s.Dao.GetUserByUsername(form.Username)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if user == nil {
		return errors.ParameterError(c, errors.Message("username", "用户名不存在"))
	}
	if !utils.ComparePasswordHash(form.Password, user.Password) {
		return errors.ParameterError(c, errors.Message("password", "密码错误"))
	}
	return s.IssueToken(c, user)
}

// SigninWithEmail 使用邮箱登录
func (s *Service) SigninWithEmail(c *fiber.Ctx) error {
	form := &models.SigninWithEmail{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	user, err := s.Dao.GetUserByEmail(form.Email)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if user == nil {
		return errors.ParameterError(c, errors.Message("email", "邮箱不存在"))
	}
	if !utils.ComparePasswordHash(form.Password, user.Password) {
		return errors.ParameterError(c, errors.Message("password", "密码错误"))
	}
	return s.IssueToken(c, user)
}

// SigninWithPhone 使用电话号码登录
func (s *Service) SigninWithPhone(c *fiber.Ctx) error {
	//Todo
	return errors.SucceededWithData(c, &models.Token{})
}

// IssueToken 发布jwt
func (s *Service) IssueToken(c *fiber.Ctx, user *tables.User) error {
	payload := map[string]any{"user_id": user.Id}
	token, err := utils.GenerateJwtWithKey(
		payload,
		s.Config.JWT_SECRET_KEY,
		time.Duration(s.Config.JWT_EXPIRATION)*time.Millisecond,
	)
	if err != nil {
		return errors.UnknownError(c, err.Error())
	}
	return errors.SucceededWithData(c, &models.Token{Token: token})
}

// ResetPassword 可在已登录的状态下修改密码
func (s *Service) ResetPassword(c *fiber.Ctx) error {
	form := &models.ResetPassword{}
	if err := c.BodyParser(form); err != nil {
		return errors.BodyParserError(c, err.Error())
	}
	if err := form.Validate(); err != nil {
		return errors.ParameterError(c, err)
	}
	user := c.UserContext().Value(ContextKey("user")).(*tables.User)
	if !utils.ComparePasswordHash(form.OldPassword, user.Password) {
		return errors.ParameterError(c, errors.Message("password", "密码错误"))
	}
	newPassword := utils.GeneratePasswordHash(form.NewPassword)
	if err := s.Dao.UpdateUser(user, map[string]any{"password": newPassword}); err != nil {
		return errors.DatabaseError(c)
	}
	return errors.Succeeded(c)
}

// RequiredSignin 校验是否登录
func (s *Service) RequiredSignin(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	jwtString, ok := headers["Authorization"]
	if !ok {
		return errors.UnauthorizedError(c)
	}
	token, err := utils.ParseJwtWithKey(jwtString, s.Config.JWT_SECRET_KEY)
	if err == jwt.ErrTokenExpired() {
		return errors.UnauthorizedError(c, "令牌过期")
	}
	if err != nil {
		return errors.UnauthorizedError(c, "无效令牌")
	}
	userId, ok := token.Get("user_id")
	if !ok {
		return errors.UnauthorizedError(c, "无效令牌")
	}
	user, err := s.Dao.GetUserById(userId.(string))
	if err != nil {
		return errors.DatabaseError(c)
	}
	if user == nil {
		return errors.UnauthorizedError(c, "用户不存在")
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, ContextKey("user"), user)
	c.SetUserContext(ctx)
	return c.Next()
}

// RequiredPermission 检查用户是否有权限
// owner是租户的拥有者，拥有所有权限
func (s *Service) RequiredPermission(c *fiber.Ctx) error {
	headers := c.GetReqHeaders()
	tenantId, ok := headers["Tenant"]
	if !ok {
		if s.Config.ALLOW_USER_CREATE_TENANT && c.Route().Name == CreateTenantRouteName {
			return c.Next()
		}
		return errors.ForbiddenError(c)
	}
	user := c.UserContext().Value(ContextKey("user")).(*tables.User)
	tenant, err := s.Dao.GetTenantByIdWithOwner(tenantId, user.Id)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if tenant != nil {
		ctx := c.UserContext()
		ctx = context.WithValue(ctx, ContextKey("tenantId"), tenant.Id)
		c.SetUserContext(ctx)
		return c.Next()
	}

	userTenant, err := s.Dao.GetUserTenantByUnique(user.Id, tenantId)
	if err != nil {
		return errors.DatabaseError(c)
	}
	if userTenant == nil {
		return errors.ForbiddenError(c)
	}
	//TODO: 判断接口权限
	return c.Next()
}
