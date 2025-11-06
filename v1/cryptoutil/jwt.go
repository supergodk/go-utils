package cryptoutil

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	// KeyAlgorithmEdDSA EdDSA 算法（Ed25519）
	//
	// KeyAlgorithmEdDSA is the EdDSA algorithm (Ed25519)
	KeyAlgorithmEdDSA = "EdDSA"
	// KeyAlgorithmRS256 RS256 算法（RSA）
	//
	// KeyAlgorithmRS256 is the RS256 algorithm (RSA)
	KeyAlgorithmRS256 = "RS256"
	// KeyAlgorithmES256 ES256 算法（ECDSA）
	//
	// KeyAlgorithmES256 is the ES256 algorithm (ECDSA)
	KeyAlgorithmES256 = "ES256"
)

// GenerateTokenOptions JWT token 生成选项
// PrivateKey: 私钥字节数组（PKCS#8 格式或 Ed25519 原始密钥）
// TokenIssuer: Token 发行者标识
// TokenExpireDuration: Token 过期时长
// ClaimMap: 自定义声明映射（键值对）
// KeyAlgorithm: 密钥算法字符串，应使用常量：KeyAlgorithmEdDSA、KeyAlgorithmRS256、KeyAlgorithmES256
//
// GenerateTokenOptions contains options for generating JWT tokens.
// PrivateKey: Private key bytes (PKCS#8 format or Ed25519 raw key)
// TokenIssuer: Token issuer identifier
// TokenExpireDuration: Token expiration duration
// ClaimMap: Custom claims mapping (key-value pairs)
// KeyAlgorithm: Key algorithm string, should use constants: KeyAlgorithmEdDSA, KeyAlgorithmRS256, KeyAlgorithmES256
type GenerateTokenOptions struct {
	PrivateKey          []byte
	TokenIssuer         string
	TokenExpireDuration time.Duration
	ClaimMap            map[string]any
	KeyAlgorithm        string
}

// JwtGenerator JWT token 生成器
// 使用结构体字段作为默认配置，可以通过 GenerateToken 方法的 options 参数覆盖
// PrivateKey: 私钥字节数组（PKCS#8 格式或 Ed25519 原始密钥）
// TokenIssuer: Token 发行者标识
// TokenExpireDuration: Token 过期时长
// ClaimMap: 自定义声明映射（键值对）
// KeyAlgorithm: 密钥算法字符串，应使用常量：KeyAlgorithmEdDSA、KeyAlgorithmRS256、KeyAlgorithmES256
//
// JwtGenerator is a JWT token generator.
// Uses struct fields as default configuration, which can be overridden by the options parameter in GenerateToken method.
// PrivateKey: Private key bytes (PKCS#8 format or Ed25519 raw key)
// TokenIssuer: Token issuer identifier
// TokenExpireDuration: Token expiration duration
// ClaimMap: Custom claims mapping (key-value pairs)
// KeyAlgorithm: Key algorithm string, should use constants: KeyAlgorithmEdDSA, KeyAlgorithmRS256, KeyAlgorithmES256
type JwtGenerator struct {
	PrivateKey          []byte
	TokenIssuer         string
	TokenExpireDuration time.Duration
	ClaimMap            map[string]any
	KeyAlgorithm        string
}

// GenerateToken 使用 Builder 模式生成 JWT token
// 如果 options 参数为 nil，则使用 JwtGenerator 结构体的字段作为默认配置
// 支持的签名算法：EdDSA (Ed25519)、RS256 (RSA)、ES256 (ECDSA)
// KeyAlgorithm 字段应使用预定义常量：KeyAlgorithmEdDSA、KeyAlgorithmRS256、KeyAlgorithmES256
// 参数:
//   - options: Token 生成选项，如果为 nil 则使用结构体默认配置
//
// 返回:
//   - string: 生成的 JWT token 字符串
//   - error: 如果生成失败，返回错误
//
// GenerateToken generates a JWT token using the Builder pattern.
// If the options parameter is nil, it uses the JwtGenerator struct fields as default configuration.
// Supported signing algorithms: EdDSA (Ed25519), RS256 (RSA), ES256 (ECDSA)
// The KeyAlgorithm field should use predefined constants: KeyAlgorithmEdDSA, KeyAlgorithmRS256, KeyAlgorithmES256
// Parameters:
//   - options: Token generation options, if nil, uses struct default configuration
//
// Returns:
//   - string: Generated JWT token string
//   - error: Returns an error if generation fails
func (j *JwtGenerator) GenerateToken(options *GenerateTokenOptions) (string, error) {

	if options == nil {
		options = &GenerateTokenOptions{
			PrivateKey:          j.PrivateKey,
			TokenIssuer:         j.TokenIssuer,
			TokenExpireDuration: j.TokenExpireDuration,
			ClaimMap:            j.ClaimMap,
			KeyAlgorithm:        j.KeyAlgorithm,
		}
	}

	// 使用 Builder 创建 token
	tokenBuilder := jwt.NewBuilder().
		Issuer(options.TokenIssuer).
		IssuedAt(time.Now()).
		NotBefore(time.Now()).
		Expiration(time.Now().Add(options.TokenExpireDuration))
	for key, value := range options.ClaimMap {
		tokenBuilder = tokenBuilder.Claim(key, value)
	}
	token, err := tokenBuilder.Build()
	if err != nil {
		return "", err
	}

	var signed []byte
	switch options.KeyAlgorithm {
	case KeyAlgorithmEdDSA:
		signed, err = jwt.Sign(token, jwt.WithKey(jwa.EdDSA(), ed25519.PrivateKey(options.PrivateKey)))
	case KeyAlgorithmRS256:
		privKey, err := x509.ParsePKCS8PrivateKey(options.PrivateKey)
		if err != nil {
			return "", err
		}
		signed, err = jwt.Sign(token, jwt.WithKey(jwa.RS256(), privKey.(*rsa.PrivateKey)))
	case KeyAlgorithmES256:
		privKey, err := x509.ParsePKCS8PrivateKey(options.PrivateKey)
		if err != nil {
			return "", err
		}
		signed, err = jwt.Sign(token, jwt.WithKey(jwa.ES256(), privKey.(*ecdsa.PrivateKey)))
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported key algorithm: %s", options.KeyAlgorithm)
	}

	return string(signed), nil
}
