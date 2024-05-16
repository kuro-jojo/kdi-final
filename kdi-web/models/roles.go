package models

import "slices"

// Role constants
const (
	// Teamspaces roles (or admin roles)
	UpdateTeamspaceRole = "UPDATE_TEAMSPACE"
	DeleteTeamspaceRole = "DELETE_TEAMSPACE"

	// Project roles
	CreateProjectRole = "CREATE_PROJECT"
	DeleteProjectRole = "DELETE_PROJECT"
	UpdateProjectRole = "UPDATE_PROJECT"
	ViewProjectRole   = "VIEW_PROJECT"
	ListProjectsRole  = "LIST_PROJECTS"

	// Member roles
	AddMemberRole    = "ADD_MEMBER"
	RemoveMemberRole = "REMOVE_MEMBER"
	UpdateMemberRole = "UPDATE_MEMBER_PROFILE"
	ViewMemberRole   = "VIEW_MEMBER"
	ListMembersRole  = "LIST_MEMBERS"

	// Cluster roles
	AddClusterRole    = "ADD_CLUSTER"
	DeleteClusterRole = "DELETE_CLUSTER"
	UpdateClusterRole = "UPDATE_CLUSTER"
	ViewClusterRole   = "VIEW_CLUSTER"
	ListClustersRole  = "LIST_CLUSTERS"
)

func GetRoles() []string {
	return []string{
		UpdateTeamspaceRole,
		DeleteTeamspaceRole,

		CreateProjectRole,
		DeleteProjectRole,
		UpdateProjectRole,
		ViewProjectRole,
		ListProjectsRole,

		AddMemberRole,
		RemoveMemberRole,
		UpdateMemberRole,
		ViewMemberRole,
		ListMembersRole,

		AddClusterRole,
		DeleteClusterRole,
		UpdateClusterRole,
		ViewClusterRole,
		ListClustersRole,
	}
}

func IsRoleValid(role string) bool {
	roles := GetRoles()

	return slices.Contains(roles, role)
}
