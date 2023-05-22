package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly   = errors.New("")
	ErrNoRows        = errors.New("orm: 没有数据")
	ErrAliasWhere    = errors.New("orm: where条件不能用别名")
	ErrInsertZeroRow = errors.New("orm: 插入0行")
)

func NewErrUnSupportType(expr any) error {
	return fmt.Errorf("Model: 不支持类型 %v", expr)
}

func NewErrUnSupportExpression(expr any) error {
	return fmt.Errorf("orm: 不支持的表达式类型 %v", expr)
}

func NewErrUnKnownField(fd string) error {
	return fmt.Errorf("orm: 未知字段 %s", fd)
}

func NewErrUnKnownColumn(fd string) error {
	return fmt.Errorf("orm: 未知列 %s", fd)
}

func NewErrInvalidTagContent(tag string) error {
	return fmt.Errorf("orm: 错误的标签设置: %s", tag)
}

func NewErrUnSupportAssignable(expr any) error {
	return fmt.Errorf("orm: 不支持的赋值表达式类型: %v", expr)
}
