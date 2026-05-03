package usecase

import (
	"strings"

	iamv1 "secure-rag-platform/services/iam/gen/v1"
)

func allowedByAttributes(attrs map[string]any, subject *iamv1.SubjectContext) bool {
	if subject == nil {
		return true
	}
	if len(attrs) == 0 {
		return true
	}

	hasRule := false
	allowed := false

	if v, ok := attrs["public"]; ok {
		hasRule = true
		if flag, ok := v.(bool); ok && flag {
			allowed = true
		}
	}

	if v, ok := attrs["allowed_roles"]; ok {
		hasRule = true
		roles := toStringSlice(v)
		if intersects(roles, subject.GetRoles()) {
			allowed = true
		}
	}

	if v, ok := attrs["allowed_users"]; ok {
		hasRule = true
		users := toStringSlice(v)
		for _, user := range users {
			if strings.TrimSpace(user) == subject.GetUserId() {
				allowed = true
				break
			}
		}
	}

	if !hasRule {
		return true
	}
	return allowed
}

func toStringSlice(value any) []string {
	switch v := value.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	default:
		return nil
	}
}

func intersects(left []string, right []string) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(left))
	for _, item := range left {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		set[item] = struct{}{}
	}
	for _, item := range right {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := set[item]; ok {
			return true
		}
	}
	return false
}
