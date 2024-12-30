package util

import "github.com/pkg/errors"

// Stack 定义栈结构
type Stack struct {
	items []interface{}
}

func NewStack() *Stack {
	return &Stack{}
}

// Push 向栈中添加元素
func (s *Stack) Push(item interface{}) {
	s.items = append(s.items, item)
}

// Pop 从栈中移除并返回栈顶元素
func (s *Stack) Pop() (interface{}, error) {
	// 检查栈是否为空
	if len(s.items) == 0 {
		return nil, errors.New("stack is empty")
	}
	// 获取栈顶元素并移除
	index := len(s.items) - 1
	item := s.items[index]
	s.items = s.items[:index]
	return item, nil
}

// Peek 查看栈顶元素，但不移除它
func (s *Stack) Peek() (interface{}, error) {
	if len(s.items) == 0 {
		return nil, errors.New("stack is empty")
	}
	return s.items[len(s.items)-1], nil
}

// IsEmpty 判断栈是否为空
func (s *Stack) IsEmpty() bool {
	return len(s.items) == 0
}

// Size 获取栈的大小
func (s *Stack) Size() int {
	return len(s.items)
}
