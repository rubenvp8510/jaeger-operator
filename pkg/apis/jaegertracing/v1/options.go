package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Values represent and string or a slice of string
type Values map[string]interface{}

// DeepCopy indicate how to do a deep copy of Values type
func (v *Values) DeepCopy() *Values {
	out := make(Values, len(*v))
	for key, val := range *v {
		switch val.(type) {
		case string:
			out[key] = val

		case []string:
			out[key] = append([]string(nil), val.([]string)...)
		}
	}
	return &out
}

// Options defines a common options parameter to the different structs
type Options struct {
	opts Values  `json:"-"`
	json *[]byte `json:"-"`
}

// NewOptions build a new Options object based on the given map
func NewOptions(o map[string]interface{}) Options {
	options := Options{}
	options.parse(o)
	return options
}

// Filter creates a new Options object with just the elements identified by the supplied prefix
func (o *Options) Filter(prefix string) Options {
	options := Options{}
	options.opts = make(map[string]interface{})

	archivePrefix := prefix + "-archive."
	prefix += "."

	for k, v := range o.opts {
		if strings.HasPrefix(k, prefix) || strings.HasPrefix(k, archivePrefix) {
			options.opts[k] = v
		}
	}

	return options
}

// UnmarshalJSON implements an alternative parser for this field
func (o *Options) UnmarshalJSON(b []byte) error {
	var entries map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if err := d.Decode(&entries); err != nil {
		return err
	}
	o.parse(entries)
	o.json = &b
	return nil
}

// MarshalJSON specifies how to convert this object into JSON
func (o Options) MarshalJSON() ([]byte, error) {
	if nil != o.json {
		return *o.json, nil
	}

	if len(o.opts) == 0 {
		return []byte("{}"), nil
	}

	if len(o.opts) > 0 {
		return json.Marshal(o.opts)
	}

	return *o.json, nil
}

func (o *Options) parse(entries map[string]interface{}) {
	o.opts = make(map[string]interface{})
	for k, v := range entries {
		o.opts = entry(o.opts, k, v)
	}
}

func entry(entries map[string]interface{}, key string, value interface{}) map[string]interface{} {
	switch value.(type) {
	case map[string]interface{}:
		for k, v := range value.(map[string]interface{}) {
			entries = entry(entries, fmt.Sprintf("%s.%v", key, k), v)
		}
	case []interface{}:
		values := make([]string, 0, len(value.([]interface{})))
		for _, v := range value.([]interface{}) {
			values = append(values, v.(string))
		}
		entries[key] = values
	case interface{}:
		entries[key] = fmt.Sprintf("%v", value)
	}

	return entries
}

// ToArgs converts the options to a value suitable for the Container.Args field
func (o *Options) ToArgs() []string {
	if len(o.opts) > 0 {
		args := make([]string, 0, len(o.opts))
		for k, v := range o.opts {
			switch v.(type) {
			case string:
				args = append(args, fmt.Sprintf("--%s=%v", k, v))
			case []string:
				for _, vv := range v.([]string) {
					args = append(args, fmt.Sprintf("--%s=%v", k, vv))
				}
			}
		}
		return args
	}
	return nil
}

// Map returns a map representing the option entries. Items are flattened, with dots as separators. For instance
// an option "cassandra" with a nested "servers" object becomes an entry with the key "cassandra.servers"
func (o *Options) Map() map[string]interface{} {
	return o.opts
}

// GenericMap returns the map representing the option entries as interface{}, suitable for usage with NewOptions()
func (o *Options) GenericMap() map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range o.opts {
		out[k] = v
	}
	return out
}
