package config

import (
	"log"

	"github.com/Fabian-G/quest/query"
	"github.com/spf13/viper"
)

func TagTypes() map[string]query.DType {
	typeDefsConfig := viper.GetStringMapString("tag.types")
	typeDefs := make(map[string]query.DType)

	for key, value := range typeDefsConfig {
		switch value {
		case "string":
			typeDefs[key] = query.QString
		case "date":
			typeDefs[key] = query.QDate
		case "duration":
			typeDefs[key] = query.QDuration
		case "int":
			typeDefs[key] = query.QInt
		default:
			log.Fatalf("unknown type %s for tag key %s in config file", value, key)
		}
	}
	return typeDefs
}
