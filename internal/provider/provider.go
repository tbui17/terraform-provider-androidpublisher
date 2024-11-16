// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"google.golang.org/api/androidpublisher/v3"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure GoogleProvider satisfies various provider interfaces.
var _ provider.Provider = &GoogleProvider{}

// GoogleProvider defines the provider implementation.
type GoogleProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GoogleProviderModel describes the provider data model.
type GoogleProviderModel struct {
}

type GoogleProviderContext struct {
	Client                  *http.Client
	AndroidPublisherService *androidpublisher.Service
}

func (p *GoogleProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "androidpublisher"
	resp.Version = p.version
}

func (p *GoogleProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interacts with Google Play Developer APIs. https://developers.google.com/android-publisher",
	}
}

func (p *GoogleProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GoogleProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	service, err := androidpublisher.NewService(ctx)
	if err != nil {
		resp.Diagnostics.AddError("error creating Android Publisher service: %s", err.Error())
		return
	}

	providerContext := &GoogleProviderContext{
		Client:                  http.DefaultClient,
		AndroidPublisherService: service,
	}

	resp.DataSourceData = providerContext
	resp.ResourceData = providerContext

}

func (p *GoogleProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
	}
}

func (p *GoogleProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GoogleProvider{
			version: version,
		}
	}
}
