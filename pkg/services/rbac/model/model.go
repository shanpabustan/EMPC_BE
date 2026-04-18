package mdlRbac

// Navigation/Access Models
type NavigationAccessReq struct {
	NavigationID int  `json:"navigation_id"`
	CanView      bool `json:"can_view"`
	CanAdd       bool `json:"can_add"`
	CanEdit      bool `json:"can_edit"`
	CanDelete    bool `json:"can_delete"`
	CanOverride  bool `json:"can_override"`
}

type NavigationItemReq struct {
	ParentID  *int   `json:"parent_id"`
	Label     string `json:"label"`
	Slug      string `json:"slug"`
	SortOrder int    `json:"sort_order"`
}

type NavigationItem struct {
	ID        int              `json:"id"`
	ParentID  *int             `json:"parent_id"`
	Label     string           `json:"label"`
	Slug      string           `json:"slug"`
	SortOrder int              `json:"sort_order"`
	Children  []NavigationItem `json:"children,omitempty"`
}

type RoleRequest struct {
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
}

type RoleResponse struct {
	ID          int    `json:"id"`
	RoleName    string `json:"role_name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type RoleNavigationAccess struct {
	RoleID          int    `json:"role_id"`
	RoleName        string `json:"role_name"`
	NavigationID    int    `json:"navigation_id"`
	NavigationLabel string `json:"navigation_label"`
	CanView         bool   `json:"can_view"`
	CanAdd          bool   `json:"can_add"`
	CanEdit         bool   `json:"can_edit"`
	CanDelete       bool   `json:"can_delete"`
	CanOverride     bool   `json:"can_override"`
}

type RoleWithNavigationAccess struct {
	RoleName    string                       `json:"role_name"`
	Navigations []NavigationAccessPermission `json:"navigations"`
}

type NavigationAccessPermission struct {
	NavigationID    int    `json:"navigation_id"`
	NavigationLabel string `json:"navigation_label"`
	Slug            string `json:"slug"`
	CanView         bool   `json:"can_view"`
	CanAdd          bool   `json:"can_add"`
	CanEdit         bool   `json:"can_edit"`
	CanDelete       bool   `json:"can_delete"`
	CanOverride     bool   `json:"can_override"`
}

type UserRoleResponse struct {
	StaffID   string `json:"staff_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	RoleID    *int   `json:"role_id"`
	RoleName  string `json:"role_name"`
}

type ActionResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NavigationTreeNode represents a node in the navigation tree
type NavigationTreeNode struct {
	Label       string               `json:"label"`
	Slug        string               `json:"slug"`
	CanAdd      bool                 `json:"can_add,omitempty"`
	CanEdit     bool                 `json:"can_edit,omitempty"`
	CanDelete   bool                 `json:"can_delete,omitempty"`
	CanOverride bool                 `json:"can_override,omitempty"`
	Children    []NavigationTreeNode `json:"children,omitempty"`
}
