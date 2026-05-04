package secure_rag.document

# Пример ABAC‑политика для доступа к документам в компании.
#
# Документ может иметь следующие атрибуты (все необязательные):
# - public: true                           — доступен всем активным пользователям
# - required_department: "finance"         — доступ только сотрудникам отдела
# - required_clearance: 2                  — минимальный уровень допуска (1..N)
# - required_project: "phoenix"            — доступ участникам проекта
# - allowed_system_roles: ["access_admin"] — доступ по системной роли
#
# Логика:
# - если subject не активен -> deny
# - если public=true -> allow
# - иначе: все заданные ограничения должны выполняться одновременно

default allow := false

allow if {
	active_subject
	public_document
}

allow if {
	active_subject
	not public_document
	department_ok
	clearance_ok
	project_ok
	role_ok
}

active_subject := object.get(input.subject, "is_active", false) == true

public_document := object.get(input.document.attributes, "public", false) == true

department_ok if {
	required := object.get(input.document.attributes, "required_department", "")
	required == ""
}

department_ok if {
	required := object.get(input.document.attributes, "required_department", "")
	required == object.get(input.subject.attributes, "department", "")
}

clearance_ok if {
	required := object.get(input.document.attributes, "required_clearance", null)
	required == null
}

clearance_ok if {
	required := object.get(input.document.attributes, "required_clearance", null)
	object.get(input.subject.attributes, "clearance", 0) >= required
}

project_ok if {
	required := object.get(input.document.attributes, "required_project", "")
	required == ""
}

project_ok if {
	required := object.get(input.document.attributes, "required_project", "")
	subject_has_project(required)
}

subject_has_project(project) if {
	some p in object.get(input.subject.attributes, "projects", [])
	p == project
}

role_ok if {
	allowed := object.get(input.document.attributes, "allowed_system_roles", [])
	count(allowed) == 0
}

role_ok if {
	allowed := object.get(input.document.attributes, "allowed_system_roles", [])
	subject_has_any_role(allowed)
}

subject_has_any_role(allowed) if {
	some r in object.get(input.subject, "roles", [])
	some a in allowed
	r == a
}
