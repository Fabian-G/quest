package config

import (
	"log"
	"slices"

	"github.com/Fabian-G/quest/qselect"
	"github.com/spf13/viper"
)

var InternalEditTag = "quest-object-id"

func buildTagTypes(v *viper.Viper) map[string]qselect.DType {
	tagDefsConfig := v.GetStringMap(TagsKey)
	typeDefs := map[string]qselect.DType{
		// Add internal tags
		InternalEditTag: qselect.QString,
	}

	for key := range tagDefsConfig {
		sub := v.Sub(TagsKey + "." + key)
		typ := qselect.DType(sub.GetString("type"))
		if !slices.Contains(qselect.AllDTypes, typ) {
			log.Fatalf("unknown type %s for tag key %s in config file", typ, key)
		}
		typeDefs[key] = typ
	}
	return typeDefs
}

func buildHumanizedTags(v *viper.Viper) []string {
	humanizedTags := make([]string, 0)

	tagDefsConfig := v.GetStringMap(TagsKey)
	for key := range tagDefsConfig {
		if v.GetBool(TagsKey + "." + key + ".humanize") {
			humanizedTags = append(humanizedTags, key)
		}
	}
	return humanizedTags
}
