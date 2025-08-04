package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hknerts/terraform-provider-adgroups/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &GroupDataSource{}

func NewGroupDataSource() datasource.DataSource {
	return &GroupDataSource{}
}

// GroupDataSource defines the data source implementation.
type GroupDataSource struct {
	client *client.Client
}

// GroupDataSourceModel describes the data source data model.
type GroupDataSourceModel struct {
	ID             types.String   `tfsdk:"id"`
	DN             types.String   `tfsdk:"dn"`
	CN             types.String   `tfsdk:"cn"`
	Name           types.String   `tfsdk:"name"`
	SamAccountName types.String   `tfsdk:"sam_account_name"`
	Description    types.String   `tfsdk:"description"`
	GroupType      types.Int64    `tfsdk:"group_type"`
	ManagedBy      types.String   `tfsdk:"managed_by"`
	Members        []types.String `tfsdk:"members"`
	MemberOf       []types.String `tfsdk:"member_of"`
	ObjectGUID     types.String   `tfsdk:"object_guid"`
	ObjectSid      types.String   `tfsdk:"object_sid"`
}

func (d *GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (d *GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches information about an Active Directory group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the group (same as DN).",
			},
			"dn": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Distinguished Name of the group. Either 'dn' or 'cn' must be specified.",
			},
			"cn": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Common Name of the group. Either 'dn' or 'cn' must be specified.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Display name of the group.",
			},
			"sam_account_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Security Account Manager (SAM) account name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of the group.",
			},
			"group_type": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Group type value.",
			},
			"managed_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Distinguished Name of the user or group that manages this group.",
			},
			"members": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of Distinguished Names of group members.",
			},
			"member_of": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of Distinguished Names of groups this group is a member of.",
			},
			"object_guid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The objectGUID of the group.",
			},
			"object_sid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The objectSid of the group.",
			},
		},
	}
}

func (d *GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var group *client.Group
	var err error

	// Either DN or CN must be specified
	if !data.DN.IsNull() && !data.DN.IsUnknown() {
		group, err = d.client.GetGroup(data.DN.ValueString())
	} else if !data.CN.IsNull() && !data.CN.IsUnknown() {
		group, err = d.client.GetGroupByCN(data.CN.ValueString())
	} else {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'dn' or 'cn' must be specified to look up the group.",
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Map response to the data model
	data.ID = types.StringValue(group.DN)
	data.DN = types.StringValue(group.DN)
	data.CN = types.StringValue(group.CN)
	data.Name = types.StringValue(group.Name)
	data.SamAccountName = types.StringValue(group.SamAccountName)
	data.Description = types.StringValue(group.Description)
	
	if group.GroupType != "" {
		if groupTypeInt, err := strconv.ParseInt(group.GroupType, 10, 64); err == nil {
			data.GroupType = types.Int64Value(groupTypeInt)
		}
	}
	
	data.ManagedBy = types.StringValue(group.ManagedBy)
	data.ObjectGUID = types.StringValue(group.ObjectGUID)
	data.ObjectSid = types.StringValue(group.ObjectSid)

	// Convert members slice
	members := make([]types.String, len(group.Members))
	for i, member := range group.Members {
		members[i] = types.StringValue(member)
	}
	data.Members = members

	// Convert memberOf slice
	memberOf := make([]types.String, len(group.MemberOf))
	for i, member := range group.MemberOf {
		memberOf[i] = types.StringValue(member)
	}
	data.MemberOf = memberOf

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
