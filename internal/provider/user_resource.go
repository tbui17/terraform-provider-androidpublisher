// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/tbui17/terraform-provider-androidpublisher/internal/grant"
	"github.com/tbui17/terraform-provider-androidpublisher/internal/lib"

	"google.golang.org/api/androidpublisher/v3"
	"net/http"
)

func (m *UserResourceModel) SetFromUser(ctx context.Context, user androidpublisher.User) {
	m.AccessState = types.StringValue(user.AccessState)
	m.Name = types.StringValue(user.Name)
	m.Email = types.StringValue(user.Email)

	m.Grants = grant.GrantsToTfModel(user.Grants)
}

func (m *UserResourceModel) GetParent() string {
	return "developers/" + m.DeveloperID.ValueString()
}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UserResource{}

// UserResource defines the resource implementation.
type UserResource struct {
	client                  *http.Client
	androidPublisherService *androidpublisher.Service
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	AccessState                 types.String `tfsdk:"access_state"`
	DeveloperID                 types.String `tfsdk:"developer_id"`
	Email                       types.String `tfsdk:"email"`
	ExpirationTime              types.String `tfsdk:"expiration_time"`
	Grants                      types.List   `tfsdk:"grants"`
	Name                        types.String `tfsdk:"name"`
	DeveloperAccountPermissions types.List   `tfsdk:"developer_account_permissions"`
}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Google Play Android Publisher User resource https://developers.google.com/android-publisher/api-ref/rest/v3/users",

		Attributes: map[string]schema.Attribute{
			"developer_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the developer account",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The user's email address",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"developer_account_permissions": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "The list of permissions granted to the user",
			},
			"expiration_time": schema.StringAttribute{
				MarkdownDescription: "The time at which the user's access expires",
				Optional:            true,
			},
			"access_state": schema.StringAttribute{
				MarkdownDescription: "The state of the user's access to the Play Console",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Resource name for this user, following the pattern \"developers/{developer}/ users/{email}\".",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"grants": schema.ListNestedAttribute{
				MarkdownDescription: "The list of grants for the user",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the grant",
							Computed:            true,
						},
						"package_name": schema.StringAttribute{
							MarkdownDescription: "The package name of the app for which the user has access",
							Computed:            true,
						},
						"app_level_permissions": schema.ListAttribute{
							MarkdownDescription: "The list of app-level permissions granted to the user",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	service, err := androidpublisher.NewService(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Android Publisher client",
			fmt.Sprintf("Unable to create Android Publisher client: %v", err),
		)
		return
	}

	r.androidPublisherService = service
	r.client = client

}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permissions, diags := lib.TFListToList[string](ctx, data.DeveloperAccountPermissions)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	user := &androidpublisher.User{
		Email:                       data.Email.ValueString(),
		DeveloperAccountPermissions: permissions,
		Name:                        lib.GetName(data.Email.ValueString(), data.DeveloperID.ValueString()),
		ExpirationTime:              data.ExpirationTime.ValueString(),
	}

	parent := data.GetParent()

	request := r.androidPublisherService.Users.Create(parent, user)

	usr, err := request.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error creating user", fmt.Sprintf("Unable to create user: %v", err))
		return
	}

	data.SetFromUser(ctx, *usr)

	tflog.Trace(ctx, "created a user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.GetUser(data)
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", fmt.Sprintf("Unable to read user: %v", err))
		return
	}

	if result == nil {
		resp.Diagnostics.AddError("Could not find user", fmt.Sprintf("Could not find user with provided params %v", data))
		return
	}
	data.SetFromUser(ctx, *result)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) GetUser(data UserResourceModel) (*androidpublisher.User, error) {
	request := r.androidPublisherService.Users.List(data.GetParent()).PageSize(-1)

	response, err := request.Do()
	if err != nil {

		return nil, err
	}

	var result *androidpublisher.User

	emailStr := data.Email.ValueString()
	for _, user := range response.Users {
		if user.Email == emailStr {
			result = user
			break
		}
	}
	return result, nil
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	permissions, diags := lib.TFListToList[string](ctx, data.DeveloperAccountPermissions)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	user := &androidpublisher.User{
		DeveloperAccountPermissions: permissions,
		ExpirationTime:              data.ExpirationTime.ValueString(),
	}

	userName := lib.GetName(data.Email.ValueString(), data.DeveloperID.ValueString())
	updateFields := "developerAccountPermissions,expirationTime"
	request := r.androidPublisherService.Users.Patch(userName, user).UpdateMask(updateFields)
	usr, err := request.Do()
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", fmt.Sprintf("Unable to update user: %v", err))
		return
	}

	data.SetFromUser(ctx, *usr)

	tflog.Trace(ctx, "created a user resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.androidPublisherService.Users.Delete(data.Name.ValueString()).Do()
	if err != nil {
		resp.Diagnostics.AddError("Error deleting user", fmt.Sprintf("Unable to delete user: %v", err))
		return
	}

}
