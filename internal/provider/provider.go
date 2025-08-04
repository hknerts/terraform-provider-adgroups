package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hknerts/terraform-provider-adgroups/internal/client"
)

// Ensure ADGroupsProvider satisfies various provider interfaces.
var _ provider.Provider = &ADGroupsProvider{}

// ADGroupsProvider defines the provider implementation.
type ADGroupsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ADGroupsProviderModel describes the provider data model.
type ADGroupsProviderModel struct {
	Server   types.String `tfsdk:"server"`
	Port     types.Int64  `tfsdk:"port"`
	BaseDN   types.String `tfsdk:"base_dn"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	UseTLS   types.Bool   `tfsdk:"use_tls"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *ADGroupsProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "adgroups"
	resp.Version = p.version
}

func (p *ADGroupsProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Active Directory Groups provider allows you to manage Active Directory groups and their memberships using LDAP.\\n\\n" +
			"This provider supports creating, reading, updating, and deleting AD groups, as well as managing group memberships.",
		Attributes: map[string]schema.Attribute{
			"server": schema.StringAttribute{
				MarkdownDescription: "Active Directory server hostname or IP address. Can also be set via the `AD_SERVER` environment variable.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "LDAP port (default: 389 for non-TLS, 636 for TLS). Can also be set via the `AD_PORT` environment variable.",
				Optional:            true,
			},
			"base_dn": schema.StringAttribute{
				MarkdownDescription: "Base Distinguished Name for LDAP operations. Can also be set via the `AD_BASE_DN` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for LDAP authentication. Can also be set via the `AD_USERNAME` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for LDAP authentication. Can also be set via the `AD_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"use_tls": schema.BoolAttribute{
				MarkdownDescription: "Use TLS for LDAP connection (default: false). Can also be set via the `AD_USE_TLS` environment variable.",
				Optional:            true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification (default: false). Can also be set via the `AD_INSECURE` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *ADGroupsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ADGroupsProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values and environment variable fallbacks
	server := data.Server.ValueString()
	if server == "" {
		server = os.Getenv("AD_SERVER")
	}
	if server == "" {
		resp.Diagnostics.AddError(
			"Missing AD Server Configuration",
			"The provider cannot create the AD client as there is a missing or empty value for the AD server. "+
				"Set the server value in the configuration or use the AD_SERVER environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	port := int(data.Port.ValueInt64())
	if port == 0 {
		if portEnv := os.Getenv("AD_PORT"); portEnv != "" {
			// Parse port from env var if needed
			port = 389 // default fallback
		} else {
			if data.UseTLS.ValueBool() || os.Getenv("AD_USE_TLS") == "true" {
				port = 636
			} else {
				port = 389
			}
		}
	}

	baseDN := data.BaseDN.ValueString()
	if baseDN == "" {
		baseDN = os.Getenv("AD_BASE_DN")
	}
	if baseDN == "" {
		resp.Diagnostics.AddError(
			"Missing AD Base DN Configuration",
			"The provider cannot create the AD client as there is a missing or empty value for the base DN. "+
				"Set the base_dn value in the configuration or use the AD_BASE_DN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	username := data.Username.ValueString()
	if username == "" {
		username = os.Getenv("AD_USERNAME")
	}
	if username == "" {
		resp.Diagnostics.AddError(
			"Missing AD Username Configuration",
			"The provider cannot create the AD client as there is a missing or empty value for the username. "+
				"Set the username value in the configuration or use the AD_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	password := data.Password.ValueString()
	if password == "" {
		password = os.Getenv("AD_PASSWORD")
	}
	if password == "" {
		resp.Diagnostics.AddError(
			"Missing AD Password Configuration",
			"The provider cannot create the AD client as there is a missing or empty value for the password. "+
				"Set the password value in the configuration or use the AD_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	useTLS := data.UseTLS.ValueBool()
	if os.Getenv("AD_USE_TLS") == "true" {
		useTLS = true
	}

	insecure := data.Insecure.ValueBool()
	if os.Getenv("AD_INSECURE") == "true" {
		insecure = true
	}

	// Create client
	config := &client.ClientConfig{
		Server:   server,
		Port:     port,
		BaseDN:   baseDN,
		Username: username,
		Password: password,
		UseTLS:   useTLS,
		Insecure: insecure,
	}

	adClient, err := client.NewClient(config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create AD Client",
			"An unexpected error occurred when creating the AD client. "+
				"If the error is not clear, please contact the provider developers.\\n\\n"+
				"AD Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = adClient
	resp.ResourceData = adClient
}

func (p *ADGroupsProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupResource,
		NewGroupMembershipResource,
	}
}

func (p *ADGroupsProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupDataSource,
		NewGroupsDataSource,
		NewUserDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ADGroupsProvider{
			version: version,
		}
	}
}
