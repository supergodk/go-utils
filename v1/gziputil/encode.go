// Package gziputil 提供 gzip 数据压缩和解压缩相关的工具函数
//
// Package gziputil provides gzip compression and decompression utility functions.
package gziputil

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
)

var (
	// ErrGzipCompress 表示 gzip 压缩失败
	//
	// ErrGzipCompress indicates that gzip compression failed.
	ErrGzipCompress = errors.New("gzip compression failed")
)

// Gzip 压缩数据为 gzip 格式
// level 参数指定压缩级别，范围从 gzip.HuffmanOnly 到 gzip.BestCompression
// 如果 level 超出有效范围，将使用默认压缩级别
//
// Gzip compresses data into gzip format.
// The level parameter specifies the compression level, ranging from gzip.HuffmanOnly to gzip.BestCompression.
// If the level is out of valid range, the default compression level will be used.
func Gzip(input []byte, level int) ([]byte, error) {
	if len(input) == 0 {
		return nil, errors.New("input data is empty")
	}

	// 压缩级别校验
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}

	// 创建缓冲区接收压缩结果
	var buf bytes.Buffer

	// 创建gzip写入器
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGzipCompress, err)
	}

	// 写入数据
	_, err = writer.Write(input)
	if err != nil {
		writer.Close()
		return nil, fmt.Errorf("%w: %v", ErrGzipCompress, err)
	}

	// 关闭写入器完成压缩
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGzipCompress, err)
	}

	return buf.Bytes(), nil
}
