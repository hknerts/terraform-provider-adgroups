package client

import (
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// Group represents an Active Directory group
type Group struct {
	DN           string   `json:"dn"`
	CN           string   `json:"cn"`
	Name         string   `json:"name"`
	SamAccountName string `json:"sam_account_name"`
	Description  string   `json:"description"`
	GroupType    string   `json:"group_type"`
	ManagedBy    string   `json:"managed_by"`
	Members      []string `json:"members"`
	MemberOf     []string `json:"member_of"`
	ObjectGUID   string   `json:"object_guid"`
	ObjectSid    string   `json:"object_sid"`
}

// GetGroup retrieves a group by its distinguished name
func (c *Client) GetGroup(dn string) (*Group, error) {
	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=group)",
		[]string{
			"cn",
			"name",
			"sAMAccountName",
			"description",
			"groupType",
			"managedBy",
			"member",
			"memberOf",
			"objectGUID",
			"objectSid",
		},
		nil,
	)

	result, err := c.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for group %s: %w", dn, err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("group not found: %s", dn)
	}

	entry := result.Entries[0]
	group := &Group{
		DN:             entry.DN,
		CN:             entry.GetAttributeValue("cn"),
		Name:           entry.GetAttributeValue("name"),
		SamAccountName: entry.GetAttributeValue("sAMAccountName"),
		Description:    entry.GetAttributeValue("description"),
		GroupType:      entry.GetAttributeValue("groupType"),
		ManagedBy:      entry.GetAttributeValue("managedBy"),
		Members:        entry.GetAttributeValues("member"),
		MemberOf:       entry.GetAttributeValues("memberOf"),
		ObjectGUID:     entry.GetAttributeValue("objectGUID"),
		ObjectSid:      entry.GetAttributeValue("objectSid"),
	}

	return group, nil
}

// GetGroupByCN retrieves a group by its common name
func (c *Client) GetGroupByCN(cn string) (*Group, error) {
	filter := fmt.Sprintf("(&(objectClass=group)(cn=%s))", EscapeFilter(cn))
	
	searchRequest := ldap.NewSearchRequest(
		c.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{
			"cn",
			"name",
			"sAMAccountName",
			"description",
			"groupType",
			"managedBy",
			"member",
			"memberOf",
			"objectGUID",
			"objectSid",
		},
		nil,
	)

	result, err := c.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for group with CN %s: %w", cn, err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("group not found with CN: %s", cn)
	}

	entry := result.Entries[0]
	group := &Group{
		DN:             entry.DN,
		CN:             entry.GetAttributeValue("cn"),
		Name:           entry.GetAttributeValue("name"),
		SamAccountName: entry.GetAttributeValue("sAMAccountName"),
		Description:    entry.GetAttributeValue("description"),
		GroupType:      entry.GetAttributeValue("groupType"),
		ManagedBy:      entry.GetAttributeValue("managedBy"),
		Members:        entry.GetAttributeValues("member"),
		MemberOf:       entry.GetAttributeValues("memberOf"),
		ObjectGUID:     entry.GetAttributeValue("objectGUID"),
		ObjectSid:      entry.GetAttributeValue("objectSid"),
	}

	return group, nil
}

// CreateGroup creates a new group
func (c *Client) CreateGroup(ou, cn, description string, groupType int) (*Group, error) {
	dn := fmt.Sprintf("CN=%s,%s", EscapeDN(cn), ou)
	
	addRequest := ldap.NewAddRequest(dn, nil)
	addRequest.Attribute("objectClass", []string{"top", "group"})
	addRequest.Attribute("cn", []string{cn})
	addRequest.Attribute("name", []string{cn})
	addRequest.Attribute("sAMAccountName", []string{cn})
	
	if description != "" {
		addRequest.Attribute("description", []string{description})
	}
	
	addRequest.Attribute("groupType", []string{fmt.Sprintf("%d", groupType)})

	err := c.Add(addRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create group %s: %w", dn, err)
	}

	// Return the created group
	return c.GetGroup(dn)
}

// UpdateGroup updates an existing group
func (c *Client) UpdateGroup(dn string, updates map[string][]string) error {
	modifyRequest := ldap.NewModifyRequest(dn, nil)

	for attr, values := range updates {
		if len(values) == 0 {
			modifyRequest.Delete(attr, []string{})
		} else {
			modifyRequest.Replace(attr, values)
		}
	}

	err := c.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("failed to update group %s: %w", dn, err)
	}

	return nil
}

// DeleteGroup deletes a group
func (c *Client) DeleteGroup(dn string) error {
	delRequest := ldap.NewDelRequest(dn, nil)
	
	err := c.Delete(delRequest)
	if err != nil {
		return fmt.Errorf("failed to delete group %s: %w", dn, err)
	}

	return nil
}

// AddMemberToGroup adds a member to a group
func (c *Client) AddMemberToGroup(groupDN, memberDN string) error {
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Add("member", []string{memberDN})

	err := c.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("failed to add member %s to group %s: %w", memberDN, groupDN, err)
	}

	return nil
}

// RemoveMemberFromGroup removes a member from a group
func (c *Client) RemoveMemberFromGroup(groupDN, memberDN string) error {
	modifyRequest := ldap.NewModifyRequest(groupDN, nil)
	modifyRequest.Delete("member", []string{memberDN})

	err := c.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("failed to remove member %s from group %s: %w", memberDN, groupDN, err)
	}

	return nil
}

// ListGroups lists all groups in the directory
func (c *Client) ListGroups(filter string) ([]*Group, error) {
	if filter == "" {
		filter = "(objectClass=group)"
	}

	searchRequest := ldap.NewSearchRequest(
		c.baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{
			"cn",
			"name",
			"sAMAccountName",
			"description",
			"groupType",
			"managedBy",
			"member",
			"memberOf",
			"objectGUID",
			"objectSid",
		},
		nil,
	)

	result, err := c.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	var groups []*Group
	for _, entry := range result.Entries {
		group := &Group{
			DN:             entry.DN,
			CN:             entry.GetAttributeValue("cn"),
			Name:           entry.GetAttributeValue("name"),
			SamAccountName: entry.GetAttributeValue("sAMAccountName"),
			Description:    entry.GetAttributeValue("description"),
			GroupType:      entry.GetAttributeValue("groupType"),
			ManagedBy:      entry.GetAttributeValue("managedBy"),
			Members:        entry.GetAttributeValues("member"),
			MemberOf:       entry.GetAttributeValues("memberOf"),
			ObjectGUID:     entry.GetAttributeValue("objectGUID"),
			ObjectSid:      entry.GetAttributeValue("objectSid"),
		}
		groups = append(groups, group)
	}

	return groups, nil
}

// MoveGroup moves a group to a different organizational unit
func (c *Client) MoveGroup(currentDN, newParentDN string) error {
	// Extract the CN from current DN
	parts := strings.Split(currentDN, ",")
	if len(parts) == 0 {
		return fmt.Errorf("invalid DN format: %s", currentDN)
	}
	
	cn := parts[0]
	newDN := fmt.Sprintf("%s,%s", cn, newParentDN)

	// Create modify DN request
	modifyDNRequest := ldap.NewModifyDNRequest(currentDN, cn, true, newParentDN)
	
	err := c.conn.ModifyDN(modifyDNRequest)
	if err != nil {
		return fmt.Errorf("failed to move group from %s to %s: %w", currentDN, newDN, err)
	}

	return nil
}
