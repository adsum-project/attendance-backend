package authorization

import (
	"context"
	"slices"
)

func IsOwner(ctx context.Context, ownerID string) bool {
	userID, _ := ctx.Value("userID").(string)
	if userID != ownerID {
		return false
	}
	return true
}

func IsOwnerOrAdmin(ctx context.Context, ownerID string) bool {
	return IsOwner(ctx, ownerID) || HasRole(ctx, "admin")
}

func HasRole(ctx context.Context, role string) bool {
	claims, _ := ctx.Value("claims").(map[string]any)
	return slices.Contains(getRolesFromClaims(claims), role)
}

func HasRoles(ctx context.Context, roles ...string) bool {
	claims, _ := ctx.Value("claims").(map[string]any)
	userRoles := getRolesFromClaims(claims)
	for _, role := range roles {
		if slices.Contains(userRoles, role) {
			return true
		}
	}
	return false
}

func HasAllRoles(ctx context.Context, roles ...string) bool {
	claims, _ := ctx.Value("claims").(map[string]any)
	userRoles := getRolesFromClaims(claims)
	for _, role := range roles {
		if !slices.Contains(userRoles, role) {
			return false
		}
	}
	return true
}


func getRolesFromClaims(claims map[string]any) []string {
	if claims == nil {
		return nil
	}
	raw := claims["roles"]
	if raw == nil {
		return nil
	}
	if strSlice, ok := raw.([]string); ok {
		return strSlice
	}
	slice, ok := raw.([]interface{})
	if !ok {
		return nil
	}
	var roles []string
	for _, item := range slice {
		if role, ok := item.(string); ok {
			roles = append(roles, role)
		}
	}
	return roles
}