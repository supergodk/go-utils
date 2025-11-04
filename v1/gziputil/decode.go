// Package gziputil 提供 gzip 数据压缩和解压缩相关的工具函数
//
// Package gziputil provides gzip compression and decompression utility functions.
package gziputil

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// IsGzipped 检查数据是否为 gzip 格式
// 通过检查数据的魔数（0x1f 0x8b）来判断
//
// IsGzipped checks whether the data is in gzip format.
// It determines by checking the magic number (0x1f 0x8b) of the data.
func IsGzipped(data []byte) bool {
	// gzip格式的魔数是0x1f 0x8b
	return len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b
}

// UnGzip 解压缩 gzip 格式的数据
// 如果输入数据不是 gzip 格式，函数会自动尝试多种 Base64 编码方式解码后再解压
// 支持的 Base64 编码包括：标准编码、无填充编码、URL 安全编码和 URL 安全无填充编码
//
// UnGzip decompresses data in gzip format.
// If the input data is not in gzip format, the function will automatically try multiple Base64 encoding methods to decode before decompressing.
// Supported Base64 encodings include: standard encoding, unpadded encoding, URL-safe encoding, and URL-safe unpadded encoding.
func UnGzip(input []byte) ([]byte, error) {
	// 尝试处理 Base64 包装的数据：若原始输入不是 gzip 头，则尝试多种 Base64 解码
	raw := input
	if !IsGzipped(raw) {
		// 标准 Base64
		if dec, err := base64.StdEncoding.DecodeString(string(input)); err == nil {
			raw = dec
		} else if dec, err := base64.RawStdEncoding.DecodeString(string(input)); err == nil { // 无填充
			raw = dec
		} else if dec, err := base64.URLEncoding.DecodeString(string(input)); err == nil { // URL 安全
			raw = dec
		} else if dec, err := base64.RawURLEncoding.DecodeString(string(input)); err == nil { // URL 安全无填充
			raw = dec
		}
	}

	gr, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return io.ReadAll(gr)
}
