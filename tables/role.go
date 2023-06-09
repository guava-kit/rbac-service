package tables

type Role struct {
	Base
	Name     string `gorm:"type:varchar(32);not null;"`
	TenantId string `gorm:"type:varchar(36);not null;"`
	Tenant   Tenant `gorm:"foreignKey:Id;references:TenantId;constraint:OnDelete:CASCADE;"`
}

type RolePermission struct {
	Base
	RoleId       string      `gorm:"type:varchar(36);not null;uniqueIndex:idx_role_permission;"`
	PermissionId string      `gorm:"type:varchar(36);not null;uniqueIndex:idx_role_permission;"`
	TenantId     string      `gorm:"type:varchar(36);not null;uniqueIndex:idx_role_permission;"`
	Role         *Role       `gorm:"foreignKey:Id;references:RoleId;constraint:OnDelete:CASCADE;"`
	Permission   *Permission `gorm:"foreignKey:Id;references:PermissionId;constraint:OnDelete:CASCADE;"`
	Tenant       Tenant      `gorm:"foreignKey:Id;references:TenantId;constraint:OnDelete:CASCADE;"`
}

func (Role) TableName() string {
	return "rbac_role"
}

func (RolePermission) TableName() string {
	return "rbac_role_permission"
}
