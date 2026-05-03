package secure_rag.document

default allow := false

allow if {
	input.document.attributes.public == true
}

allow if {
	input.subject.is_active == true
	input.document.attributes.required_department == input.subject.attributes.department
}

allow if {
	input.subject.is_active == true
	input.document.attributes.required_clearance <= input.subject.attributes.clearance
}

allow if {
	input.subject.is_active == true
	required_project := input.document.attributes.required_project
	some project in input.subject.attributes.projects
	project == required_project
}

allow if {
	input.subject.is_active == true
	some role in input.subject.roles
	some allowed in input.document.attributes.allowed_system_roles
	role == allowed
}
