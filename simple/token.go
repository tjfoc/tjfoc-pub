package main

const MAX_TOKENSIZE int = 100

type TokenKind int

type TokenType int

type StateType int

type FlowType int

const (
	BAD_TOKEN           TokenKind = iota //意外的标识符
	NUMBER_TOKEN                         //数字标识符
	CHAR_TOKEN                           //字符标识符
	CHAR_SIGN_TOKEN                      //单引号'标识符
	STRING_TOKEN                         //字符串标识符
	STRING_SIGN_TOKEN                    //双引号"标识符
	BOOL_TOKEN                           //布尔标识符
	IF_TOKEN                             //if标识符
	ELSE_TOKEN                           //else标识符
	FOR_TOKEN                            //for标识符
	FUNC_TOKEN                           //func标识符
	PARAM_TOKEN                          //变量标识符
	STATE_TOKEN                          //声明变量标识符
	STATE_TYPE_TOKEN                     //声明类型标识符let,set
	TOKEN_TYPE_TOKEN                     //变量类型标识符int,float...
	ADD_OPERATOR_TOKEN                   //加法
	SUB_OPERATOR_TOKEN                   //减法
	MUL_OPERATOR_TOKEN                   //乘法
	DIV_OPERATOR_TOKEN                   //除法
	MOD_OPERATOR_TOKEN                   //取余
	ASS_OPERATOR_TOKEN                   //赋值
	LEFT_PAREN_TOKEN                     //左括号
	RIGHT_PAREN_TOKEN                    //右括号
	LEFT_BRACKET_TOKEN                   //左中括号
	RIGHT_BRACKET_TOKEN                  //右中括号
	LEFT_BRACES_TOKEN                    //左大括号
	RIGHT_BRACES_TOKEN                   //右大括号
	END_OF_LINE_TOKEN                    //行结束符
	EQ_TOKEN                             // ==
	NE_TOKEN                             // !=
	GT_TOKEN                             // >
	GE_TOKEN                             // >=
	LT_TOKEN                             // <
	LE_TOKEN                             // <=
	LOGICAL_AND_TOKEN                    // &&
	LOGICAL_OR_TOKEN                     // ||

)

const (
	INT8 TokenType = iota
	INT16
	INT32
	INT64
	UINT8
	UINT16
	UINT32
	UINT64
	FLOAT32
	FLOAT64
	BOOL
	STRING
	CHAR
	ERRORTYPE
)

const (
	LET StateType = iota
	SET
)

const (
	IF FlowType = iota
	ELSE
	ELSEIF
)

var KeyWords = []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "bool", "float32", "float64", "string", "char"}
var FlowWords = []string{"if", "else", "else if", "for"}
var StatementWords = []string{"let", "set"} //set写入数据库,let不写

type Token struct {
	Kind      TokenKind   `json:"kind"`
	Value     interface{} `json:"value"`
	Str       string      `json:"str"`
	TokenType TokenType   `json:"tokenType"`
	StateType StateType   `json:"stateType"`
}

type ForToken struct {
	Condition1 []rune   `json:"condition1"`
	Condition2 []rune   `json:"condition2"`
	Condition3 []rune   `json:"condition3"`
	Body       [][]rune `json:"condition"`
}

type ArrayListToken struct {
	StateType StateType     `json:"stateType"`
	TokenType TokenType     `json:"tokenType"`
	Str       string        `json:"str"`
	Boundary  []int         `json:"boundary"`
	Dimension int           `json:"dimension"`
	Value     []interface{} `json:"value"`
}

type MapListToken struct {
	StateType StateType   `json:"stateType"`
	KeyType   TokenType   `json:"KeyType"`
	ValueType TokenType   `json:"ValueType"`
	Str       string      `json:"str"`
	Dimension int         `json:"dimension"`
	Value     interface{} `json:"value"`
}
