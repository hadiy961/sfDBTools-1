// File : pkg/database/database_detail_collector.go
// Deskripsi : Fungsi untuk mengumpulkan detail informasi database secara concurrent
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package database

import (
	"context"
	"fmt"
	"runtime"
	"sfDBTools/internal/applog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

// DatabaseDetailInfo berisi informasi detail database
type DatabaseDetailInfo struct {
	DatabaseName   string `json:"database_name"`
	SizeBytes      int64  `json:"size_bytes"`
	SizeHuman      string `json:"size_human"`
	TableCount     int    `json:"table_count"`
	ProcedureCount int    `json:"procedure_count"`
	FunctionCount  int    `json:"function_count"`
	ViewCount      int    `json:"view_count"`
	UserGrantCount int    `json:"user_grant_count"`
	CollectionTime string `json:"collection_time"`
	Error          string `json:"error,omitempty"` // jika ada error saat collect
}

// DatabaseDetailJob untuk worker pattern
type DatabaseDetailJob struct {
	DatabaseName string
	Client       *Client
}

// CollectDatabaseDetails mengumpulkan detail informasi untuk semua database secara concurrent
func CollectDatabaseDetails(ctx context.Context, client *Client, dbNames []string, logger applog.Logger) map[string]DatabaseDetailInfo {
	const jobTimeout = 300 * time.Second // Increase overall timeout

	// If there are no databases, return early.
	if len(dbNames) == 0 {
		logger.Infof("No databases to collect details for")
		return map[string]DatabaseDetailInfo{}
	}

	// Determine number of workers dynamically from available CPUs.
	maxWorkers := runtime.NumCPU()
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	// Do not create more workers than databases.
	if maxWorkers > len(dbNames) {
		maxWorkers = len(dbNames)
	}

	total := len(dbNames)
	// progress counters
	var started int32
	var completed int32
	var failed int32

	logger.Infof("Mengumpulkan detail informasi untuk %d database... workers=%d", total, maxWorkers)

	startTime := time.Now()

	jobs := make(chan DatabaseDetailJob, len(dbNames))
	results := make(chan DatabaseDetailInfo, len(dbNames))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		workerID := w
		go databaseDetailWorker(ctx, jobs, results, &wg, jobTimeout, logger, &started, &completed, &failed, total, workerID)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for _, dbName := range dbNames {
			jobs <- DatabaseDetailJob{
				DatabaseName: dbName,
				Client:       client,
			}
		}
	}()

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	detailMap := make(map[string]DatabaseDetailInfo)
	for result := range results {
		detailMap[result.DatabaseName] = result
	}

	duration := time.Since(startTime)
	logger.Infof("Pengumpulan detail database selesai dalam %v", duration)

	return detailMap
}

// databaseDetailWorker adalah worker untuk mengumpulkan detail database
func databaseDetailWorker(ctx context.Context, jobs <-chan DatabaseDetailJob, results chan<- DatabaseDetailInfo, wg *sync.WaitGroup, timeout time.Duration, logger applog.Logger, started *int32, completed *int32, failed *int32, total int, workerID int) {
	defer wg.Done()

	for job := range jobs {
		// mark started
		atomic.AddInt32(started, 1)

		// Create timeout context for this job
		jobStart := time.Now()
		jobCtx, cancel := context.WithTimeout(ctx, timeout)

		result := collectSingleDatabaseDetail(jobCtx, job.Client, job.DatabaseName, logger)

		// send result
		results <- result

		// update counters
		if result.Error != "" {
			atomic.AddInt32(failed, 1)
		}
		done := atomic.AddInt32(completed, 1)

		// log per-database finish and overall progress
		elapsed := time.Since(jobStart)
		if result.Error != "" {
			logger.Warnf("worker-%d: finished %s in %v (error=%s)", workerID, job.DatabaseName, elapsed, result.Error)
		}

		percent := (float64(done) / float64(total)) * 100.0
		logger.Infof("Progress: %d/%d completed (%.1f%%), failed=%d (%s)", done, total, percent, atomic.LoadInt32(failed), job.DatabaseName)

		cancel()
	}
}

// collectSingleDatabaseDetail mengumpulkan detail untuk satu database
func collectSingleDatabaseDetail(ctx context.Context, client *Client, dbName string, logger applog.Logger) DatabaseDetailInfo {
	startTime := time.Now()
	detail := DatabaseDetailInfo{
		DatabaseName:   dbName,
		CollectionTime: startTime.Format("2006-01-02 15:04:05"),
	}

	// Channel untuk concurrent collection
	type metricResult struct {
		metricType string
		value      int64
		err        error
	}

	metricChan := make(chan metricResult, 6)
	var metricWg sync.WaitGroup

	// Collect database size with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx := ctx
		size, err := client.GetDatabaseSize(metricCtx, dbName)
		metricChan <- metricResult{"size", size, err}
	}()

	// Collect table count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx := ctx
		count, err := client.GetTableCount(metricCtx, dbName)
		metricChan <- metricResult{"tables", int64(count), err}
	}()

	// Collect procedure count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx := ctx
		count, err := client.GetProcedureCount(metricCtx, dbName)
		metricChan <- metricResult{"procedures", int64(count), err}
	}()

	// Collect function count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx := ctx
		count, err := client.GetFunctionCount(metricCtx, dbName)
		metricChan <- metricResult{"functions", int64(count), err}
	}()

	// Collect view count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx := ctx
		count, err := client.GetViewCount(metricCtx, dbName)
		metricChan <- metricResult{"views", int64(count), err}
	}()

	// Collect user grant count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx := ctx
		count, err := client.GetUserGrantCount(metricCtx, dbName)
		metricChan <- metricResult{"user_grants", int64(count), err}
	}()

	// Close channel when all metrics are collected
	go func() {
		metricWg.Wait()
		close(metricChan)
	}()

	// Process results
	var errors []string
	for result := range metricChan {
		if result.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", result.metricType, result.err))
			continue
		}

		switch result.metricType {
		case "size":
			detail.SizeBytes = result.value
			detail.SizeHuman = humanize.Bytes(uint64(result.value))
		case "tables":
			detail.TableCount = int(result.value)
		case "procedures":
			detail.ProcedureCount = int(result.value)
		case "functions":
			detail.FunctionCount = int(result.value)
		case "views":
			detail.ViewCount = int(result.value)
		case "user_grants":
			detail.UserGrantCount = int(result.value)
		}
	}

	if len(errors) > 0 {
		detail.Error = fmt.Sprintf("Errors: %v", errors)
		if logger != nil {
			logger.Warnf("Error mengumpulkan detail database %s: %v", dbName, errors)
		}
	}

	return detail
}
