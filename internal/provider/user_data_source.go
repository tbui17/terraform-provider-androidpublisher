// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"github.com/tbui17/terraform-provider-androidpublisher/internal/grant"
	"github.com/tbui17/terraform-provider-androidpublisher/internal/lib"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

// UserDataSource defines the data source implementation.
type UserDataSource struct {
	*GoogleProviderContext
}

type UserData struct {
	AccessState                 types.String `tfsdk:"access_state"`
	Email                       types.String `tfsdk:"email"`
	ExpirationTime              types.String `tfsdk:"expiration_time"`
	Grants                      types.List   `tfsdk:"grants"`
	Name                        types.String `tfsdk:"name"`
	DeveloperAccountPermissions types.List   `tfsdk:"developer_account_permissions"`
}

// UserDataModel describes the resource data model.
type UserDataModel struct {
	DeveloperID types.String `tfsdk:"developer_id"`
	Value       []UserData   `tfsdk:"value"`
}

func (m *UserDataModel) GetDeveloperIdFragment() string {
	return lib.DeveloperIDToParentFragment(m.DeveloperID.ValueString())
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{

		MarkdownDescription: "Retrieves a list of users associated with a developer account. Maps to the https://developers.google.com/android-publisher/api-ref/rest/v3/users/list endpoint.",

		Attributes: map[string]schema.Attribute{
			"developer_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the developer account",
				Required:            true,
			},

			"value": schema.ListNestedAttribute{
				MarkdownDescription: "The list of users",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"email": schema.StringAttribute{
							MarkdownDescription: "The user's email address",
							Computed:            true,
						},
						"developer_account_permissions": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "The list of permissions granted to the user",
						},
						"expiration_time": schema.StringAttribute{
							MarkdownDescription: "The time at which the user's access expires",
							Computed:            true,
						},
						"access_state": schema.StringAttribute{
							MarkdownDescription: "The state of the user's access to the Play Console",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Resource name for this user, following the pattern \"developers/{developer}/ users/{email}\".",
							Computed:            true,
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
				},
				Computed: true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	gCtx, ok := req.ProviderData.(*GoogleProviderContext)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.GoogleProviderContext = gCtx
}

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataModel

	// Read Terraform configuration data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := d.AndroidPublisherService.Users.List(data.GetDeveloperIdFragment())
	usersResponse, err := request.PageSize(-1).Do()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list users", err.Error())
		return
	}
	var userDataEntries []UserData
	for _, user := range usersResponse.Users {
		userData := UserToUserData(*user)
		userDataEntries = append(userDataEntries, userData)
	}

	data.Value = userDataEntries

	// Write logs using the tflog package.
	// Documentation: https://terraform.io/plugin/log.
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func UserToUserData(user androidpublisher.User) UserData {
	return UserData{
		AccessState:                 types.StringValue(user.AccessState),
		Email:                       types.StringValue(user.Email),
		ExpirationTime:              types.StringValue(user.ExpirationTime),
		Name:                        types.StringValue(user.Name),
		DeveloperAccountPermissions: lib.StrListToTfModel(user.DeveloperAccountPermissions),
		Grants:                      grant.GrantsToTfModel(user.Grants),
	}
}
