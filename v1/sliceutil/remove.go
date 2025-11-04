// Package sliceutil 提供切片操作相关的工具函数
//
// Package sliceutil provides slice manipulation utility functions.
package sliceutil

import "slices"

// Remove 从切片中移除指定的元素，返回一个新的切片
// T 必须是可比较的类型（comparable），例如 int、string、float64 等
// 参数:
//   - slice: 原始切片
//   - elements: 要移除的元素列表
//
// 返回:
//   - 移除指定元素后的新切片，原始切片不会被修改
//
// Remove removes specified elements from a slice and returns a new slice.
// T must be a comparable type (comparable), such as int, string, float64, etc.
// Parameters:
//   - slice: The original slice
//   - elements: The list of elements to remove
//
// Returns:
//   - A new slice with specified elements removed, the original slice is not modified
func Remove[T comparable](slice []T, elements []T) []T {
	if len(slice) == 0 {
		return []T{}
	}

	if len(elements) == 0 {
		return slices.Clone(slice)
	}

	// 创建一个元素查找映射，提高查找效率
	elementMap := make(map[T]struct{}, len(elements))
	for _, e := range elements {
		elementMap[e] = struct{}{}
	}

	// 创建一个新的切片，用于存储结果
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		// 只有当元素不在要移除的集合中时，才将其加入结果切片
		if _, exists := elementMap[v]; !exists {
			result = append(result, v)
		}
	}

	return result
}
