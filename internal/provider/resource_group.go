package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hknerts/terraform-provider-adgroups/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GroupResource{}
var _ resource.ResourceWithImportState = &GroupResource{}

func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource defines the resource implementation.
type GroupResource struct {
	client *client.Client
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	ID             types.String `tfsdk:"id"`
	DN             types.String `tfsdk:"dn"`
	CN             types.String `tfsdk:"cn"`
	Name           types.String `tfsdk:"name"`
	SamAccountName types.String `tfsdk:"sam_account_name"`
	Description    types.String `tfsdk:"description"`
	GroupType      types.Int64  `tfsdk:"group_type"`
	ManagedBy      types.String `tfsdk:"managed_by"`
	OU             types.String `tfsdk:"ou"`
	ObjectGUID     types.String `tfsdk:"object_guid"`
	ObjectSid      types.String `tfsdk:"object_sid"`
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an Active Directory group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the group (same as DN).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Distinguished Name of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Common Name of the group. This will be used as the sAMAccountName if not specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Display name of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sam_account_name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Security Account Manager (SAM) account name. Defaults to CN if not specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the group.",
			},
			"group_type": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Group type. Common values: 2 (Global Distribution), 4 (Domain Local Distribution), 8 (Universal Distribution), -2147483646 (Global Security), -2147483644 (Domain Local Security), -2147483640 (Universal Security). Defaults to -2147483646 (Global Security).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"managed_by": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Distinguished Name of the user or group that manages this group.",
			},
			"ou": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Organizational Unit where the group will be created (e.g., 'OU=Groups,DC=example,DC=com').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_guid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The objectGUID of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"object_sid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The objectSid of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set default group type if not specified
	groupType := int(-2147483646) // Global Security Group
	if !data.GroupType.IsNull() && !data.GroupType.IsUnknown() {
		groupType = int(data.GroupType.ValueInt64())
	}

	// Create the group
	group, err := r.client.CreateGroup(
		data.OU.ValueString(),
		data.CN.ValueString(),
		data.Description.ValueString(),
		groupType,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create group, got error: %s", err))
		return
	}

	// Update managedBy if specified
	if !data.ManagedBy.IsNull() && !data.ManagedBy.IsUnknown() {
		updates := map[string][]string{
			"managedBy": {data.ManagedBy.ValueString()},
		}
		err = r.client.UpdateGroup(group.DN, updates)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update group managedBy, got error: %s", err))
			return
		}
		
		// Refresh group data
		group, err = r.client.GetGroup(group.DN)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group after update, got error: %s", err))
			return
		}
	}

	// Map response body to schema and populate Computed attribute values
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

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get the group from AD
	group, err := r.client.GetGroup(data.DN.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Group was deleted outside of Terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Update the model with current values
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	var state GroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare updates
	updates := make(map[string][]string)

	if !data.Description.Equal(state.Description) {
		if data.Description.IsNull() {
			updates["description"] = []string{}
		} else {
			updates["description"] = []string{data.Description.ValueString()}
		}
	}

	if !data.ManagedBy.Equal(state.ManagedBy) {
		if data.ManagedBy.IsNull() {
			updates["managedBy"] = []string{}
		} else {
			updates["managedBy"] = []string{data.ManagedBy.ValueString()}
		}
	}

	if !data.GroupType.Equal(state.GroupType) {
		updates["groupType"] = []string{fmt.Sprintf("%d", data.GroupType.ValueInt64())}
	}

	// Apply updates if any
	if len(updates) > 0 {
		err := r.client.UpdateGroup(data.DN.ValueString(), updates)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update group, got error: %s", err))
			return
		}
	}

	// Read the updated group
	group, err := r.client.GetGroup(data.DN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group after update, got error: %s", err))
		return
	}

	// Update the model
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

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the group
	err := r.client.DeleteGroup(data.DN.ValueString())
	if err != nil {
		if !strings.Contains(err.Error(), "not found") {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete group, got error: %s", err))
			return
		}
		// If group is already deleted, that's fine
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by DN
	resource.ImportStatePassthroughID(ctx, path.Root("dn"), req, resp)
	
	// Set ID to the same value as DN
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
