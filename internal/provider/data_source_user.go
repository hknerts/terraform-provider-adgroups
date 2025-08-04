package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hknerts/terraform-provider-adgroups/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	client *client.Client
}

// UserDataSourceModel describes the data source data model.
type UserDataSourceModel struct {
	ID             types.String   `tfsdk:"id"`
	DN             types.String   `tfsdk:"dn"`
	CN             types.String   `tfsdk:"cn"`
	SamAccountName types.String   `tfsdk:"sam_account_name"`
	UserPrincipalName types.String `tfsdk:"user_principal_name"`
	DisplayName    types.String   `tfsdk:"display_name"`
	GivenName      types.String   `tfsdk:"given_name"`
	Surname        types.String   `tfsdk:"surname"`
	Email          types.String   `tfsdk:"email"`
	MemberOf       []types.String `tfsdk:"member_of"`
	ObjectGUID     types.String   `tfsdk:"object_guid"`
	ObjectSid      types.String   `tfsdk:"object_sid"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about an Active Directory user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the user (same as DN).",
			},
			"dn": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Distinguished Name of the user. Either 'dn' or 'sam_account_name' must be specified.",
			},
			"cn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Common Name of the user.",
			},
			"sam_account_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Security Account Manager (SAM) account name. Either 'dn' or 'sam_account_name' must be specified.",
			},
			"user_principal_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "User Principal Name (UPN).",
			},
			"display_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Display name of the user.",
			},
			"given_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Given name (first name) of the user.",
			},
			"surname": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Surname (last name) of the user.",
			},
			"email": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Email address of the user.",
			},
			"member_of": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of Distinguished Names of groups this user is a member of.",
			},
			"object_guid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The objectGUID of the user.",
			},
			"object_sid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The objectSid of the user.",
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For now, this is a placeholder implementation
	// In a full implementation, you would add user lookup methods to the client
	resp.Diagnostics.AddError(
		"Not Implemented",
		"User data source is not yet implemented. This is a placeholder for future functionality.",
	)
	return

	// TODO: Implement user lookup functionality
	// var user *client.User
	// var err error
	//
	// // Either DN or SAM account name must be specified
	// if !data.DN.IsNull() && !data.DN.IsUnknown() {
	//     user, err = d.client.GetUser(data.DN.ValueString())
	// } else if !data.SamAccountName.IsNull() && !data.SamAccountName.IsUnknown() {
	//     user, err = d.client.GetUserBySAM(data.SamAccountName.ValueString())
	// } else {
	//     resp.Diagnostics.AddError(
	//         "Missing Required Attribute",
	//         "Either 'dn' or 'sam_account_name' must be specified to look up the user.",
	//     )
	//     return
	// }
	//
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
	//     return
	// }
	//
	// // Map response to the data model
	// data.ID = types.StringValue(user.DN)
	// data.DN = types.StringValue(user.DN)
	// data.CN = types.StringValue(user.CN)
	// data.SamAccountName = types.StringValue(user.SamAccountName)
	// data.UserPrincipalName = types.StringValue(user.UserPrincipalName)
	// data.DisplayName = types.StringValue(user.DisplayName)
	// data.GivenName = types.StringValue(user.GivenName)
	// data.Surname = types.StringValue(user.Surname)
	// data.Email = types.StringValue(user.Email)
	// data.ObjectGUID = types.StringValue(user.ObjectGUID)
	// data.ObjectSid = types.StringValue(user.ObjectSid)
	//
	// // Convert memberOf slice
	// memberOf := make([]types.String, len(user.MemberOf))
	// for i, member := range user.MemberOf {
	//     memberOf[i] = types.StringValue(member)
	// }
	// data.MemberOf = memberOf
	//
	// // Save data into Terraform state
	// resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
