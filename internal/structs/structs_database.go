package structs

// DatabaseSizeInfo menyimpan informasi ukuran database
type DatabaseSizeInfo struct {
	DatabaseName string
	DataSize     int64
	IndexSize    int64
	TotalSize    int64
}
