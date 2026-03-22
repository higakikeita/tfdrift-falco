package falco

import (
	"github.com/keitahigaki/tfdrift-falco/pkg/util"
)

// getStringField safely gets a string field from the map
func getStringField(fields map[string]string, key string) string {
	return util.GetStringField(fields, key)
}
