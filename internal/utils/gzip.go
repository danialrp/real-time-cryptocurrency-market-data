package utils

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

func DecompressGZIP(data []byte) ([]byte, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("data is too short to be a valid gzip format")
	}

	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	decompressedData, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read decompressed data: %v", err)
	}

	return decompressedData, nil
}

func CompressGZIP(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)

	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to write data to gzip: %v", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %v", err)
	}

	return buf.Bytes(), nil
}
