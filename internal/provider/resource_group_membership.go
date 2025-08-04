package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hknerts/terraform-provider-adgroups/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &GroupMembershipResource{}
var _ resource.ResourceWithImportState = &GroupMembershipResource{}

func NewGroupMembershipResource() resource.Resource {
	return &GroupMembershipResource{}
}

// GroupMembershipResource defines the resource implementation.
type GroupMembershipResource struct {
	client *client.Client
}

// GroupMembershipResourceModel describes the resource data model.
type GroupMembershipResourceModel struct {
	ID       types.String `tfsdk:"id"`
	GroupDN  types.String `tfsdk:"group_dn"`
	MemberDN types.String `tfsdk:"member_dn"`
}

func (r *GroupMembershipResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group_membership"
}

func (r *GroupMembershipResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages membership of a user or group in an Active Directory group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the group membership (format: groupDN|memberDN).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_dn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Distinguished Name of the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"member_dn": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Distinguished Name of the member (user or group) to add to the group.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *GroupMembershipResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GroupMembershipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupMembershipResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groupDN := data.GroupDN.ValueString()
	memberDN := data.MemberDN.ValueString()

	// Check if the group exists
	_, err := r.client.GetGroup(groupDN)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to find group %s, got error: %s", groupDN, err))
		return
	}

	// Add member to group
	err = r.client.AddMemberToGroup(groupDN, memberDN)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add member to group, got error: %s", err))
		return
	}

	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s|%s", groupDN, memberDN))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMembershipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupMembershipResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groupDN := data.GroupDN.ValueString()
	memberDN := data.MemberDN.ValueString()

	// Get the group from AD
	group, err := r.client.GetGroup(groupDN)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Group was deleted outside of Terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read group, got error: %s", err))
		return
	}

	// Check if the member is still in the group
	memberFound := false
	for _, member := range group.Members {
		if strings.EqualFold(member, memberDN) {
			memberFound = true
			break
		}
	}

	if !memberFound {
		// Member was removed outside of Terraform
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupMembershipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Group membership doesn't support updates - any change requires replacement
	// This is handled by the schema with RequiresReplace plan modifiers
}

func (r *GroupMembershipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupMembershipResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	groupDN := data.GroupDN.ValueString()
	memberDN := data.MemberDN.ValueString()

	// Remove member from group
	err := r.client.RemoveMemberFromGroup(groupDN, memberDN)
	if err != nil {
		// Check if it's because the group or member doesn't exist
		if strings.Contains(err.Error(), "not found") || 
		   strings.Contains(err.Error(), "does not exist") ||
		   strings.Contains(err.Error(), "no such object") {
			// Already removed, that's fine
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove member from group, got error: %s", err))
		return
	}
}

func (r *GroupMembershipResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "groupDN|memberDN"
	parts := strings.Split(req.ID, "|")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: groupDN|memberDN. Got: %q", req.ID),
		)
		return
	}

	groupDN := parts[0]
	memberDN := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_dn"), groupDN)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member_dn"), memberDN)...)
}
