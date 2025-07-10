package functions

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
)

// RandomFunction はランダム数値生成関数
type RandomFunction struct{}

func (f *RandomFunction) Name() string {
	return "random"
}

func (f *RandomFunction) Signature() string {
	return "!random <max>"
}

func (f *RandomFunction) Description() string {
	return "0からmax-1までのランダムな整数を生成します"
}

func (f *RandomFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("random function expects 1 argument, got %d", len(args))
	}

	var max int64
	switch v := args[0].(type) {
	case int:
		max = int64(v)
	case int64:
		max = v
	case float64:
		max = int64(v)
	default:
		return nil, fmt.Errorf("random function expects numeric argument")
	}

	if max <= 0 {
		return nil, fmt.Errorf("random function expects positive number")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return nil, fmt.Errorf("random generation failed: %w", err)
	}

	return n.Int64(), nil
}

// RandomStringFunction はランダム文字列生成関数
type RandomStringFunction struct{}

func (f *RandomStringFunction) Name() string {
	return "random_string"
}

func (f *RandomStringFunction) Signature() string {
	return "!random_string <length> [charset]"
}

func (f *RandomStringFunction) Description() string {
	return "指定した長さのランダムな文字列を生成します（オプションで文字セットを指定可能）"
}

func (f *RandomStringFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, fmt.Errorf("random_string function expects 1 or 2 arguments, got %d", len(args))
	}

	var length int64
	switch v := args[0].(type) {
	case int:
		length = int64(v)
	case int64:
		length = v
	case float64:
		length = int64(v)
	default:
		return nil, fmt.Errorf("random_string function expects numeric length argument")
	}

	if length <= 0 {
		return nil, fmt.Errorf("random_string function expects positive length")
	}

	// デフォルトの文字セット
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// カスタム文字セットが指定された場合
	if len(args) == 2 {
		if customCharset, ok := args[1].(string); ok {
			charset = customCharset
		} else {
			return nil, fmt.Errorf("random_string function expects string charset argument")
		}
	}

	if len(charset) == 0 {
		return nil, fmt.Errorf("charset cannot be empty")
	}

	var result strings.Builder
	charsetLen := big.NewInt(int64(len(charset)))

	for i := int64(0); i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return nil, fmt.Errorf("random string generation failed: %w", err)
		}
		result.WriteByte(charset[n.Int64()])
	}

	return result.String(), nil
}

// UUIDFunction はUUID生成関数
type UUIDFunction struct{}

func (f *UUIDFunction) Name() string {
	return "uuid"
}

func (f *UUIDFunction) Signature() string {
	return "!uuid"
}

func (f *UUIDFunction) Description() string {
	return "ランダムなUUID（v4）を生成します"
}

func (f *UUIDFunction) Execute(_ context.Context, args []interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("uuid function expects no arguments, got %d", len(args))
	}

	id := uuid.New()
	return id.String(), nil
}
