package secure_rag.document

# ABAC-политика доступа к документам
# Учитывает активность субъекта и ограничения документа по отделу, допуску, проекту и роли

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
