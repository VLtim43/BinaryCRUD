package persistence

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// appendBuilder creates the record instance that will be written, using the next available ID.
type appendBuilder[T any] func(nextID uint32) (T, error)

// recordWriter knows how to serialize a record of type T to the binary file.
type recordWriter[T any] func(*bufio.Writer, T, bool) error

// recordPostProcessor allows the caller to mutate each record after it has been read.
type recordPostProcessor[T any] func(*T, uint32)

// recordFilter decides whether a record should be included in the final result slice.
type recordFilter[T any] func(*T) bool

// recordBeforeWriteHook emits debug information (or other side effects) right before the write.
type recordBeforeWriteHook func(count uint32, offset int64)

// appendBinaryRecord adds a new record of type T to the target file while keeping the header in sync.
// It centralises the boilerplate of opening the file, reading the header, writing the record, and
// bumping the stored record count so item and order persistence can share the same workflow.
func appendBinaryRecord[T any](
	filename string,
	recordLabel string,
	debug bool,
	builder appendBuilder[T],
	writer recordWriter[T],
	beforeWrite recordBeforeWriteHook,
) (*AppendResult, uint32, error) {
	if err := InitFile(filename); err != nil {
		return nil, 0, err
	}

	file, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open %s file: %w", recordLabel, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	count, err := ReadHeader(reader)
	if err != nil {
		return nil, 0, err
	}

	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to seek to end of %s file: %w", recordLabel, err)
	}

	if beforeWrite != nil {
		beforeWrite(count, offset)
	}

	record, err := builder(count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build %s: %w", recordLabel, err)
	}

	writerBuf := bufio.NewWriter(file)
	if err := writer(writerBuf, record, debug); err != nil {
		return nil, 0, fmt.Errorf("failed to write %s: %w", recordLabel, err)
	}

	if err := writerBuf.Flush(); err != nil {
		return nil, 0, fmt.Errorf("failed to flush %s write buffer: %w", recordLabel, err)
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, 0, fmt.Errorf("failed to rewind %s file: %w", recordLabel, err)
	}

	writerBuf = bufio.NewWriter(file)
	if err := WriteHeader(writerBuf, count+1); err != nil {
		return nil, 0, fmt.Errorf("failed to update %s header: %w", recordLabel, err)
	}

	if err := writerBuf.Flush(); err != nil {
		return nil, 0, fmt.Errorf("failed to flush %s header buffer: %w", recordLabel, err)
	}

	return &AppendResult{
		RecordID: count,
		Offset:   offset,
	}, count, nil
}

// readRecords scans an entire file and returns all records that pass the provided filter.
// Callers provide type-specific readers, optional post-processing for each record, and a filter
// that can drop tombstoned entries without duplicating traversal logic.
func readRecords[T any](
	filename string,
	recordLabel string,
	recordReader func(*bufio.Reader) (*T, error),
	postProcess recordPostProcessor[T],
	filter recordFilter[T],
) ([]T, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return []T{}, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s file: %w", recordLabel, err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	count, err := ReadHeader(reader)
	if err != nil {
		return nil, err
	}

	records := make([]T, 0, count)

	for i := uint32(0); i < count; i++ {
		record, err := recordReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s %d: %w", recordLabel, i+1, err)
		}

		if postProcess != nil {
			postProcess(record, i)
		}

		if filter == nil || filter(record) {
			records = append(records, *record)
		}
	}

	return records, nil
}
