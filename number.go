package kns

import "fmt"

type Number struct {
	NumberSequenceID string
	No               int
}

func (n *Number) Format(pattern string) string {
	return fmt.Sprintf("%d", n.No)
}

func (n *Number) String() string {
	return fmt.Sprintf("%d", n.No)
}
