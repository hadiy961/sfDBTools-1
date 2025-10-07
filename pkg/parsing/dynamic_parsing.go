// File : pkg/parsing/dynamic_parsing.go
// Deskripsi : Fungsi utilitas untuk parsing flags Cobra secara dinamis dari struct menggunakan reflection
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package parsing

import (
	"fmt"
	"reflect"
	"sfDBTools/pkg/common"

	"github.com/spf13/cobra"
)

// DynamicParseFlags mengiterasi struct target, membaca tags 'flag' dan 'env',
// dan mengisi nilai field menggunakan helper common.Get*FlagOrEnv.
func DynamicParseFlags(cmd *cobra.Command, target interface{}) error {
	// Pastikan target adalah pointer ke struct
	val := reflect.ValueOf(target).Elem()
	typ := val.Type()

	// Iterasi field struct
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Tangani Embedded/Nested Struct (Rekursif)
		if field.Anonymous || (fieldVal.Kind() == reflect.Struct && field.Type.Name() != "Time" /* hindari struct standar Go */) {
			// Jika field adalah embedded struct atau nested struct
			if fieldVal.Kind() == reflect.Struct && fieldVal.CanAddr() {
				if err := DynamicParseFlags(cmd, fieldVal.Addr().Interface()); err != nil {
					return err
				}
			}
			continue
		}

		// Baca tags 'flag' dan 'env'
		flagName := field.Tag.Get("flag")
		envName := field.Tag.Get("env")

		if flagName == "" {
			continue // Lewati field yang tidak memiliki tag 'flag'
		}

		// Dapatkan flag dari Cobra
		flag := cmd.Flag(flagName)
		if flag == nil {
			// Seharusnya tidak terjadi jika flags didaftarkan dengan benar di init()
			return fmt.Errorf("flag not registered: %s", flagName)
		}

		// Tentukan jenis field dan panggil helper yang sesuai
		switch field.Type.Kind() {
		case reflect.String:
			// Ambil nilai default COBRA (sudah termasuk nilai dari config/init)
			defaultVal := flag.Value.String()
			parsedVal := common.GetStringFlagOrEnv(cmd, flagName, envName, defaultVal)
			fieldVal.SetString(parsedVal)

		case reflect.Int:
			// Ambil nilai default COBRA
			defaultVal, _ := cmd.Flags().GetInt(flagName)
			parsedVal := common.GetIntFlagOrEnv(cmd, flagName, envName, defaultVal)
			fieldVal.SetInt(int64(parsedVal))

		case reflect.Bool:
			// Ambil nilai default COBRA
			defaultVal, _ := cmd.Flags().GetBool(flagName)
			parsedVal := common.GetBoolFlagOrEnv(cmd, flagName, envName, defaultVal)
			fieldVal.SetBool(parsedVal)

		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				// Ambil nilai default COBRA
				defaultVal, _ := cmd.Flags().GetStringSlice(flagName)
				parsedVal := common.GetStringSliceFlagOrEnv(cmd, flagName, envName, defaultVal)

				// Assign slice
				sliceVal := reflect.MakeSlice(field.Type, len(parsedVal), len(parsedVal))
				for k, v := range parsedVal {
					sliceVal.Index(k).SetString(v)
				}
				fieldVal.Set(sliceVal)
			} else {
				return fmt.Errorf("unsupported slice type for flag %s: %s", flagName, field.Type)
			}
		default:
			return fmt.Errorf("unsupported field type for flag %s: %s", flagName, field.Type)
		}
	}
	return nil
}
