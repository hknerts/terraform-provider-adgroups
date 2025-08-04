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
var _ datasource.DataSource = &GroupsDataSource{}

func NewGroupsDataSource() datasource.DataSource {
	return &GroupsDataSource{}
}

// GroupsDataSource defines the data source implementation.
type GroupsDataSource struct {
	client *client.Client
}

// GroupsDataSourceModel describes the data source data model.
type GroupsDataSourceModel struct {
	ID     types.String                   `tfsdk:"id"`
	Filter types.String                   `tfsdk:"filter"`
	Groups []GroupsDataSourceGroupModel `tfsdk:"groups"`
}

type GroupsDataSourceGroupModel struct {
	DN             types.String `tfsdk:"dn"`
	CN             types.String `tfsdk:"cn"`
	Name           types.String `tfsdk:"name"`
	SamAccountName types.String `tfsdk:"sam_account_name"`
	Description    types.String `tfsdk:"description"`
	GroupType      types.Int64  `tfsdk:"group_type"`
	ManagedBy      types.String `tfsdk:"managed_by"`
	ObjectGUID     types.String `tfsdk:"object_guid"`
	ObjectSid      types.String `tfsdk:"object_sid"`
}

func (d *GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of Active Directory groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for this data source.",
			},
			"filter": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "LDAP filter to apply when searching for groups. Defaults to '(objectClass=group)'.",
			},
			"groups": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of groups matching the filter.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"dn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Distinguished Name of the group.",
						},
						"cn": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Common Name of the group.",
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
						"object_guid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The objectGUID of the group.",
						},
						"object_sid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The objectSid of the group.",
						},
					},
				},
			},
		},
	}
}

func (d *GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	filter := "(objectClass=group)"
	if !data.Filter.IsNull() && !data.Filter.IsUnknown() {
		filter = data.Filter.ValueString()
	}

	groups, err := d.client.ListGroups(filter)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list groups, got error: %s", err))
		return
	}

	// Map response to the data model
	data.ID = types.StringValue("groups")
	data.Filter = types.StringValue(filter)

	groupModels := make([]GroupsDataSourceGroupModel, len(groups))
	for i, group := range groups {
		groupModels[i] = GroupsDataSourceGroupModel{
			DN:             types.StringValue(group.DN),
			CN:             types.StringValue(group.CN),
			Name:           types.StringValue(group.Name),
			SamAccountName: types.StringValue(group.SamAccountName),
			Description:    types.StringValue(group.Description),
			ManagedBy:      types.StringValue(group.ManagedBy),
			ObjectGUID:     types.StringValue(group.ObjectGUID),
			ObjectSid:      types.StringValue(group.ObjectSid),
		}
		
		// Parse group type if available
		if group.GroupType != "" {
			if groupTypeInt, parseErr := fmt.Sscanf(group.GroupType, "%d", new(int64)); parseErr == nil && groupTypeInt == 1 {
				groupModels[i].GroupType = types.Int64Value(int64(groupTypeInt))
			}
		}
	}
	data.Groups = groupModels

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
