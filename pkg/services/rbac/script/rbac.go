package scpRbac

import (
	"encoding/json"
	"fmt"
	"strings"

	"EMPC_BE/pkg/config"
	errRbac "EMPC_BE/pkg/services/rbac/error"
	mdlRbac "EMPC_BE/pkg/services/rbac/model"
)

// ============================================
// NAVIGATION / RBAC ACCESS MANAGEMENT
// ============================================

// AssignNavigationAccess assigns access permissions for a role to a navigation item
func AssignNavigationAccess(roleID int, navigationID int, canView, canAdd, canEdit, canDelete, canOverride bool) (*mdlRbac.ActionResult, error) {
	db := config.DBConnList[0]

	query := `
		INSERT INTO role_navigation_access (role_id, navigation_id, can_view, can_add, can_edit, can_delete, can_override)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (role_id, navigation_id) 
		DO UPDATE SET 
			can_view = EXCLUDED.can_view,
			can_add = EXCLUDED.can_add,
			can_edit = EXCLUDED.can_edit,
			can_delete = EXCLUDED.can_delete,
			can_override = EXCLUDED.can_override
		RETURNING role_id, navigation_id;
	`

	var result struct {
		RoleID       int `json:"role_id"`
		NavigationID int `json:"navigation_id"`
	}

	err := db.Raw(query, roleID, navigationID, canView, canAdd, canEdit, canDelete, canOverride).Scan(&result).Error
	if err != nil {
		return &mdlRbac.ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to assign navigation access: %v", err),
		}, err
	}

	return &mdlRbac.ActionResult{
		Success: true,
		Message: "Navigation access assigned successfully",
		Data:    result,
	}, nil
}

// GetAllRolesNavigationAccess retrieves all roles with their navigation access permissions
func GetAllRolesNavigationAccess() ([]mdlRbac.RoleWithNavigationAccess, error) {
	db := config.DBConnList[0]

	query := `
		SELECT 
			r.id as role_id,
			r.role_name,
			COALESCE(
				json_agg(
					json_build_object(
						'navigation_id', n.id,
						'navigation_label', n.label,
						'slug', n.slug,
						'can_view', rna.can_view,
						'can_add', rna.can_add,
						'can_edit', rna.can_edit,
						'can_delete', rna.can_delete,
						'can_override', rna.can_override
					)
					ORDER BY n.sort_order
				) FILTER (WHERE n.id IS NOT NULL),
				'[]'::json
			) as navigations
		FROM sys_roles r
		LEFT JOIN role_navigation_access rna ON r.id = rna.role_id
		LEFT JOIN sys_navigation n ON rna.navigation_id = n.id
		GROUP BY r.id, r.role_name
		ORDER BY r.id;
	`

	var results []struct {
		RoleID      int             `json:"role_id"`
		RoleName    string          `json:"role_name"`
		Navigations json.RawMessage `json:"navigations"`
	}

	err := db.Raw(query).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching role navigation access: %v", err)
	}

	var roleAccessList []mdlRbac.RoleWithNavigationAccess
	for _, r := range results {
		var navigations []mdlRbac.NavigationAccessPermission
		if err := json.Unmarshal(r.Navigations, &navigations); err != nil {
			return nil, fmt.Errorf("error parsing navigations: %v", err)
		}

		roleAccessList = append(roleAccessList, mdlRbac.RoleWithNavigationAccess{
			RoleName:    r.RoleName,
			Navigations: navigations,
		})
	}

	return roleAccessList, nil
}

// GetRoleNavigationAccess retrieves navigation access for a specific role
// Returns a hierarchical navigation tree with permissions
func GetRoleNavigationAccess(roleID int) (interface{}, error) {
	db := config.DBConnList[0]

	// First check if role exists
	var roleExists bool
	roleCheckQuery := `SELECT EXISTS(SELECT 1 FROM sys_roles WHERE id = $1)`
	err := db.Raw(roleCheckQuery, roleID).Scan(&roleExists).Error
	if err != nil {
		return nil, fmt.Errorf("error checking role existence: %v", err)
	}
	
	if !roleExists {
		return nil, errRbac.ErrResourceNotFound
	}

	// Get navigation tree as JSONB
	query := `SELECT get_nested_navigation($1)::jsonb as navigation_tree`
	
	var navigationTree string
	err = db.Raw(query, roleID).Scan(&navigationTree).Error
	if err != nil {
		return nil, fmt.Errorf("error fetching role navigation access: %v", err)
	}

	// If no results or empty, return empty array
	if navigationTree == "" || navigationTree == "null" {
		return []interface{}{}, nil
	}

	// Parse the JSON string into an interface
	var result interface{}
	if err := json.Unmarshal([]byte(navigationTree), &result); err != nil {
		return nil, fmt.Errorf("error parsing navigation tree: %v", err)
	}

	return result, nil
}

// RemoveNavigationAccess removes access permissions for a role from a navigation item
func RemoveNavigationAccess(roleID int, navigationID int) (*mdlRbac.ActionResult, error) {
	db := config.DBConnList[0]

	query := `DELETE FROM role_navigation_access WHERE role_id = ? AND navigation_id = ?`

	result := db.Exec(query, roleID, navigationID)
	if result.Error != nil {
		return &mdlRbac.ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to remove navigation access: %v", result.Error),
		}, result.Error
	}

	if result.RowsAffected == 0 {
		return &mdlRbac.ActionResult{
			Success: false,
			Message: "No access entry found to remove",
		}, nil
	}

	return &mdlRbac.ActionResult{
		Success: true,
		Message: "Navigation access removed successfully",
	}, nil
}

// ============================================
// NAVIGATION ITEMS MANAGEMENT
// ============================================

// CreateNavigation creates a new navigation item
func CreateNavigation(parentID *int, label, slug string, sortOrder int) (*mdlRbac.ActionResult, error) {
	db := config.DBConnList[0]

	query := `
		INSERT INTO sys_navigation (parent_id, label, slug, sort_order)
		VALUES (?, ?, ?, ?)
		RETURNING id, parent_id, label, slug, sort_order;
	`

	var result mdlRbac.NavigationItem
	err := db.Raw(query, parentID, label, slug, sortOrder).Scan(&result).Error
	if err != nil {
		return &mdlRbac.ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create navigation item: %v", err),
		}, err
	}

	return &mdlRbac.ActionResult{
		Success: true,
		Message: "Navigation item created successfully",
		Data:    result,
	}, nil
}

// GetAllNavigation retrieves all navigation items (flat list)
func GetAllNavigation() ([]mdlRbac.NavigationItem, error) {
	db := config.DBConnList[0]

	query := `
		SELECT id, parent_id, label, slug, sort_order
		FROM sys_navigation
		ORDER BY sort_order, id;
	`

	var navigation []mdlRbac.NavigationItem
	err := db.Raw(query).Scan(&navigation).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch navigation items: %v", err)
	}

	return navigation, nil
}

// GetNavigationTree retrieves navigation items as a hierarchical tree
func GetNavigationTree() ([]mdlRbac.NavigationItem, error) {
	db := config.DBConnList[0]

	// Get all navigation items
	query := `
		SELECT id, parent_id, label, slug, sort_order
		FROM sys_navigation
		ORDER BY sort_order, id;
	`

	var allItems []mdlRbac.NavigationItem
	err := db.Raw(query).Scan(&allItems).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch navigation items: %v", err)
	}

	// Build tree structure
	tree := buildNavigationTree(allItems, nil)
	return tree, nil
}

// buildNavigationTree recursively builds the navigation tree
func buildNavigationTree(items []mdlRbac.NavigationItem, parentID *int) []mdlRbac.NavigationItem {
	var tree []mdlRbac.NavigationItem

	for _, item := range items {
		if (parentID == nil && item.ParentID == nil) ||
			(parentID != nil && item.ParentID != nil && *item.ParentID == *parentID) {
			children := buildNavigationTree(items, &item.ID)
			if len(children) > 0 {
				item.Children = children
			}
			tree = append(tree, item)
		}
	}

	return tree
}

// UpdateNavigation updates an existing navigation item
func UpdateNavigation(navigationID int, parentID *int, label, slug string, sortOrder int) error {
	db := config.DBConnList[0]

	// Check if navigation item exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM sys_navigation WHERE id = ?)`
	err := db.Raw(checkQuery, navigationID).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check navigation existence: %v", err)
	}
	if !exists {
		return errRbac.ErrResourceNotFound
	}

	query := `
		UPDATE sys_navigation 
		SET parent_id = ?, label = ?, slug = ?, sort_order = ?
		WHERE id = ?;
	`

	result := db.Exec(query, parentID, label, slug, sortOrder, navigationID)
	if result.Error != nil {
		return fmt.Errorf("failed to update navigation item: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errRbac.ErrResourceNotFound
	}

	return nil
}

// DeleteNavigation removes a navigation item by its ID
func DeleteNavigation(navigationID int) error {
	db := config.DBConnList[0]

	// Check if navigation item has children
	var childCount int64
	childQuery := `SELECT COUNT(*) FROM sys_navigation WHERE parent_id = ?`
	err := db.Raw(childQuery, navigationID).Scan(&childCount).Error
	if err != nil {
		return fmt.Errorf("failed to check children: %v", err)
	}
	if childCount > 0 {
		return errRbac.ErrResourceInUse
	}

	query := `DELETE FROM sys_navigation WHERE id = ?`
	result := db.Exec(query, navigationID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete navigation item: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errRbac.ErrResourceNotFound
	}

	return nil
}

// ============================================
// ROLES MANAGEMENT
// ============================================

// CreateRole creates a new role
func CreateRole(roleName, description string) (*mdlRbac.ActionResult, error) {
	db := config.DBConnList[0]

	query := `
		INSERT INTO sys_roles (role_name, description, created_at)
		VALUES (?, ?, NOW())
		RETURNING id, role_name, description, created_at;
	`

	var result mdlRbac.RoleResponse
	err := db.Raw(query, roleName, description).Scan(&result).Error
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return &mdlRbac.ActionResult{
				Success: false,
				Message: "Role name already exists",
			}, errRbac.ErrResourceNameTaken
		}
		return &mdlRbac.ActionResult{
			Success: false,
			Message: fmt.Sprintf("Failed to create role: %v", err),
		}, err
	}

	return &mdlRbac.ActionResult{
		Success: true,
		Message: "Role created successfully",
		Data:    result,
	}, nil
}

// GetAllRoles retrieves all roles
func GetAllRoles() ([]mdlRbac.RoleResponse, error) {
	db := config.DBConnList[0]

	query := `
		SELECT id, role_name, description, created_at
		FROM sys_roles
		ORDER BY id;
	`

	var roles []mdlRbac.RoleResponse
	err := db.Raw(query).Scan(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %v", err)
	}

	return roles, nil
}

// UpdateRole updates an existing role
func UpdateRole(roleID int, roleName, description string) error {
	db := config.DBConnList[0]

	// Check if role exists
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM sys_roles WHERE id = ?)`
	err := db.Raw(checkQuery, roleID).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check role existence: %v", err)
	}
	if !exists {
		return errRbac.ErrResourceNotFound
	}

	query := `
		UPDATE sys_roles 
		SET role_name = ?, description = ?
		WHERE id = ?;
	`

	result := db.Exec(query, roleName, description, roleID)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "unique constraint") {
			return errRbac.ErrResourceNameTaken
		}
		return fmt.Errorf("failed to update role: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errRbac.ErrResourceNotFound
	}

	return nil
}

// DeleteRole removes a role by its ID
func DeleteRole(roleID int) error {
	db := config.DBConnList[0]

	// Check if role is assigned to any users
	var userCount int64
	userQuery := `SELECT COUNT(*) FROM users WHERE role_id = ?`
	err := db.Raw(userQuery, roleID).Scan(&userCount).Error
	if err != nil {
		return fmt.Errorf("failed to check user assignments: %v", err)
	}
	if userCount > 0 {
		return errRbac.ErrResourceInUse
	}

	query := `DELETE FROM sys_roles WHERE id = ?`
	result := db.Exec(query, roleID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete role: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errRbac.ErrResourceNotFound
	}

	return nil
}

// ============================================
// USER ROLE ASSIGNMENT
// ============================================

// AssignUserRole assigns a role to a user by staff ID
func AssignUserRole(staffID string, roleID int) error {
	db := config.DBConnList[0]

	// Check if user exists
	var userExists bool
	userCheckQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE staff_id = ?)`
	err := db.Raw(userCheckQuery, staffID).Scan(&userExists).Error
	if err != nil {
		return fmt.Errorf("failed to check user existence: %v", err)
	}
	if !userExists {
		return errRbac.ErrResourceNotFound
	}

	// Check if role exists
	var roleExists bool
	roleCheckQuery := `SELECT EXISTS(SELECT 1 FROM sys_roles WHERE id = ?)`
	err = db.Raw(roleCheckQuery, roleID).Scan(&roleExists).Error
	if err != nil {
		return fmt.Errorf("failed to check role existence: %v", err)
	}
	if !roleExists {
		return errRbac.ErrResourceNotFound
	}

	query := `UPDATE users SET role_id = ?, updated_at = NOW() WHERE staff_id = ?`
	result := db.Exec(query, roleID, staffID)
	if result.Error != nil {
		return fmt.Errorf("failed to assign role to user: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return errRbac.ErrResourceNotFound
	}

	return nil
}

// GetUserRole retrieves the role assigned to a user
func GetUserRole(staffID string) (*mdlRbac.UserRoleResponse, error) {
	db := config.DBConnList[0]

	query := `
		SELECT 
			u.staff_id,
			u.first_name,
			u.last_name,
			u.role_id,
			r.role_name
		FROM users u
		LEFT JOIN sys_roles r ON u.role_id = r.id
		WHERE u.staff_id = ?;
	`

	var userRole mdlRbac.UserRoleResponse
	err := db.Raw(query, staffID).Scan(&userRole).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user role: %v", err)
	}

	if userRole.StaffID == "" {
		return nil, errRbac.ErrResourceNotFound
	}

	return &userRole, nil
}

// GetAllUsersWithRoles retrieves all users with their assigned roles
func GetAllUsersWithRoles() ([]mdlRbac.UserRoleResponse, error) {
	db := config.DBConnList[0]

	query := `
		SELECT 
			u.staff_id,
			u.first_name,
			u.middle_name,
			u.last_name,
			u.email,
			u.phone_no,
			u.is_active,
			u.role_id,
			r.role_name
		FROM users u
		LEFT JOIN sys_roles r ON u.role_id = r.id
		WHERE u.deleted_at IS NULL
		ORDER BY u.staff_id;
	`

	var users []mdlRbac.UserRoleResponse
	err := db.Raw(query).Scan(&users).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users with roles: %v", err)
	}

	return users, nil
}