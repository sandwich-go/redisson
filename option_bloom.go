package redisson

//go:generate optiongen --option_with_struct_name=false --new_func=newBloomOptions --empty_composite_nil=true --usage_tag_name=usage
func BloomOptionsOptionDeclareWithDefault() any {
	return map[string]any{
		// annotation@KeyPrefix(If enabled, Exists and ExistsMulti methods will be available as read-only operations. NOTE: If enabled, minimum redis version should be 7.0.0.)
		"EnableReadOperation": false,
	}
}
