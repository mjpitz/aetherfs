// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package flagset

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

type ref struct {
	name string
	t    reflect.Type
	v    reflect.Value
}

func (r *ref) Set(value string) error {
	if r.v.CanInterface() {
		switch v := r.v.Interface().(type) {
		case time.Duration:
			duration, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			r.v.Set(reflect.ValueOf(duration))
			return nil
		case cli.Generic:
			return v.Set(value)
		}
	}

	switch r.t.Kind() {
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		r.v.SetBool(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		r.v.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		r.v.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		r.v.SetFloat(v)

	case reflect.String:
		r.v.SetString(value)

	default:
		return fmt.Errorf("unsupported kind: %s", r.t.Kind().String())

	}

	return nil
}

func (r *ref) String() string {
	if r.v.CanInterface() {
		switch v := r.v.Interface().(type) {
		case time.Duration:
			return v.String()
		case cli.Generic:
			return v.String()
		}
	}

	switch r.t.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(r.v.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(r.v.Int(), 10)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(r.v.Uint(), 10)

	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(r.v.Float(), 'f', 7, 64)

	}

	return r.v.String()
}

var _ cli.Generic = &ref{}

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
		fieldValue := value.Field(i)
		field := value.Type().Field(i)

		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "-" {
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		// recursive field types
		if fieldValue.Kind() == reflect.Struct {
			pre := prefix
			if name != "" {
				pre = append(pre, name)
			}

			flags = append(flags, extract(pre, fieldValue)...)
			continue
		}

		// all other data types
		var aliases []string
		if alias := field.Tag.Get("aliases"); alias != "" {
			aliases = strings.Split(alias, ",")
		}

		flagName := format(prefix, name)

		flags = append(flags, &cli.GenericFlag{
			Name:    flagName,
			Aliases: aliases,
			Usage:   field.Tag.Get("usage"),
			EnvVars: []string{strings.ToUpper(flagName)},
			Value: &ref{
				name: flagName,
				t:    field.Type,
				v:    fieldValue,
			},
		})
	}

	return flags
}

// Extract parses the provided object to create a flagset.
func Extract(v interface{}) []cli.Flag {
	value := reflect.ValueOf(v)

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	return extract([]string{}, value)
}
