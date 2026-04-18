package ctrRbac

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	helper "EMPC_BE/pkg/global/json_response"
	errRbac "EMPC_BE/pkg/services/rbac/error"
	mdlRbac "EMPC_BE/pkg/services/rbac/model"
	scpRbac "EMPC_BE/pkg/services/rbac/script"

	"github.com/FDSAP-Git-Org/hephaestus/respcode"
	"github.com/gofiber/fiber/v3"
)

// ============================================
// NAVIGATION / RBAC ACCESS MANAGEMENT
// ============================================

// AssignNavigationAccess assigns access permissions for a role to a navigation item
func AssignNavigationAccess(c fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"Invalid role ID", err, http.StatusBadRequest)
	}

	var req mdlRbac.NavigationAccessReq
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validation
	if roleID <= 0 || req.NavigationID <= 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing required fields (roleId, navigationId)", http.StatusBadRequest)
	}

	// Call service
	result, err := scpRbac.AssignNavigationAccess(roleID, req.NavigationID, 
		req.CanView, req.CanAdd, req.CanEdit, req.CanDelete, req.CanOverride)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to assign navigation access", err, http.StatusInternalServerError)
	}

	if !result.Success {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			result.Message, http.StatusBadRequest)
	}

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		result.Message, http.StatusOK)
}

// GetAllRolesNavigationAccess retrieves all roles with their navigation access permissions
func GetAllRolesNavigationAccess(c fiber.Ctx) error {
	allAccess, err := scpRbac.GetAllRolesNavigationAccess()
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch roles and navigation access", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Successfully fetched all role navigation access", allAccess, http.StatusOK)
}

// GetRoleNavigationAccess retrieves navigation access for a specific role
func GetRoleNavigationAccess(c fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID <= 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Invalid role ID", http.StatusBadRequest)
	}

	roleAccess, err := scpRbac.GetRoleNavigationAccess(roleID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_404,
				"Role not found", http.StatusNotFound)
		}

		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch role navigation access", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Fetching role navigation access successful", roleAccess, http.StatusOK)
}

// RemoveNavigationAccess removes access permissions for a role from a navigation item
func RemoveNavigationAccess(c fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"Invalid role ID", err, http.StatusBadRequest)
	}

	var req mdlRbac.NavigationAccessReq
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Invalid request body", err, http.StatusBadRequest)
	}

	// Validation
	if roleID <= 0 || req.NavigationID <= 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing required fields (roleId, navigationId)", http.StatusBadRequest)
	}

	// Call service
	result, err := scpRbac.RemoveNavigationAccess(roleID, req.NavigationID)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			result.Message, err, http.StatusInternalServerError)
	}

	if !result.Success {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			result.Message, http.StatusBadRequest)
	}
	cleanMessage := strings.ReplaceAll(result.Message, `\"`, "")

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		cleanMessage, http.StatusOK)
}

// ============================================
// NAVIGATION ITEMS MANAGEMENT
// ============================================

// CreateNavigation creates a new navigation item
func CreateNavigation(c fiber.Ctx) error {
	var req mdlRbac.NavigationItemReq
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Invalid request body", err, http.StatusBadRequest)
	}

	if req.Label == "" {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing navigation label", http.StatusBadRequest)
	}

	result, err := scpRbac.CreateNavigation(req.ParentID, req.Label, req.Slug, req.SortOrder)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to create navigation item", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Navigation item created successfully", result, http.StatusOK)
}

// GetAllNavigation retrieves all navigation items (flat list)
func GetAllNavigation(c fiber.Ctx) error {
	navigation, err := scpRbac.GetAllNavigation()
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch navigation items", err, http.StatusInternalServerError)
	}

	if len(navigation) == 0 {
		return helper.JSONResponseV1(c, respcode.SUC_CODE_204,
			"No navigation items found", http.StatusOK)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Navigation items fetched successfully", navigation, http.StatusOK)
}

// GetNavigationTree retrieves navigation items as a hierarchical tree
func GetNavigationTree(c fiber.Ctx) error {
	tree, err := scpRbac.GetNavigationTree()
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch navigation tree", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Navigation tree fetched successfully", tree, http.StatusOK)
}

// UpdateNavigation updates an existing navigation item
func UpdateNavigation(c fiber.Ctx) error {
	navigationIDStr := c.Params("navigationId")
	navigationID, err := strconv.Atoi(navigationIDStr)
	if err != nil || navigationID == 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Invalid or missing navigation_id parameter", http.StatusBadRequest)
	}

	var req mdlRbac.NavigationItemReq
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Invalid request body", err, http.StatusBadRequest)
	}

	if req.Label == "" {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing navigation label", http.StatusBadRequest)
	}

	err = scpRbac.UpdateNavigation(navigationID, req.ParentID, req.Label, req.Slug, req.SortOrder)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to update navigation item", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		"Navigation item updated successfully", http.StatusOK)
}

// DeleteNavigation removes a navigation item by its ID
func DeleteNavigation(c fiber.Ctx) error {
	navigationIDStr := c.Params("navigationId")
	navigationID, err := strconv.Atoi(navigationIDStr)
	if err != nil || navigationID == 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Invalid or missing navigation_id parameter", http.StatusBadRequest)
	}

	err = scpRbac.DeleteNavigation(navigationID)
	if err != nil {
		if errors.Is(err, errRbac.ErrResourceNotFound) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_404,
				"Navigation item not found", http.StatusNotFound)
		}
		if errors.Is(err, errRbac.ErrResourceInUse) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_409,
				"Navigation item has children and cannot be deleted", http.StatusConflict)
		}
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to delete navigation item", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		"Navigation item deleted successfully", http.StatusOK)
}

// ============================================
// ROLES MANAGEMENT
// ============================================

// CreateRole creates a new role
func CreateRole(c fiber.Ctx) error {
	var req mdlRbac.RoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Invalid request body", err, http.StatusBadRequest)
	}

	if req.RoleName == "" {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing role name", http.StatusBadRequest)
	}

	result, err := scpRbac.CreateRole(req.RoleName, req.Description)
	if err != nil {
		if errors.Is(err, errRbac.ErrResourceNameTaken) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_409,
				"Role name already exists", http.StatusConflict)
		}
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to create role", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Role created successfully", result, http.StatusOK)
}

// GetAllRoles retrieves all roles
func GetAllRoles(c fiber.Ctx) error {
	roles, err := scpRbac.GetAllRoles()
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch roles", err, http.StatusInternalServerError)
	}

	if len(roles) == 0 {
		return helper.JSONResponseV1(c, respcode.SUC_CODE_204,
			"No roles found", http.StatusOK)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Roles fetched successfully", roles, http.StatusOK)
}

// UpdateRole updates an existing role
func UpdateRole(c fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID == 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Invalid or missing role_id parameter", http.StatusBadRequest)
	}

	var req mdlRbac.RoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_301,
			"Invalid request body", err, http.StatusBadRequest)
	}

	if req.RoleName == "" {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing role name", http.StatusBadRequest)
	}

	err = scpRbac.UpdateRole(roleID, req.RoleName, req.Description)
	if err != nil {
		if errors.Is(err, errRbac.ErrResourceNotFound) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_404,
				"Role not found", http.StatusNotFound)
		}
		if errors.Is(err, errRbac.ErrResourceNameTaken) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_409,
				"Role name already exists", http.StatusConflict)
		}
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to update role", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		"Role updated successfully", http.StatusOK)
}

// DeleteRole removes a role by its ID
func DeleteRole(c fiber.Ctx) error {
	roleIDStr := c.Params("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil || roleID == 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Invalid or missing role_id parameter", http.StatusBadRequest)
	}

	err = scpRbac.DeleteRole(roleID)
	if err != nil {
		if errors.Is(err, errRbac.ErrResourceNotFound) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_404,
				"Role not found", http.StatusNotFound)
		}
		if errors.Is(err, errRbac.ErrResourceInUse) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_409,
				"Role is assigned to users and cannot be deleted", http.StatusConflict)
		}
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to delete role", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		"Role deleted successfully", http.StatusOK)
}

// ============================================
// USER ROLE ASSIGNMENT
// ============================================

// AssignRoleToUser assigns a role to a user
func AssignRoleToUser(c fiber.Ctx) error {
	staffID := c.Params("staffId")
	roleIDStr := c.Params("roleId")
	roleID, err := strconv.Atoi(roleIDStr)
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_400,
			"Invalid role ID", err, http.StatusBadRequest)
	}

	if staffID == "" || roleID == 0 {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing staff_id or role_id", http.StatusBadRequest)
	}

	if err := scpRbac.AssignUserRole(staffID, roleID); err != nil {
		if errors.Is(err, errRbac.ErrResourceNotFound) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_404,
				"User not found", http.StatusNotFound)
		}
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to assign role to user", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseV1(c, respcode.SUC_CODE_200,
		"Role assigned to user successfully", http.StatusOK)
}

// GetUserRole retrieves the role assigned to a user
func GetUserRole(c fiber.Ctx) error {
	staffID := c.Params("staffId")

	if staffID == "" {
		return helper.JSONResponseV1(c, respcode.ERR_CODE_400,
			"Missing staff_id", http.StatusBadRequest)
	}

	userRole, err := scpRbac.GetUserRole(staffID)
	if err != nil {
		if errors.Is(err, errRbac.ErrResourceNotFound) {
			return helper.JSONResponseV1(c, respcode.ERR_CODE_404,
				"User not found", http.StatusNotFound)
		}
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch user role", err, http.StatusInternalServerError)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"User role fetched successfully", userRole, http.StatusOK)
}

// GetAllUsersWithRoles retrieves all users with their assigned roles
func GetAllUsersWithRoles(c fiber.Ctx) error {
	users, err := scpRbac.GetAllUsersWithRoles()
	if err != nil {
		return helper.JSONResponseWithErrorV1(c, respcode.ERR_CODE_500,
			"Failed to fetch users with roles", err, http.StatusInternalServerError)
	}

	if len(users) == 0 {
		return helper.JSONResponseV1(c, respcode.SUC_CODE_204,
			"No users found", http.StatusOK)
	}

	return helper.JSONResponseWithDataV1(c, respcode.SUC_CODE_200,
		"Users fetched successfully", users, http.StatusOK)
}