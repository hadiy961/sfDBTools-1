// File : internal/applog/logger_loader.go
// Deskripsi : Fungsi untuk memuat dan menginisialisasi logger aplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package applog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	config "sfDBTools/internal/appconfig"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	// Ganti dengan path modul Go Anda yang sebenarnya
)

// Logger adalah tipe alias untuk logrus.Logger
type Logger = *logrus.Logger

// Field represents a structured log field
type Field struct {
	Key   string
	Value interface{}
}

// Common field constructors
func String(key, val string) Field            { return Field{Key: key, Value: val} }
func Strings(key string, vals []string) Field { return Field{Key: key, Value: vals} }
func Int(key string, val int) Field           { return Field{Key: key, Value: val} }
func Int64(key string, val int64) Field       { return Field{Key: key, Value: val} }
func Float64(key string, val float64) Field   { return Field{Key: key, Value: val} }
func Bool(key string, val bool) Field         { return Field{Key: key, Value: val} }
func Error(err error) Field                   { return Field{Key: "error", Value: err} }

// Time returns a Field containing time.Time value for structured logging
func Time(key string, t time.Time) Field {
	return Field{Key: key, Value: t}
}

// NewLogger menginisialisasi dan mengembalikan logger yang sudah dikonfigurasi.
func NewLogger() Logger {
	// inisiasi configurasi logger
	LoadConfig, err := config.LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}
	cfg := LoadConfig.Log

	// 0. Inisialisasi dasar logrus
	log := logrus.New()

	// ** PENTING ** Kita harus selalu mengaktifkan caller untuk ConditionalCallerHook
	log.SetReportCaller(true)

	// 1. Set Level Logging & Dapatkan ConfiguredLevel
	configuredLevel, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		configuredLevel = logrus.InfoLevel
		log.SetLevel(configuredLevel)
		log.Warnf("Level log '%s' tidak valid. Menggunakan default: info", cfg.Level)
	} else {
		log.SetLevel(configuredLevel)
	}

	// 2. Load Timezone
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		loc = time.Local
		log.Warnf("Timezone '%s' tidak valid. Menggunakan Timezone Lokal.", cfg.Timezone)
	}

	// 3. Set Format Logging
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	} else {
		// Menggunakan CustomTextFormatter untuk format spesifik [timestamp][level][caller] - Message
		log.SetFormatter(&CustomTextFormatter{
			Location: loc,
		})
	}

	// 4. Tambahkan Hooks
	// Hook Kustom untuk Caller (mengimplementasikan logika conditional caller)
	log.AddHook(&ConditionalCallerHook{
		ConfiguredLevel: configuredLevel, // Teruskan level konfigurasi
	})

	// Timezone Hook, hanya diperlukan jika JSON formatter digunakan
	if cfg.Format == "json" {
		log.AddHook(&TimezoneHook{Location: loc})
	}

	// 5. Konfigurasi Output
	var writers []io.Writer

	// A. Output ke File (Rotasi)
	if cfg.Output.File.Enabled {
		fileCfg := cfg.Output.File
		// Pastikan direktori ada
		if _, err := os.Stat(fileCfg.Dir); os.IsNotExist(err) {
			os.MkdirAll(fileCfg.Dir, 0755)
		}

		// Tentukan nama file dengan pola dan tanggal jika perlu
		logFilename := fileCfg.FilenamePattern
		if fileCfg.Rotation.Daily {
			logFilename = strings.Replace(logFilename, "{date}", time.Now().Format("2006-01-02"), 1)
		}
		logFilePath := filepath.Join(fileCfg.Dir, logFilename)

		// Konversi MaxSize dari string ke int (dalam MB)
		maxSizeMB := parseMaxSize(fileCfg.Rotation.MaxSize)

		// Gunakan lumberjack untuk rotasi file log
		fileRotator := &lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    maxSizeMB,
			MaxBackups: 0,
			MaxAge:     fileCfg.Rotation.RetentionDays,
			Compress:   fileCfg.Rotation.CompressOld,
		}
		writers = append(writers, fileRotator)
	}

	// B. Output ke Console (stdout)
	if cfg.Output.Console.Enabled {
		writers = append(writers, os.Stdout)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	// 6. Gabungkan semua output writers
	log.SetOutput(io.MultiWriter(writers...))

	return log
}

// --- HOOKS DAN FORMATTER KUSTOM UNTUK LOGRINGKAS ---

// ConditionalCallerHook menambahkan informasi caller (file:line) secara kondisional
type ConditionalCallerHook struct {
	ConfiguredLevel logrus.Level
}

// Levels harus mengembalikan AllLevels karena logika pengecekan ada di Fire()
func (hook *ConditionalCallerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *ConditionalCallerHook) Fire(entry *logrus.Entry) error {
	// Logika Kunci:
	// Jika level config adalah INFO (mode produksi) DAN entry log adalah INFO/DEBUG/TRACE:
	if hook.ConfiguredLevel <= logrus.InfoLevel && entry.Level < logrus.WarnLevel {
		return nil // JANGAN TAMBAHKAN CALLER (Log INFO/DEBUG harus ringkas)
	}

	// Untuk semua kasus lain (Level config WARN/ERROR/FATAL ATAU entry log WARN/ERROR/FATAL): TAMBAHKAN CALLER
	if entry.Caller != nil {
		// Bentuk string caller: file:line
		fileName := filepath.Base(entry.Caller.File)
		callerInfo := fmt.Sprintf("%s:%d", fileName, entry.Caller.Line)

		// Set field "caller_info" agar bisa diakses oleh custom formatter
		entry.Data["caller_info"] = callerInfo
	}
	return nil
}

// CustomTextFormatter mengimplementasikan logrus.Formatter
type CustomTextFormatter struct {
	Location *time.Location
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 1. Ambil & Format Waktu
	timestamp := entry.Time.In(f.Location).Format("2006-01-02 15:04:05")

	// 2. Format Level
	level := strings.ToUpper(entry.Level.String())

	// 3. Ambil Caller (dari ConditionalCallerHook)
	callerInfo := ""
	if info, ok := entry.Data["caller_info"]; ok {
		callerInfo = fmt.Sprintf("[%s]", info)
	}

	// 4. Format Message dan Field Tambahan
	fields := ""
	if len(entry.Data) > 0 {
		for k, v := range entry.Data {
			if k != "caller_info" { // Abaikan caller_info
				fields += fmt.Sprintf(" %s=%v", k, v)
			}
		}
	}

	// Format akhir: [timestamp][LEVEL][caller:line] - Message
	output := fmt.Sprintf("[%s][%s]%s - %s%s\n",
		timestamp,
		level,
		callerInfo,
		entry.Message,
		fields,
	)

	return []byte(output), nil
}

// TimezoneHook (dapat digunakan untuk JSON formatter)
type TimezoneHook struct {
	Location *time.Location
}

// Fire mengubah zona waktu entri log
func (hook *TimezoneHook) Fire(entry *logrus.Entry) error {
	entry.Time = entry.Time.In(hook.Location)
	return nil
}

// Levels mengembalikan semua level
func (hook *TimezoneHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// parseMaxSize mengkonversi string ukuran seperti "100MB" menjadi integer megabyte (100)
func parseMaxSize(sizeStr string) int {
	if len(sizeStr) < 2 {
		return 100 // Default 100MB
	}
	sizeStr = strings.TrimSuffix(sizeStr, "MB")

	var size int
	fmt.Sscanf(sizeStr, "%d", &size)
	return size
}
