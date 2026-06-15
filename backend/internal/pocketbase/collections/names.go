package collections

const (
	Groups           = "hb_groups"
	Users            = "users"
	Locations        = "hb_locations"
	Labels           = "hb_labels"
	Products         = "hb_products"
	Items            = "hb_items"
	ItemFields       = "hb_item_fields"
	Attachments      = "hb_attachments"
	Maintenance      = "hb_maintenance_entries"
	Notifiers        = "hb_notifiers"
	GroupInvitations = "hb_group_invitations"
)

// GroupScopeRule allows any group member to read records in their group.
const GroupScopeRule = `@request.auth.id != "" && group = @request.auth.group`

// GroupMemberWriteRule allows any group member to create or update records in their group.
const GroupMemberWriteRule = `@request.auth.id != "" && group = @request.auth.group`

// GroupOwnerWriteRule allows only group owners to create, update, or delete group data.
const GroupOwnerWriteRule = `@request.auth.id != "" && group = @request.auth.group && @request.auth.role = "owner"`

// GroupMemberViewRule allows a user to view their own group record.
const GroupMemberViewRule = `@request.auth.id != "" && id = @request.auth.group`

// GroupOwnerUpdateRule allows only the group owner to update the group record.
const GroupOwnerUpdateRule = `@request.auth.id != "" && id = @request.auth.group && @request.auth.role = "owner"`

// InvitationOwnerWriteRule allows only group owners to manage invitations.
const InvitationOwnerWriteRule = `@request.auth.id != "" && @request.auth.role = "owner" && group = @request.auth.group`

// UserSelfUpdateRule allows users to update their own profile but not escalate role or change group.
const UserSelfUpdateRule = `@request.auth.id != "" && id = @request.auth.id && (@request.body.role:isset = false || @request.body.role = @request.auth.role) && (@request.body.group:isset = false || @request.body.group = @request.auth.group)`
