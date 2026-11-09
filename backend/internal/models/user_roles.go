package models

// Public product names map to the established persisted role vocabulary.
func ParseUserRole(value string) (UserRole, bool) {
	switch value {
	case "EMPLOYEE", "employee":
		return RoleEmployee, true
	case "OFFICE_MANAGER", "manager":
		return RoleManager, true
	case "ADMIN", "admin":
		return RoleAdmin, true
	default:
		return "", false
	}
}
