package flags

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// DynamicAddFlags menggunakan reflection untuk mendaftarkan flags Cobra dari struct.
// sourceStruct harus berupa pointer ke struct yang telah diisi dengan nilai default.
// Nilai default diambil langsung dari field struct.
func DynamicAddFlags(cmd *cobra.Command, target interface{}) error {
	val := reflect.ValueOf(target)

	// Periksa apakah yang dilewatkan adalah pointer (&)
	if val.Kind() != reflect.Ptr {
		return fmt.Errorf("DynamicAddFlags memerlukan pointer ke struct (gunakan '&')")
	}

	// Dapatkan nilai struct dari pointer
	val = val.Elem()
	typ := val.Type()

	// Panggil fungsi rekursif
	return addFlagsRecursive(cmd, val, typ)
}

func addFlagsRecursive(cmd *cobra.Command, val reflect.Value, typ reflect.Type) error {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Tangani Embedded/Nested Struct (Rekursif)
		if field.Anonymous || (fieldVal.Kind() == reflect.Struct && field.Type.Name() != "Time") {
			if fieldVal.CanAddr() {
				if err := addFlagsRecursive(cmd, fieldVal, fieldVal.Type()); err != nil {
					return err
				}
			}
			continue
		}

		flagName := field.Tag.Get("flag")
		shorthand := field.Tag.Get("shorthand")
		defaultTag := field.Tag.Get("default")
		if flagName == "" {
			continue
		}

		// Asumsi: Penggunaan "usage" telah diimplementasikan dengan benar
		usage := fmt.Sprintf("Option for %s", strings.ToLower(field.Name))

		// --- PARSE DEFAULT VALUE DARI TAG STRING ---
		defaultVal, err := parseTagDefault(defaultTag, field.Type.Kind())
		if err != nil {
			return fmt.Errorf("gagal parsing default tag '%s' untuk field %s: %w", defaultTag, field.Name, err)
		}

		// Pointer ke field struct untuk pendaftaran flag
		ptr := fieldVal.Addr().Interface()

		// PENTING: Gunakan fungsi NON-P (StringVar, IntVar, dll.)
		switch field.Type.Kind() {
		case reflect.String:
			cmd.Flags().StringP(flagName, shorthand, defaultVal.(string), usage)
		case reflect.Int:
			cmd.Flags().IntP(flagName, shorthand, defaultVal.(int), usage)
		case reflect.Bool:
			cmd.Flags().BoolP(flagName, shorthand, defaultVal.(bool), usage)
		case reflect.Slice:
			if field.Type.Elem().Kind() == reflect.String {
				defaultSlice := fieldVal.Interface().([]string)
				// Menggunakan StringSliceVar
				cmd.Flags().StringSliceVar(ptr.(*[]string), flagName, defaultSlice, usage)
			} else {
				return fmt.Errorf("tipe slice yang tidak didukung untuk flag %s: %s", flagName, field.Type)
			}
		default:
			return fmt.Errorf("tipe field yang tidak didukung untuk pendaftaran flag %s: %s", flagName, field.Type)
		}
	}
	return nil
}

// parseTagDefault mengkonversi string dari struct tag 'default' ke tipe data yang sesuai.
func parseTagDefault(tag string, kind reflect.Kind) (interface{}, error) {
	// Penanganan khusus untuk slice: jika tag kosong, kembalikan slice kosong
	if kind == reflect.Slice && tag == "" {
		return []string{}, nil
	}

	// Jika tag kosong, kembalikan nilai nol Go untuk tipe tersebut
	if tag == "" {
		// Mengembalikan nilai nol Go jika tag kosong
		switch kind {
		case reflect.String:
			return "", nil
		case reflect.Bool:
			return false, nil
		case reflect.Int:
			return 0, nil
		default:
			return nil, nil
		}
	}

	// Konversi berdasarkan jenis field
	switch kind {
	case reflect.String:
		return tag, nil
	case reflect.Bool:
		// Konversi string ke boolean
		val, err := strconv.ParseBool(tag)
		if err != nil {
			return nil, fmt.Errorf("boolean default tidak valid: %s", tag)
		}
		return val, nil
	case reflect.Int:
		// Konversi string ke integer
		val, err := strconv.Atoi(tag)
		if err != nil {
			return nil, fmt.Errorf("integer default tidak valid: %w", err)
		}
		return val, nil
	case reflect.Slice:
		// Diasumsikan format comma-separated string untuk []string
		slice := strings.Split(tag, ",")
		var cleanedSlice []string
		for _, s := range slice {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				cleanedSlice = append(cleanedSlice, trimmed)
			}
		}
		return cleanedSlice, nil

	default:
		return nil, fmt.Errorf("tipe field yang tidak didukung: %s", kind)
	}
}
