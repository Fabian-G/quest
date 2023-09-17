package config

import (
	"log"
	"slices"

	"github.com/Fabian-G/quest/query"
	"github.com/spf13/viper"
)

func TagTypes() map[string]query.DType {
	typeDefsConfig := viper.GetStringMapString("tag.types")
	typeDefs := make(map[string]query.DType)

	for key, value := range typeDefsConfig {
		typ := query.DType(value)
		if !slices.Contains(query.AllDTypes, typ) {
			log.Fatalf("unknown type %s for tag key %s in config file", value, key)
		}
		typeDefs[key] = typ
	}
	return typeDefs
}
