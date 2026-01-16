package queue

import (
	"testing"
)

// BenchmarkMergeSortedJobs benchmarks the mergeSortedJobs function
func BenchmarkMergeSortedJobs(b *testing.B) {
	// Create two sorted job slices
	jobs1 := make([]Job, 100)
	jobs2 := make([]Job, 100)
	for i := 0; i < 100; i++ {
		jobs1[i] = Job{Priority: 200 - i*2}
		jobs2[i] = Job{Priority: 199 - i*2}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mergeSortedJobs(jobs1, jobs2)
	}
}

// BenchmarkSliceInsertion benchmarks the optimized slice insertion
func BenchmarkSliceInsertion(b *testing.B) {
	b.Run("optimized_copy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobs := make([]Job, 100)
			for j := 0; j < 100; j++ {
				jobs[j] = Job{Priority: 100 - j}
			}
			// Optimized insertion at middle
			insertIdx := 50
			job := Job{Priority: 75}
			jobs = append(jobs, Job{})
			copy(jobs[insertIdx+1:], jobs[insertIdx:])
			jobs[insertIdx] = job
		}
	})

	b.Run("double_append", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobs := make([]Job, 100)
			for j := 0; j < 100; j++ {
				jobs[j] = Job{Priority: 100 - j}
			}
			// Old double-append pattern
			insertIdx := 50
			job := Job{Priority: 75}
			jobs = append(jobs[:insertIdx], append([]Job{job}, jobs[insertIdx:]...)...)
			_ = jobs
		}
	})
}

// BenchmarkPreallocatedSlice benchmarks the preallocated vs non-preallocated slice
func BenchmarkPreallocatedSlice(b *testing.B) {
	b.Run("preallocated", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobs := make([]*Job, 0, 100)
			for j := 0; j < 100; j++ {
				jobs = append(jobs, &Job{Priority: j})
			}
		}
	})

	b.Run("non_preallocated", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobs := []*Job{}
			for j := 0; j < 100; j++ {
				jobs = append(jobs, &Job{Priority: j})
			}
		}
	})
}

// BenchmarkSliceClear benchmarks clearing a slice
func BenchmarkSliceClear(b *testing.B) {
	b.Run("reuse_capacity", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobs := make([]*Job, 100)
			for j := 0; j < 100; j++ {
				jobs[j] = &Job{Priority: j}
			}
			// Clear by reslicing to reuse capacity
			jobs = jobs[:0]
			_ = jobs
		}
	})

	b.Run("new_slice", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			jobs := make([]*Job, 100)
			for j := 0; j < 100; j++ {
				jobs[j] = &Job{Priority: j}
			}
			// Create new empty slice
			jobs = []*Job{}
			_ = jobs
		}
	})
}

// BenchmarkMin benchmarks the Min function
func BenchmarkMin(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Min(100, 50)
		_ = Min(50, 100)
	}
}

// BenchmarkJobCreation benchmarks job creation with nil vs empty stacktrace
func BenchmarkJobCreation(b *testing.B) {
	b.Run("nil_stacktrace", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = &Job{
				Id:       "test",
				Priority: 1,
				Status:   WaitStatus,
			}
		}
	})

	b.Run("empty_stacktrace", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = &Job{
				Id:         "test",
				Priority:   1,
				Status:     WaitStatus,
				Stacktrace: []string{},
			}
		}
	})
}

