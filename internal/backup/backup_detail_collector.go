// File : internal/backup/backup_detail_collector.go
// Deskripsi : Fungsi untuk mengumpulkan detail informasi database secara concurrent
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package backup

import (
	"context"
	"fmt"
	"sfDBTools/pkg/database"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
)

// CollectDatabaseDetails mengumpulkan detail informasi untuk semua database secara concurrent
func (s *Service) CollectDatabaseDetails(ctx context.Context, client *database.Client, dbNames []string) map[string]DatabaseDetailInfo {
	const maxWorkers = 5
	const jobTimeout = 60 * time.Second // Increase overall timeout

	s.Logger.Infof("Mengumpulkan detail informasi untuk %d database...", len(dbNames))
	startTime := time.Now()

	jobs := make(chan DatabaseDetailJob, len(dbNames))
	results := make(chan DatabaseDetailInfo, len(dbNames))

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < maxWorkers; w++ {
		wg.Add(1)
		go s.databaseDetailWorker(ctx, jobs, results, &wg, jobTimeout)
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
	s.Logger.Infof("Pengumpulan detail database selesai dalam %v", duration)

	return detailMap
}

// databaseDetailWorker adalah worker untuk mengumpulkan detail database
func (s *Service) databaseDetailWorker(ctx context.Context, jobs <-chan DatabaseDetailJob, results chan<- DatabaseDetailInfo, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()

	for job := range jobs {
		// Create timeout context for this job
		jobCtx, cancel := context.WithTimeout(ctx, timeout)

		result := s.collectSingleDatabaseDetail(jobCtx, job.Client, job.DatabaseName)
		results <- result

		cancel()
	}
}

// collectSingleDatabaseDetail mengumpulkan detail untuk satu database
func (s *Service) collectSingleDatabaseDetail(ctx context.Context, client *database.Client, dbName string) DatabaseDetailInfo {
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
		metricCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		size, err := client.GetDatabaseSize(metricCtx, dbName)
		metricChan <- metricResult{"size", size, err}
	}()

	// Collect table count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		count, err := client.GetTableCount(metricCtx, dbName)
		metricChan <- metricResult{"tables", int64(count), err}
	}()

	// Collect procedure count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		count, err := client.GetProcedureCount(metricCtx, dbName)
		metricChan <- metricResult{"procedures", int64(count), err}
	}()

	// Collect function count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		count, err := client.GetFunctionCount(metricCtx, dbName)
		metricChan <- metricResult{"functions", int64(count), err}
	}()

	// Collect view count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		count, err := client.GetViewCount(metricCtx, dbName)
		metricChan <- metricResult{"views", int64(count), err}
	}()

	// Collect user grant count with timeout
	metricWg.Add(1)
	go func() {
		defer metricWg.Done()
		metricCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
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
		s.Logger.Warnf("Error mengumpulkan detail database %s: %v", dbName, errors)
	}

	return detail
}
