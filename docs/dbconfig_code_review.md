# Code Review — Modul `internal/dbconfig`

Dokumen ini merangkum hasil review kode untuk modul `internal/dbconfig` berikut rekomendasi perbaikan, optimasi, dan rencana refactor bertahap. Fokus utama: stabilitas/bug, UX CLI, keamanan (enkripsi/password), konsistensi arsitektur, dan kemudahan perawatan.

## Ringkasan

- Ditemukan risiko panic pada alur show dengan flag `--file` karena snapshot tidak dimuat sebelum dipakai.
- Alur non-interaktif edit/create belum konsisten dalam normalisasi path dan pemuatan kunci enkripsi.
- Terdapat duplikasi logika baca→dekripsi→parse file konfigurasi; perlu diekstrak ke helper tunggal.
- Beberapa validasi belum ketat (nama file, kunci enkripsi non-empty, filter file `.cnf.enc`).
- Ada perbaikan UX kecil yang akan meningkatkan kejelasan dan keandalan CLI.

## Masalah Kritis (Prioritas Tinggi)

1. Potensi panic di `ShowDatabaseConfig()` (`dbconfig_show.go`)

   - Problema: Saat flag `--file` digunakan, fungsi langsung mengakses `s.OriginalDBConfigInfo.FilePath` tanpa memastikan snapshot telah dimuat (berisiko nil pointer).
   - Dampak: Crash/panic pada mode non-interaktif.
   - Perbaikan: Bila `--file` diisi, lakukan proses load: resolve path → baca → dekripsi → parse → set `s.OriginalDBConfigInfo` sebelum display.

  - Status: DONE — Implementasi: resolve path + load snapshot saat `--file`, cegah akses `nil` (lihat `dbconfig_show.go`).

2. Alur edit non-interaktif salah kelola path dan nama (`dbconfig_edit.go`)

   - Problema:
     - `fs.Exists(s.OriginalConfigName)` mengasumsikan `--file` adalah path absolut; jika hanya nama, pengecekan gagal.
     - `ConfigName` diisi dari nilai yang bisa kosong alih-alih dari basename `--file` (tanpa `.cnf.enc`).
     - `FilePath` diisi mentah dari input tanpa normalisasi ke path absolut config dir.
   - Perbaikan:
     - Normalisasi input `--file`: jika bukan path absolut, gabungkan dengan `ConfigDir` dan pastikan suffix `.cnf.enc`.
     - Set `ConfigName` = basename tanpa `.cnf.enc` secara konsisten.

   - Status: DONE — Implementasi: normalisasi `--file`, set `ConfigName` dari basename, dan load snapshot via helper (lihat `dbconfig_edit.go`).

3. Menyimpan tanpa validasi kunci enkripsi (`dbconfig_save.go`)

   - Problema: `SaveDBConfig()` mengenkripsi dengan `s.DBConfigInfo.EncryptionKey` tanpa memastikan tidak kosong.
   - Dampak: Enkripsi dengan kunci kosong atau error yang tidak informatif.
   - Perbaikan: Validasi wajib. Jika kosong → coba ambil dari env (atau jalur resolver bersama) → jika tetap kosong, fail-fast dengan pesan jelas.

  - Status: DONE — Implementasi: validasi kunci enkripsi non-empty sebelum enkripsi (lihat `dbconfig_save.go`).

## UX/Perilaku CLI

- `CreateDatabaseConfig()` dan `EditDatabaseConfig()` saat wizard dibatalkan mengembalikan `nil`. Sebaiknya bedakan “dibatalkan pengguna” vs “gagal teknis” dengan error sentinel agar exit code non-zero pada error teknis.
  - Status: DONE — Implementasi sentinel `ErrUserCancelled` dan propagate dari wizard ke Create/Edit.
- `PromptDeleteConfigs()` tidak memfilter ekstensi. Filter hanya file `.cnf.enc` agar aman.
  - Status: DONE — Implementasi: filter `.cnf.enc` pada prompt select/delete.
- `ValidateDatabaseConfig()` juga perlu memuat snapshot saat `--file` digunakan (konsisten dengan perbaikan pada show/edit).
  - Status: DONE — Implementasi: resolve path + load snapshot saat `--file` (lihat `dbconfig_validate.go`).

## Keamanan

- Kunci enkripsi:
  - Sumber kunci bercampur (env/flag/prompt) dan tidak konsisten. Standarkan melalui satu helper: `resolveEncryptionKey()` yang mengembalikan (key, source, err).
    - Status: DONE — Implementasi helper `resolveEncryptionKey()` dan dipakai di load snapshot + save + wizard.
  - Hindari menyimpan `EncryptionKey` pada snapshot yang bisa terbaca; cukup simpan sementara di memori saat proses.
    - Status: DONE — Snapshot tidak lagi menyimpan `EncryptionKey` saat pemilihan file.
- Logging:
  - Pastikan tidak pernah log plaintext password maupun encryption key.

## Konsistensi & Arsitektur

Kurangi duplikasi dan pusatkan operasi berulang ke helper:

- `resolveConfigPath(input, configDir) (absPath, name string, err error)`
- Status: DONE — Implementasi helper `resolveConfigPath()` dan dipakai di show/edit/validate.
- `resolveEncryptionKey(existing string) (key, source string, err error)`
- Status: DONE — Implementasi helper `resolveEncryptionKey()` dan dipakai di show/edit/validate.
- `loadAndParseConfig(absPath, key string) (*structs.DBConfigInfo, error)`
- Status: DONE — Implementasi helper `loadAndParseConfig()` dan dipakai di show/edit/validate.

Manfaat: satu pintu untuk read→decrypt→parse; dipakai oleh show/edit/validate/reveal.

Standarisasi manipulasi nama file:

- `normalizeConfigName(name string) string` untuk memastikan tanpa `.cnf.enc`.
- `ensureConfigFilePath(nameOrPath string, configDir string) (absPath string, normalizedName string, err error)`.
  - Status: DONE — `normalizeConfigName` tercakup oleh helper `common.TrimConfigSuffix()`, dan `ensureConfigFilePath` tercakup oleh `resolveConfigPath()`.

Pisahkan tanggung jawab UI vs log atau pertahankan keduanya secara konsisten. Hindari pesan ganda yang membingungkan.

## Validasi dan Error Handling

- `CheckConfigurationNameUnique()` (mode edit): jika `original == ""`, jangan menganggap target yang tidak ada sebagai error “file tidak ditemukan” tanpa konteks. Berikan pesan yang mengarahkan user untuk menggunakan `--file` atau wizard.
- `ValidateInput()` terlalu minimal. Tambahkan validasi `host`, `port > 0`, `user` sebelum simpan.
- Gunakan sentinel error (mis. `ErrUserCancelled`) untuk membedakan cancel vs gagal teknis.

## Quick Wins (Aman, Dampak Besar)

- `ShowDatabaseConfig()`: muat snapshot saat `--file` digunakan; hindari akses nil.
  - Status: DONE
- `EditDatabaseConfig()` non-interaktif: normalisasi `--file` → path absolut + set `ConfigName` dari basename.
  - Status: DONE
- `SaveDBConfig()`: validasi `EncryptionKey` non-empty; ambil dari resolver jika perlu.
  - Status: DONE (validasi non-empty; kini memakai resolver terpusat)
- `PromptDeleteConfigs()` dan `promptSelectExistingConfig()`: filter hanya file dengan ekstensi `.cnf.enc`.
  - Status: DONE
- Tambah helper kecil `common.TrimConfigSuffix(name string)` dan gunakan di seluruh lokasi yang memanipulasi `.cnf.enc`.
  - Status: DONE

## Rencana Refactor Bertahap

1. Ekstrak helper bersama:

- `resolveConfigPath`, `resolveEncryptionKey`, `loadAndParseConfig` (1 file helper di package ini terlebih dulu).

  - Status: DONE — Semua helper terpusat sudah ada dan dipakai lintas flow.

1. Ganti penggunaan manual di `show`, `edit` (keduanya), `validate`, dan `reveal` untuk memakai helper.

  - Status: DONE — Seluruh jalur tersebut sekarang menggunakan helper bersama, termasuk `reveal`.

1. Standarkan normalisasi nama/path saat masuk (flag, prompt) dan saat keluar (save).

  - Status: DONE — Normalisasi nama/path distandardkan via `common.TrimConfigSuffix`, `ensureConfigExt`, dan `resolveConfigPath` untuk input; `SaveDBConfig` memastikan ekstensi saat output.

1. Tambahkan sentinel error dan perbaiki alur exit code.

  - Status: DONE — Sentinel `ErrUserCancelled` digunakan di service-layer dan perintah CLI (RunE) memetakan cancel menjadi exit sukses tanpa error log.

## Rekomendasi Test Minimal

- Unit test `parseINIClient`:
  - Happy path, komentar/line kosong, section `[client]` case-insensitive.
- Unit test `CheckConfigurationNameUnique`:
  - Create: nama ada/tidak ada.
  - Edit: rename ke nama yang sudah ada, rename valid, nama tidak berubah tapi file lama tidak ada.
- Test helper `resolveConfigPath`:
  - Input nama vs path absolut; memastikan suffix `.cnf.enc` ditangani benar.

## Catatan Tambahan Spesifik Kode

- `dbconfig_prompt.go`: setelah dekripsi sukses, snapshot `EncryptionKey` sebaiknya tidak diisi dari state yang mungkin kosong; konsisten gunakan key yang dipakai saat dekripsi bila memang perlu (atau jangan simpan sama sekali demi keamanan).
  - Status: DONE — Tidak lagi menyimpan `EncryptionKey` pada snapshot hasil pemilihan file.
- `dbconfig_display.go`: kolom “Is Valid” pada tabel show tidak pernah di-set. Tampilkan hanya jika sudah menjalankan validasi atau hapus kolom untuk menghindari asumsi.
  - Status: PENDING — Belum diubah; rekomendasi UX.
- `dbconfig_save.go`: proses rename menulis baru lalu menghapus lama—sudah aman. Pertimbangkan check permission awal untuk UX yang lebih halus.
  - Status: N/A (opsional) — Tidak mengubah perilaku saat ini.

## Contoh Kontrak Helper yang Disarankan

- Input/Output dan error modes singkat:
  - `resolveConfigPath(nameOrPath, configDir)`
    - Input: nama (tanpa atau dengan `.cnf.enc`) atau path.
    - Output: `absPath`, `normalizedName` (tanpa `.cnf.enc`).
    - Error: path tidak ditemukan (non-interaktif), atau kembalikan pilihan ke prompt (interaktif).
  - `resolveEncryptionKey(existing)`
    - Input: key dari flag/state (opsional).
    - Output: key final + sumber (env/prompt/flag).
    - Error: user cancel atau tidak tersedia.
  - `loadAndParseConfig(absPath, key)`
    - Output: `*structs.DBConfigInfo` terisi host/port/user/password + metadata.
    - Error: read/decrypt/parse error (dibungkus konteks).

## Next Steps

- Implementasikan quick wins di atas (4–6 perubahan kecil).
- Setelah itu, lakukan refactor helper bersama dan perbarui pemanggilnya.
- Tambahkan unit test ringkas untuk menjaga regresi.

Jika dibutuhkan, dokumen ini bisa diperluas dengan snippet patch spesifik per file untuk mempermudah implementasi.
