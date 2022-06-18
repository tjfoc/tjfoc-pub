package stack

func New() *Stack {
	return &Stack{
		count: 0,
		stack: nil,
	}
}

//清空
func (a *Stack) Reset() {
	a.count = 0
	a.stack = nil
}

//返回栈元素数目
func (a *Stack) Count() int {
	return a.count
}

//查看栈顶元素
func (a *Stack) Top() StackNode {
	switch a.count {
	case 0:
		return nil
	default:
		return a.stack[a.count-1].node
	}
}

func (a *Stack) Pop() StackNode {
	switch a.count {
	case 0:
		return nil
	default:
		a.count--
		return a.stack[a.count].node
	}
}

func (a *Stack) Push(b StackNode) {
	switch a.stack {
	case nil:
		a.stack = append(a.stack, &stackNode{b})
	default:
		a.stack = append(a.stack[:a.count], &stackNode{b})
	}
	a.count++
}
