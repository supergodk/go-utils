// Package timeutil 提供时间处理相关的工具函数
//
// Package timeutil provides time manipulation utility functions.
package timeutil

import (
	"sync/atomic"
	"time"
)

var (
	// lastTickMicroSecond 记录上次生成的微秒时间戳
	// 使用原子操作代替互斥锁，提高并发性能
	//
	// lastTickMicroSecond records the last generated microsecond timestamp
	// Uses atomic operations instead of mutex locks to improve concurrency performance
	lastTickMicroSecond int64
)

const (
	// LocationUTC UTC 时区
	//
	// LocationUTC represents UTC timezone
	LocationUTC = "UTC"
	// LocationLocal 本地时区
	//
	// LocationLocal represents local timezone
	LocationLocal = "Local"

	// LocationShanghai 中国上海时区（东八区）
	//
	// LocationShanghai represents Shanghai, China timezone (UTC+8)
	LocationShanghai = "Asia/Shanghai"
	// LocationBeijing 中国北京时区（东八区）
	//
	// LocationBeijing represents Beijing, China timezone (UTC+8)
	LocationBeijing = "Asia/Beijing"
	// LocationHongKong 香港时区（东八区）
	//
	// LocationHongKong represents Hong Kong timezone (UTC+8)
	LocationHongKong = "Asia/Hong_Kong"
	// LocationTaipei 台北时区（东八区）
	//
	// LocationTaipei represents Taipei timezone (UTC+8)
	LocationTaipei = "Asia/Taipei"

	// LocationTokyo 日本东京时区（东九区）
	//
	// LocationTokyo represents Tokyo, Japan timezone (UTC+9)
	LocationTokyo = "Asia/Tokyo"
	// LocationSeoul 韩国首尔时区（东九区）
	//
	// LocationSeoul represents Seoul, South Korea timezone (UTC+9)
	LocationSeoul = "Asia/Seoul"
	// LocationSingapore 新加坡时区（东八区）
	//
	// LocationSingapore represents Singapore timezone (UTC+8)
	LocationSingapore = "Asia/Singapore"
	// LocationKolkata 印度加尔各答时区（东五区半）
	//
	// LocationKolkata represents Kolkata, India timezone (UTC+5:30)
	LocationKolkata = "Asia/Kolkata"

	// LocationNewYork 美国纽约时区（东部时间）
	//
	// LocationNewYork represents New York, USA timezone (Eastern Time)
	LocationNewYork = "America/New_York"
	// LocationLosAngeles 美国洛杉矶时区（太平洋时间）
	//
	// LocationLosAngeles represents Los Angeles, USA timezone (Pacific Time)
	LocationLosAngeles = "America/Los_Angeles"
	// LocationChicago 美国芝加哥时区（中部时间）
	//
	// LocationChicago represents Chicago, USA timezone (Central Time)
	LocationChicago = "America/Chicago"
	// LocationDenver 美国丹佛时区（山地时间）
	//
	// LocationDenver represents Denver, USA timezone (Mountain Time)
	LocationDenver = "America/Denver"

	// LocationLondon 英国伦敦时区（格林威治时间）
	//
	// LocationLondon represents London, UK timezone (Greenwich Mean Time)
	LocationLondon = "Europe/London"
	// LocationParis 法国巴黎时区（中欧时间）
	//
	// LocationParis represents Paris, France timezone (Central European Time)
	LocationParis = "Europe/Paris"
	// LocationBerlin 德国柏林时区（中欧时间）
	//
	// LocationBerlin represents Berlin, Germany timezone (Central European Time)
	LocationBerlin = "Europe/Berlin"
	// LocationMoscow 俄罗斯莫斯科时区（东三区）
	//
	// LocationMoscow represents Moscow, Russia timezone (UTC+3)
	LocationMoscow = "Europe/Moscow"
)

// GetCurrentMonthTime 获取当前月份的开始和结束时间戳（Unix 时间戳）
// 返回:
//   - 第一个值: 当前月份第一天的 00:00:00 的 Unix 时间戳
//   - 第二个值: 当前月份最后一天 23:59:59 的 Unix 时间戳
//
// GetCurrentMonthTime returns the start and end timestamps (Unix timestamp) of the current month.
// Returns:
//   - First value: Unix timestamp of 00:00:00 on the first day of the current month
//   - Second value: Unix timestamp of 23:59:59 on the last day of the current month
func GetCurrentMonthTime() (int64, int64) {
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location()).Add(-time.Second)
	return startOfMonth.Unix(), endOfMonth.Unix()
}

// JudgeTimeInRange 判断给定的时间戳是否在指定的时间范围内（相对于当前时间）
// 参数:
//   - timestamp: 要判断的 Unix 时间戳
//   - timeRange: 时间范围，例如 5*time.Minute 表示 5 分钟
//
// 返回:
//   - 如果当前时间与给定时间戳的绝对差值小于等于 timeRange，返回 true，否则返回 false
//
// JudgeTimeInRange determines whether the given timestamp is within the specified time range (relative to the current time).
// Parameters:
//   - timestamp: The Unix timestamp to check
//   - timeRange: The time range, e.g., 5*time.Minute means 5 minutes
//
// Returns:
//   - Returns true if the absolute difference between the current time and the given timestamp is less than or equal to timeRange, otherwise returns false
func JudgeTimeInRange(timestamp int64, timeRange time.Duration) bool {
	t := time.Unix(timestamp, 0)
	now := time.Now()
	duration := now.Sub(t).Abs() // 使用 Abs() 更简洁
	return duration <= timeRange
}

// GetYesterdayTime 获取指定时区中，给定时间戳对应的昨天和今天的开始时间戳
// 函数会根据指定的时区计算"昨天"和"今天"的边界，返回这两个时间点对应的 Unix 时间戳
// 参数:
//   - timestamp: 基准 Unix 时间戳
//   - locationName: 时区名称，例如 "Asia/Shanghai" 或使用常量如 LocationShanghai
//     如果为空字符串，则使用本地时区（Local）
//
// 返回:
//   - 第一个值: 昨天 00:00:00 的 Unix 时间戳（在指定时区下）
//   - 第二个值: 今天 00:00:00 的 Unix 时间戳（在指定时区下）
//   - error: 如果时区名称无效，返回错误
//
// GetYesterdayTime returns the start timestamps of yesterday and today corresponding to the given timestamp in the specified timezone.
// The function calculates the boundaries of "yesterday" and "today" based on the specified timezone and returns the Unix timestamps for these two time points.
// Parameters:
//   - timestamp: Base Unix timestamp
//   - locationName: Timezone name, e.g., "Asia/Shanghai" or use constants like LocationShanghai
//     If it is an empty string, the local timezone (Local) will be used
//
// Returns:
//   - First value: Unix timestamp of 00:00:00 yesterday (in the specified timezone)
//   - Second value: Unix timestamp of 00:00:00 today (in the specified timezone)
//   - error: Returns an error if the timezone name is invalid
func GetYesterdayTime(timestamp int64, locationName string) (int64, int64, error) {
	if locationName == "" {
		locationName = "Local"
	}
	loc, err := time.LoadLocation(locationName) // 或你的时区
	if err != nil {
		return 0, 0, err
	}

	tt := time.Unix(timestamp, 0).In(loc)

	// 使用 AddDate 更安全（自动处理月份边界）
	yesterday := tt.AddDate(0, 0, -1)
	yesterdayStart := time.Date(
		yesterday.Year(), yesterday.Month(), yesterday.Day(),
		0, 0, 0, 0, loc,
	).Unix()

	todayStart := time.Date(
		tt.Year(), tt.Month(), tt.Day(),
		0, 0, 0, 0, loc,
	).Unix()

	return yesterdayStart, todayStart, nil
}

// GetLastMonthTime 获取给定时间戳对应的上个月的起止时间戳
// 从上个月首日的 00:00:00 到最后一日的 23:59:59
// 参数:
//   - timestamp: 基准 Unix 时间戳
//
// 返回:
//   - 第一个值: 上个月第一天 00:00:00 的 Unix 时间戳
//   - 第二个值: 上个月最后一天 23:59:59 的 Unix 时间戳
//
// GetLastMonthTime returns the start and end timestamps of the previous month corresponding to the given timestamp.
// From 00:00:00 on the first day of the previous month to 23:59:59 on the last day of the previous month.
// Parameters:
//   - timestamp: Base Unix timestamp
//
// Returns:
//   - First value: Unix timestamp of 00:00:00 on the first day of the previous month
//   - Second value: Unix timestamp of 23:59:59 on the last day of the previous month
func GetLastMonthTime(timestamp int64) (int64, int64) {
	date := time.Unix(timestamp, 0)

	// 获取当前月份的第一天，然后减去一个月
	firstOfCurrentMonth := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	firstOfLastMonth := firstOfCurrentMonth.AddDate(0, -1, 0)

	// 上个月的最后一天 23:59:59
	lastOfLastMonth := firstOfCurrentMonth.Add(-time.Second)

	return firstOfLastMonth.Unix(), lastOfLastMonth.Unix()
}

// ClockTickMicroSecondUniqFast 生成唯一的、单调递增的微秒级时间戳
// 使用原子操作保证线程安全，适用于高并发场景
// 如果同一微秒内有多次调用，会自动递增以确保唯一性
//
// 返回:
//   - 唯一的微秒级 Unix 时间戳（Unix 时间戳 * 1000000）
//
// ClockTickMicroSecondUniqFast generates a unique, monotonically increasing microsecond-level timestamp.
// Uses atomic operations to ensure thread safety, suitable for high-concurrency scenarios.
// If there are multiple calls within the same microsecond, it will automatically increment to ensure uniqueness.
//
// Returns:
//   - A unique microsecond-level Unix timestamp (Unix timestamp * 1000000)
func ClockTickMicroSecondUniqFast() int64 {
	tNow := time.Now().UnixMicro()
	for {
		last := atomic.LoadInt64(&lastTickMicroSecond)
		if tNow <= last {
			tNow = last + 1
		}
		if atomic.CompareAndSwapInt64(&lastTickMicroSecond, last, tNow) {
			return tNow
		}
	}
}
