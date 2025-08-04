package client

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// Client represents an LDAP client for Active Directory operations
type Client struct {
	conn     *ldap.Conn
	baseDN   string
	username string
	password string
	server   string
	port     int
	useTLS   bool
}

// ClientConfig holds the configuration for the LDAP client
type ClientConfig struct {
	Server   string
	Port     int
	BaseDN   string
	Username string
	Password string
	UseTLS   bool
	Insecure bool
}

// NewClient creates a new LDAP client
func NewClient(config *ClientConfig) (*Client, error) {
	client := &Client{
		baseDN:   config.BaseDN,
		username: config.Username,
		password: config.Password,
		server:   config.Server,
		port:     config.Port,
		useTLS:   config.UseTLS,
	}

	err := client.connect(config.Insecure)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}

	return client, nil
}

// connect establishes a connection to the LDAP server
func (c *Client) connect(insecure bool) error {
	var err error
	address := fmt.Sprintf("%s:%d", c.server, c.port)

	if c.useTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: insecure,
		}
		c.conn, err = ldap.DialTLS("tcp", address, tlsConfig)
	} else {
		c.conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("failed to dial LDAP server: %w", err)
	}

	// Bind with credentials
	err = c.conn.Bind(c.username, c.password)
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to bind to LDAP server: %w", err)
	}

	return nil
}

// Close closes the LDAP connection
func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Close()
	}
	return nil
}

// Search performs an LDAP search
func (c *Client) Search(searchRequest *ldap.SearchRequest) (*ldap.SearchResult, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("LDAP connection is not established")
	}

	result, err := c.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	return result, nil
}

// Add adds an entry to LDAP
func (c *Client) Add(addRequest *ldap.AddRequest) error {
	if c.conn == nil {
		return fmt.Errorf("LDAP connection is not established")
	}

	err := c.conn.Add(addRequest)
	if err != nil {
		return fmt.Errorf("LDAP add failed: %w", err)
	}

	return nil
}

// Modify modifies an entry in LDAP
func (c *Client) Modify(modifyRequest *ldap.ModifyRequest) error {
	if c.conn == nil {
		return fmt.Errorf("LDAP connection is not established")
	}

	err := c.conn.Modify(modifyRequest)
	if err != nil {
		return fmt.Errorf("LDAP modify failed: %w", err)
	}

	return nil
}

// Delete deletes an entry from LDAP
func (c *Client) Delete(delRequest *ldap.DelRequest) error {
	if c.conn == nil {
		return fmt.Errorf("LDAP connection is not established")
	}

	err := c.conn.Del(delRequest)
	if err != nil {
		return fmt.Errorf("LDAP delete failed: %w", err)
	}

	return nil
}

// EscapeDN escapes special characters in a DN component
func EscapeDN(value string) string {
	// Escape special characters in DN values
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		",", "\\,",
		"+", "\\+",
		"\"", "\\\"",
		"<", "\\<",
		">", "\\>",
		";", "\\;",
		"=", "\\=",
		"\n", "\\0A",
		"\r", "\\0D",
		"\x00", "\\00",
	)
	return replacer.Replace(value)
}

// EscapeFilter escapes special characters in a search filter
func EscapeFilter(value string) string {
	// Escape special characters in filter values
	replacer := strings.NewReplacer(
		"\\", "\\5c",
		"*", "\\2a",
		"(", "\\28",
		")", "\\29",
		"\x00", "\\00",
	)
	return replacer.Replace(value)
}
