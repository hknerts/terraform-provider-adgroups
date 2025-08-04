package client

import (
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

// User represents an Active Directory user
type User struct {
	DN                string   `json:"dn"`
	CN                string   `json:"cn"`
	SamAccountName    string   `json:"sam_account_name"`
	UserPrincipalName string   `json:"user_principal_name"`
	DisplayName       string   `json:"display_name"`
	GivenName         string   `json:"given_name"`
	Surname           string   `json:"surname"`
	Email             string   `json:"email"`
	MemberOf          []string `json:"member_of"`
	ObjectGUID        string   `json:"object_guid"`
	ObjectSid         string   `json:"object_sid"`
}

// GetUser retrieves a user by their distinguished name
func (c *Client) GetUser(dn string) (*User, error) {
	searchRequest := ldap.NewSearchRequest(
		dn,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=user)",
		[]string{
			"cn",
			"sAMAccountName",
			"userPrincipalName",
			"displayName",
			"givenName",
			"sn",
			"mail",
			"memberOf",
			"objectGUID",
			"objectSid",
		},
		nil,
	)

	result, err := c.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user %s: %w", dn, err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found: %s", dn)
	}

	entry := result.Entries[0]
	user := &User{
		DN:                entry.DN,
		CN:                entry.GetAttributeValue("cn"),
		SamAccountName:    entry.GetAttributeValue("sAMAccountName"),
		UserPrincipalName: entry.GetAttributeValue("userPrincipalName"),
		DisplayName:       entry.GetAttributeValue("displayName"),
		GivenName:         entry.GetAttributeValue("givenName"),
		Surname:           entry.GetAttributeValue("sn"),
		Email:             entry.GetAttributeValue("mail"),
		MemberOf:          entry.GetAttributeValues("memberOf"),
		ObjectGUID:        entry.GetAttributeValue("objectGUID"),
		ObjectSid:         entry.GetAttributeValue("objectSid"),
	}

	return user, nil
}

// GetUserBySAM retrieves a user by their SAM account name
func (c *Client) GetUserBySAM(samAccountName string) (*User, error) {
	filter := fmt.Sprintf("(&(objectClass=user)(sAMAccountName=%s))", EscapeFilter(samAccountName))
	
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
			"sAMAccountName",
			"userPrincipalName",
			"displayName",
			"givenName",
			"sn",
			"mail",
			"memberOf",
			"objectGUID",
			"objectSid",
		},
		nil,
	)

	result, err := c.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user with SAM %s: %w", samAccountName, err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found with SAM: %s", samAccountName)
	}

	entry := result.Entries[0]
	user := &User{
		DN:                entry.DN,
		CN:                entry.GetAttributeValue("cn"),
		SamAccountName:    entry.GetAttributeValue("sAMAccountName"),
		UserPrincipalName: entry.GetAttributeValue("userPrincipalName"),
		DisplayName:       entry.GetAttributeValue("displayName"),
		GivenName:         entry.GetAttributeValue("givenName"),
		Surname:           entry.GetAttributeValue("sn"),
		Email:             entry.GetAttributeValue("mail"),
		MemberOf:          entry.GetAttributeValues("memberOf"),
		ObjectGUID:        entry.GetAttributeValue("objectGUID"),
		ObjectSid:         entry.GetAttributeValue("objectSid"),
	}

	return user, nil
}
