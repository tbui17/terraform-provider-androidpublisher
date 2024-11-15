// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grant

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/tbui17/terraform-provider-androidpublisher/internal/lib"

	"google.golang.org/api/androidpublisher/v3"
)

func GrantsToTfModel(grants []*androidpublisher.Grant) basetypes.ListValue {
	res := make([]attr.Value, 0)
	for _, grant := range grants {
		model := TfModelFactory{grant}
		tfModel := model.GetTfModel()
		res = append(res, tfModel)

	}
	return types.ListValueMust(
		types.ObjectType{
			AttrTypes: Schema(),
		},
		res,
	)
}

type TfModelFactory struct {
	Grant *androidpublisher.Grant
}

func Schema() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                  types.StringType,
		"package_name":          types.StringType,
		"app_level_permissions": types.ListType{ElemType: types.StringType},
	}
}

func (g *TfModelFactory) GetModel() map[string]attr.Value {
	appPerms := lib.StrListToTfModel(g.Grant.AppLevelPermissions)
	return map[string]attr.Value{
		"name":                  types.StringValue(g.Grant.Name),
		"package_name":          types.StringValue(g.Grant.PackageName),
		"app_level_permissions": appPerms,
	}
}

func (g *TfModelFactory) GetTfModel() basetypes.ObjectValue {
	return types.ObjectValueMust(
		Schema(),
		g.GetModel(),
	)
}
