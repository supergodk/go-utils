// Package cryptoutil 提供加密相关的工具函数
//
// Package cryptoutil provides cryptographic utility functions.
package cryptoutil

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

const (
	// AppleAuthKeysURL Apple 身份验证公钥 URL
	//
	// AppleAuthKeysURL is the URL for Apple identity authentication public keys
	AppleAuthKeysURL = "https://appleid.apple.com/auth/keys"

	// KeyCacheTTL 公钥缓存有效期（24小时）
	//
	// KeyCacheTTL is the public key cache validity period (24 hours)
	KeyCacheTTL = 24 * time.Hour

	// HTTPRequestTimeout HTTP 请求超时（10秒）
	//
	// HTTPRequestTimeout is the HTTP request timeout (10 seconds)
	HTTPRequestTimeout = 10 * time.Second
	// ErrInvalidRSAMethod 表示使用了非 RSA 签名方法
	//
	// ErrInvalidRSAMethod indicates that a non-RSA signing method was used
	ErrInvalidRSAMethod = "invalid RSA signing method"
	// ErrMissingKID 表示缺少令牌头部 KID
	//
	// ErrMissingKID indicates that the KID is missing from the token header
	ErrMissingKID = "missing KID in token header"
	// ErrInvalidToken 表示无效的令牌
	//
	// ErrInvalidToken indicates an invalid token
	ErrInvalidToken = "invalid token"
	// ErrAppleVerification 表示 Apple 令牌验证失败
	//
	// ErrAppleVerification indicates that Apple token verification failed
	ErrAppleVerification = "Apple token verification failed"
)

var (
	// 全局公钥缓存
	//
	// Global public key cache
	globalKeyCache = &keyCache{
		keys:      make(map[string]*rsa.PublicKey),
		fetchTime: time.Time{}, // 零值表示未初始化
	}
	// ErrPublicKeyNotFound 表示 Apple 公钥未找到
	//
	// ErrPublicKeyNotFound indicates that the Apple public key was not found
	ErrPublicKeyNotFound = errors.New("Apple public key not found")
	// ErrFetchKeys 表示获取 Apple 公钥失败
	//
	// ErrFetchKeys indicates that fetching Apple public keys failed
	ErrFetchKeys = errors.New("failed to fetch Apple public keys")
	// ErrDecodeKey 表示解码公钥失败
	//
	// ErrDecodeKey indicates that decoding the public key failed
	ErrDecodeKey = errors.New("failed to decode public key")
	// ErrInvalidKeyFormat 表示无效的公钥格式
	//
	// ErrInvalidKeyFormat indicates an invalid public key format
	ErrInvalidKeyFormat = errors.New("invalid public key format")
)

// keyCache 公钥缓存结构
// keys: 公钥映射（kid -> 公钥）
// fetchTime: 最后获取时间
// mutex: 读写锁
//
// keyCache is a public key cache structure
// keys: Public key mapping (kid -> public key)
// fetchTime: Last fetch time
// mutex: Read-write lock
type keyCache struct {
	keys      map[string]*rsa.PublicKey // 公钥映射（kid -> 公钥）
	fetchTime time.Time                 // 最后获取时间
	mutex     sync.RWMutex              // 读写锁
}

// GetApplePublicKey 获取指定 Kid 的 Apple 公钥
// 如果公钥未缓存或缓存已过期，会自动从 Apple 服务器获取
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - kid: 密钥 ID（Key ID）
//
// 返回:
//   - *rsa.PublicKey: RSA 公钥
//   - error: 如果获取失败，返回错误
//
// GetApplePublicKey retrieves the Apple public key for the specified Kid.
// If the public key is not cached or the cache has expired, it will automatically fetch from Apple servers.
// Parameters:
//   - ctx: Context for controlling request timeout and cancellation
//   - kid: Key ID
//
// Returns:
//   - *rsa.PublicKey: RSA public key
//   - error: Returns an error if retrieval fails
func GetApplePublicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	if kid == "" {
		return nil, fmt.Errorf("%w: key ID (kid) is empty", ErrPublicKeyNotFound)
	}

	// 首先尝试从缓存中读取（读锁）
	globalKeyCache.mutex.RLock()
	key, exists := globalKeyCache.keys[kid]
	isCacheValid := !globalKeyCache.fetchTime.IsZero() &&
		time.Since(globalKeyCache.fetchTime) < KeyCacheTTL
	globalKeyCache.mutex.RUnlock()

	// 如果密钥存在且缓存有效，直接返回
	if exists && isCacheValid {
		return key, nil
	}

	// 缓存无效或密钥不存在，需要刷新缓存（写锁）
	globalKeyCache.mutex.Lock()
	// 双重检查，防止在获取锁的过程中其他协程已经更新了缓存
	isCacheStillValid := !globalKeyCache.fetchTime.IsZero() &&
		time.Since(globalKeyCache.fetchTime) < KeyCacheTTL
	if !isCacheStillValid {
		// 缓存已过期，获取新的公钥
		newKeys, err := FetchApplePublicKeys(ctx)
		if err != nil {
			globalKeyCache.mutex.Unlock()
			return nil, err
		}

		// 更新缓存
		globalKeyCache.keys = newKeys
		globalKeyCache.fetchTime = time.Now()
	}

	// 从更新后的缓存中查找密钥
	key, exists = globalKeyCache.keys[kid]
	globalKeyCache.mutex.Unlock()

	if !exists {
		return nil, fmt.Errorf("%w: kid=%s", ErrPublicKeyNotFound, kid)
	}

	return key, nil
}

// FetchApplePublicKeys 从 Apple 服务器获取最新的公钥
// 返回公钥映射（kid -> 公钥）和可能的错误
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//
// 返回:
//   - map[string]*rsa.PublicKey: 公钥映射（kid -> 公钥）
//   - error: 如果获取失败，返回错误
//
// FetchApplePublicKeys fetches the latest public keys from Apple servers.
// Returns a public key mapping (kid -> public key) and possible errors.
// Parameters:
//   - ctx: Context for controlling request timeout and cancellation
//
// Returns:
//   - map[string]*rsa.PublicKey: Public key mapping (kid -> public key)
//   - error: Returns an error if fetching fails
func FetchApplePublicKeys(ctx context.Context) (map[string]*rsa.PublicKey, error) {
	// 创建带超时的HTTP请求
	reqCtx, cancel := context.WithTimeout(ctx, HTTPRequestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, AppleAuthKeysURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrFetchKeys, err)
	}

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: HTTP request failed: %v", ErrFetchKeys, err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: server returned non-200 status code: %d", ErrFetchKeys, resp.StatusCode)
	}

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrFetchKeys, err)
	}

	// 使用 jwx 库解析 JWK 集合
	keySet, err := jwk.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse JWK: %v", ErrDecodeKey, err)
	}

	// 处理公钥
	keys := make(map[string]*rsa.PublicKey)

	// 遍历 JWK 集合中的所有密钥
	for i := 0; i < keySet.Len(); i++ {
		key, ok := keySet.Key(i)
		if !ok {
			continue
		}

		// 获取 kid
		var kid string
		if err := key.Get("kid", &kid); err != nil || kid == "" {
			continue
		}

		// 只处理 RSA 密钥 - 使用 KeyType() 方法检查密钥类型
		keyType := key.KeyType()
		if keyType != jwa.RSA() {
			continue
		}

		// 转换为原始公钥
		rawKey, err := jwk.PublicRawKeyOf(key)
		if err != nil {
			fmt.Printf("warning: failed to parse public key %s: %v\n", kid, err)
			continue
		}

		// 类型断言为 RSA 公钥
		pubKey, ok := rawKey.(*rsa.PublicKey)
		if !ok {
			fmt.Printf("warning: public key %s is not RSA type, actual type: %T\n", kid, rawKey)
			continue
		}

		keys[kid] = pubKey
	}

	// 检查是否获取到了密钥
	if len(keys) == 0 {
		return nil, fmt.Errorf("%w: no valid public keys found", ErrPublicKeyNotFound)
	}

	return keys, nil
}

// VerifyAppleToken 验证 Apple JWT token 并返回用户标识（Subject）
// 函数会自动获取对应的 Apple 公钥来验证 token 的签名
// 参数:
//   - tokenString: Apple JWT token 字符串
//
// 返回:
//   - string: token 中的用户标识（Subject）
//   - error: 如果验证失败，返回错误
//
// VerifyAppleToken verifies an Apple JWT token and returns the user identifier (Subject).
// The function automatically fetches the corresponding Apple public key to verify the token signature.
// Parameters:
//   - tokenString: Apple JWT token string
//
// Returns:
//   - string: User identifier (Subject) from the token
//   - error: Returns an error if verification fails
func VerifyAppleToken(tokenString string) (string, error) {
	// 创建标准JWT声明结构
	claims := &jwt.RegisteredClaims{}

	// 解析令牌
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%s: %v", ErrInvalidRSAMethod, token.Method)
		}

		// 提取密钥ID
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("%s", ErrMissingKID)
		}

		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 获取Apple公钥
		pubKey, err := GetApplePublicKey(ctx, kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get Apple public key: %w", err)
		}

		return pubKey, nil
	})

	// 处理解析错误
	if err != nil {
		return "", fmt.Errorf("%s: %w", ErrAppleVerification, err)
	}

	// 验证令牌有效性
	if !token.Valid {
		return "", fmt.Errorf("%s", ErrInvalidToken)
	}

	return claims.Subject, nil
}
