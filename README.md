# Terraform Provider for Active Directory Groups

This provider allows Terraform to manage Active Directory groups and memberships using LDAP. It supports creating, reading, updating, and deleting AD groups and managing user memberships.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```bash
git clone https://github.com/hknerts/terraform-provider-adgroups
cd terraform-provider-adgroups
go install
```

## Adding the Provider to Your Terraform Configuration

With [Terraform v0.14 and later](https://www.terraform.io/upgrade-guides/0-14.html), development overrides for provider developers can be leveraged in order to use a locally-built provider rather than downloading one from the Terraform Registry.

Create a file called `.terraformrc` in your home directory, then add the dev_overrides block below:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/hashicorp/adgroups" = "/path/to/your/gopath/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

For information about this feature, see the [Terraform CLI documentation on provider installation](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers).

## Using the Provider

The provider requires configuration with Active Directory server details and authentication credentials.

### Provider Configuration

```hcl
terraform {
  required_providers {
    adgroups = {
      source = "hknerts/adgroups"
      version = "~> 1.0"
    }
  }
}

provider "adgroups" {
  server_host     = "ldap.example.com"
  server_port     = 389
  use_tls         = true
  bind_dn         = "CN=admin,DC=example,DC=com"
  bind_password   = var.ad_password
  base_dn         = "DC=example,DC=com"
}
```

### Provider Configuration Options

| Argument | Required | Default | Description |
|----------|----------|---------|-------------|
| `server_host` | Yes | | LDAP server hostname or IP address |
| `server_port` | No | 389 | LDAP server port (389 for LDAP, 636 for LDAPS) |
| `use_tls` | No | false | Whether to use TLS encryption |
| `bind_dn` | Yes | | Distinguished name for binding to LDAP |
| `bind_password` | Yes | | Password for the bind DN |
| `base_dn` | Yes | | Base DN for LDAP operations |

### Example Usage

#### Creating an AD Group

```hcl
resource "adgroups_group" "example" {
  name        = "example-group"
  description = "Example Active Directory group"
  ou_path     = "OU=Groups,DC=example,DC=com"
  group_scope = "Global"
  group_type  = "Security"
}
```

#### Managing Group Membership

```hcl
resource "adgroups_group_membership" "example" {
  group_dn = adgroups_group.example.dn
  members = [
    "CN=user1,OU=Users,DC=example,DC=com",
    "CN=user2,OU=Users,DC=example,DC=com"
  ]
}
```

#### Reading Group Information

```hcl
data "adgroups_group" "existing" {
  name = "existing-group"
}

output "group_dn" {
  value = data.adgroups_group.existing.dn
}
```

## Resources

### `adgroups_group`

Manages an Active Directory group.

**Arguments:**
- `name` (Required) - The name of the group
- `description` (Optional) - Description of the group
- `ou_path` (Optional) - Organizational Unit path where the group will be created
- `group_scope` (Optional) - Group scope: `DomainLocal`, `Global`, or `Universal`. Default: `Global`
- `group_type` (Optional) - Group type: `Security` or `Distribution`. Default: `Security`

**Attributes:**
- `id` - The group's distinguished name (DN)
- `dn` - The group's distinguished name
- `sid` - The group's security identifier

### `adgroups_group_membership`

Manages membership of an Active Directory group.

**Arguments:**
- `group_dn` (Required) - Distinguished name of the group
- `members` (Required) - List of member distinguished names

**Attributes:**
- `id` - The group DN

## Data Sources

### `adgroups_group`

Retrieves information about an existing Active Directory group.

**Arguments:**
- `name` (Required) - The name of the group to look up

**Attributes:**
- `id` - The group's distinguished name
- `dn` - The group's distinguished name
- `name` - The group's name
- `description` - The group's description
- `sid` - The group's security identifier
- `members` - List of group member distinguished names

## Testing the Provider

### Prerequisites for Testing

Before running tests, you need:

1. **Access to an Active Directory environment** - This can be:
   - A Windows Server with AD DS installed
   - A test Active Directory environment
   - A Docker container running Samba AD

2. **Test credentials with appropriate permissions**:
   - Create/delete groups
   - Modify group memberships
   - Read directory information

3. **Environment variables set**:
   ```bash
   export TF_ACC=1
   export AD_SERVER_HOST="your-ad-server.com"
   export AD_SERVER_PORT="389"
   export AD_USE_TLS="true"
   export AD_BIND_DN="CN=admin,DC=example,DC=com"
   export AD_BIND_PASSWORD="your-password"
   export AD_BASE_DN="DC=example,DC=com"
   ```

### Running Tests

#### Unit Tests
Run unit tests that don't require external dependencies:
```bash
go test ./internal/...
```

#### Acceptance Tests
Run full acceptance tests against a real AD environment:
```bash
make testacc
```

Or run specific test:
```bash
go test -v ./internal/provider -run TestAccGroupResource_basic
```

#### Manual Testing

1. **Set up a test environment**:
   ```bash
   # Create a test directory
   mkdir test-provider
   cd test-provider
   
   # Create a simple Terraform configuration
   cat > main.tf << EOF
   terraform {
     required_providers {
       adgroups = {
         source = "hknerts/adgroups"
         version = "~> 1.0"
       }
     }
   }
   
   provider "adgroups" {
     server_host   = "your-ad-server.com"
     server_port   = 389
     use_tls       = true
     bind_dn       = "CN=admin,DC=example,DC=com"
     bind_password = var.ad_password
     base_dn       = "DC=example,DC=com"
   }
   
   resource "adgroups_group" "test" {
     name        = "terraform-test-group"
     description = "Test group created by Terraform"
     ou_path     = "OU=Groups,DC=example,DC=com"
   }
   EOF
   ```

2. **Initialize and apply**:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

3. **Verify the group was created** in your Active Directory

4. **Clean up**:
   ```bash
   terraform destroy
   ```

### Setting Up a Test Environment

#### Option 1: Docker with Samba AD (Recommended for development)

```bash
# Run Samba AD container for testing
docker run -d \
  --name samba-ad \
  -p 389:389 \
  -p 636:636 \
  -e DOMAIN=EXAMPLE.COM \
  -e DOMAINPASS=YourPassword123! \
  -e DNSFORWARDER=8.8.8.8 \
  nowsci/samba-domain

# Wait for container to start, then test connection
ldapsearch -x -H ldap://localhost:389 -D "Administrator@EXAMPLE.COM" -w "YourPassword123!" -b "DC=example,DC=com"
```

#### Option 2: Windows Server with AD DS

1. Install Windows Server (evaluation versions available)
2. Add Active Directory Domain Services role
3. Promote server to domain controller
4. Create test organizational units and users

### Test Coverage

The provider includes tests for:
- Group creation, reading, updating, and deletion
- Group membership management
- Data source functionality
- Error handling and edge cases
- LDAP connection management

### Debugging Tests

Enable verbose logging:
```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
go test -v ./internal/provider -run TestAccGroupResource_basic
```

## Development

### Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Add tests for new functionality
6. Run tests locally
7. Submit a pull request

### Code Style

This project follows standard Go conventions and uses:
- `gofmt` for formatting
- `golangci-lint` for linting
- Go modules for dependency management

Run checks locally:
```bash
make fmt
make lint
make test
```

## License

This provider is distributed under the [Mozilla Public License 2.0](https://www.mozilla.org/en-US/MPL/2.0/).

## Support

For questions and support:
- Open an issue on [GitHub](https://github.com/hknerts/terraform-provider-adgroups/issues)
- Check the [documentation](https://registry.terraform.io/providers/hknerts/adgroups/latest/docs)
