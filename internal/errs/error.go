package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("")
)

func NewErrUnSupportType(expr any) error {
	return fmt.Errorf("model: 不支持类型 %v", expr)
}

func NewErrUnSupportExpression(expr any) error {
	return fmt.Errorf("orm: 不支持的表达式类型 %v", expr)
}

func NewErrUnKnownField(fd string) error {
	return fmt.Errorf("orm: 未知字段 %s", fd)
}
