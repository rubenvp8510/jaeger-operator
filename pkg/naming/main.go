package naming

import (
	"fmt"
	"github.com/jaegertracing/jaeger-operator/internal/version"
	"github.com/spf13/viper"
	"strings"
)

// Image returns the image associated with the supplied image if defined, otherwise
// uses the parameter name to retrieve the value. If the parameter value does not
// include a tag/digest, the Jaeger version will be appended.
func Image(image, param string) string {
	if image == "" {
		param := viper.GetString(param)
		if strings.IndexByte(param, ':') == -1 {
			image = fmt.Sprintf("%s:%s", param, version.Get().Jaeger)
		} else {
			image = param
		}
	}
	return image
}
