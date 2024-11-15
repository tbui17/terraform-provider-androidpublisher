// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lib

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func DeveloperIDToParentFragment(developerID string) string {
	return "developers/" + developerID
}

func TFListToList[T any](ctx context.Context, list types.List) ([]T, diag.Diagnostics) {
	var slice []T
	diags := list.ElementsAs(ctx, &slice, true)
	if diags.HasError() {
		return nil, diags
	}
	return slice, nil
}

func StrListToTfModel(strList []string) basetypes.ListValue {
	var res []attr.Value
	for _, str := range strList {
		res = append(res, types.StringValue(str))
	}
	return types.ListValueMust(types.StringType, res)
}

func GetName(userEmail string, developerId string) string {
	return "developers/" + developerId + "/users/" + userEmail
}
