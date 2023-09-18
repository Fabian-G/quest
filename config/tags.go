package config

import (
	"log"
	"slices"

	"github.com/Fabian-G/quest/qselect"
	"github.com/spf13/viper"
)

func TagTypes() map[string]qselect.DType {
	typeDefsConfig := viper.GetStringMapString("tag.types")
	typeDefs := make(map[string]qselect.DType)

	for key, value := range typeDefsConfig {
		typ := qselect.DType(value)
		if !slices.Contains(qselect.AllDTypes, typ) {
			log.Fatalf("unknown type %s for tag key %s in config file", value, key)
		}
		typeDefs[key] = typ
	}
	return typeDefs
}
