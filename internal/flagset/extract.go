// Copyright (C) The AetherFS Authors - All Rights Reserved
// See LICENSE for more information.

package flagset

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
)

func format(prefix []string, name string) string {
	envVar := name
	for i := 1; i <= len(prefix); i++ {
		envVar = prefix[len(prefix)-i] + "_" + envVar
	}
	return envVar
}

func extract(prefix []string, value reflect.Value) []cli.Flag {
	flags := make([]cli.Flag, 0)

	for i := 0; i < value.NumField(); i++ {
		fieldValue := reflect.Indirect(value.Field(i))
		field := value.Type().Field(i)

		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}

		// all other data types
		var aliases []string
		if alias := field.Tag.Get("aliases"); alias != "" {
			aliases = strings.Split(alias, ",")
		}

		flagName := format(prefix, name)
		defaultTag := field.Tag.Get("default")

		var err error
		switch field.Type.Kind() {
		case reflect.Struct, reflect.Ptr:
			pre := prefix
			if name != "" {
				pre = append(pre, name)
			}

			flags = append(flags, extract(pre, fieldValue)...)
		case reflect.String:
			flags = append(flags, &cli.StringFlag{
				Name:        flagName,
				Aliases:     aliases,
				Usage:       field.Tag.Get("usage"),
				EnvVars:     []string{strings.ToUpper(flagName)},
				Destination: fieldValue.Addr().Interface().(*string),
				Value:       defaultTag,
			})
		case reflect.Int:
			i := 0
			if defaultTag != "" {
				i, err = strconv.Atoi(defaultTag)
				if err != nil {
					panic(fmt.Sprintf("invalid int default: %s", defaultTag))
				}
			}

			flags = append(flags, &cli.IntFlag{
				Name:        flagName,
				Aliases:     aliases,
				Usage:       field.Tag.Get("usage"),
				EnvVars:     []string{strings.ToUpper(flagName)},
				Destination: fieldValue.Addr().Interface().(*int),
				Value:       i,
			})
		case reflect.Bool:
			b := false
			if defaultTag != "" {
				b, err = strconv.ParseBool(defaultTag)
				if err != nil {
					panic(fmt.Sprintf("invalid bool default: %s", field.Tag.Get("default")))
				}
			}

			flags = append(flags, &cli.BoolFlag{
				Name:        flagName,
				Aliases:     aliases,
				Usage:       field.Tag.Get("usage"),
				EnvVars:     []string{strings.ToUpper(flagName)},
				Destination: fieldValue.Addr().Interface().(*bool),
				Value:       b,
			})
		}
	}

	return flags
}

// Extract parses the provided object to create a flagset.
func Extract(v interface{}) []cli.Flag {
	return extract([]string{}, reflect.Indirect(reflect.ValueOf(v)))
}
