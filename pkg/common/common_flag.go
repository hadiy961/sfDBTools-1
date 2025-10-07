package common

import (
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func GetStringFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal string) string {
	val, _ := cmd.Flags().GetString(flagName)
	if val != "" {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		return env
	}
	return defaultVal
}

func GetIntFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal int) int {
	val, _ := cmd.Flags().GetInt(flagName)
	if val != 0 {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		// ignore error, fallback ke default jika gagal
		if i, err := strconv.Atoi(env); err == nil {
			return i
		}
	}
	return defaultVal
}

func GetBoolFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal bool) bool {
	val, _ := cmd.Flags().GetBool(flagName)
	// Cobra default: false jika tidak di-set, jadi cek ENV jika flag tidak di-set
	if cmd.Flags().Changed(flagName) {
		return val
	}
	env := os.Getenv(envName)
	if env != "" {
		env = strings.ToLower(env)
		return env == "1" || env == "true" || env == "yes"
	}
	return defaultVal
}

func GetStringSliceFlagOrEnv(cmd *cobra.Command, flagName, envName string, defaultVal []string) []string {
	// 1. Coba ambil nilai dari CLI Flag
	// Ignore error as we assume the flag is correctly registered.
	val, _ := cmd.Flags().GetStringSlice(flagName)

	// Jika flag diubah secara eksplisit, kembalikan nilai flag.
	if cmd.Flags().Changed(flagName) {
		return val
	}

	// 2. Cek Environment Variable (diasumsikan format comma-separated)
	env := os.Getenv(envName)
	if env != "" {
		// Pisahkan string ENV berdasarkan koma
		slice := strings.Split(env, ",")

		var cleanedSlice []string
		// Bersihkan spasi dan filter elemen kosong
		for _, s := range slice {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				cleanedSlice = append(cleanedSlice, trimmed)
			}
		}

		if len(cleanedSlice) > 0 {
			return cleanedSlice
		}
	}

	// 3. Kembalikan default.
	return defaultVal
}
