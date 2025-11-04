// Package randutil 提供随机数生成相关的工具函数
//
// Package randutil provides random number generation utility functions.
package randutil

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"sync"
	"time"
)

// 线程安全的随机数生成器
//
// Thread-safe random number generator
var (
	rnd      *mathrand.Rand
	rndMutex sync.Mutex
	once     sync.Once
)

// initRand 初始化安全的随机数生成器
// 使用加密安全的随机数作为种子，如果失败则使用当前时间作为种子
//
// initRand initializes a secure random number generator.
// It uses cryptographically secure random numbers as the seed, and falls back to the current time if that fails.
func initRand() {
	once.Do(func() {
		var seed int64
		// 尝试使用加密安全的随机数生成种子
		max := big.NewInt(math.MaxInt64)
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			// 如果加密随机数失败，使用时间作为种子
			seed = time.Now().UnixNano()
		} else {
			seed = n.Int64()
		}
		rnd = mathrand.New(mathrand.NewSource(seed))
	})
}

// GenerateRandomDigits 生成指定位数的随机数字
// 参数:
//   - digits: 随机数的位数，必须为正整数
//
// 返回:
//   - 生成的随机数
//   - 错误信息，如果参数无效
//
// GenerateRandomDigits generates a random number with the specified number of digits.
// Parameters:
//   - digits: The number of digits for the random number, must be a positive integer
//
// Returns:
//   - The generated random number
//   - An error if the parameter is invalid
func GenerateRandomDigits(digits int) (int, error) {
	if digits <= 0 {
		return 0, fmt.Errorf("%w: 位数必须为正整数", errors.New("无效的位数参数"))
	}

	// 计算最小值和最大值
	min := int(math.Pow10(digits - 1))
	max := int(math.Pow10(digits)) - 1

	// 检查溢出
	if min < 0 || max < 0 {
		return 0, fmt.Errorf("%w: 请求的位数太大", errors.New("数字溢出"))
	}

	// 使用线程安全的随机数生成
	initRand()
	rndMutex.Lock()
	defer rndMutex.Unlock()

	return rnd.Intn(max-min+1) + min, nil
}
