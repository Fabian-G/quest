package config

import (
	"log"
	"slices"

	"github.com/Fabian-G/quest/qselect"
	"github.com/spf13/viper"
)

var InternalEditTag = "quest-object-id"

func buildTagTypes(v *viper.Viper) map[string]qselect.DType {
	typeDefsConfig := v.GetStringMapString(TagsKey)
	typeDefs := map[string]qselect.DType{
		// Add internal tags
		InternalEditTag: qselect.QString,
	}

	for key, value := range typeDefsConfig {
		typ := qselect.DType(value)
		if !slices.Contains(qselect.AllDTypes, typ) {
			log.Fatalf("unknown type %s for tag key %s in config file", value, key)
		}
		typeDefs[key] = typ
	}
	return typeDefs
}
