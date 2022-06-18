package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/tjfoc/tjfoc/core/store"
)

var db interface{}
var st_look_ahead_token Token
var st_look_ahead_token_exists int

var stl_bak int = 0

var inBody int = 0
var bodyLine int = 0

/*0表示不存在,1表示存在,2表示存在且逻辑表达式为true,3表示逻辑表达式为false*/
var if_type int = 0
var else_type int = 0

//var if_st_lines [][]rune

var paramList map[string]Token               //变量列表(全局)
var arrayParamList map[string]ArrayListToken //数组列表(全局)

var dbParamList map[string]Token //数据库变量列表

var lvNum int = 0                           //局部变量层数
var localParamList map[int]map[string]Token //局部变量列表

var forNum int = 0 //for嵌套层数
var forTokenList map[int]ForToken

var fi *os.File
var inputReader *bufio.Reader

func my_get_token(token *Token) {
	if st_look_ahead_token_exists == 1 {
		*token = st_look_ahead_token
		st_look_ahead_token_exists = 0
	} else {
		getToken(token)
	}
}

func unget_token(token *Token) {
	st_look_ahead_token = *token
	st_look_ahead_token_exists = 1
}

func parse_primary_expression() interface{} {
	var token Token
	var value interface{}
	var minus_flages int = 0
	fortoken, _ := forTokenList[forNum]

	my_get_token(&token)
	if token.Kind == SUB_OPERATOR_TOKEN {
		minus_flages = 1
	} else {
		unget_token(&token)
	}

	my_get_token(&token)
	fmt.Println(token.Str)

	//判断是否声明变量
	if token.Kind == STATE_TYPE_TOKEN {
		state_token := token
		my_get_token(&token)
		//获取变量名，放入map
		if token.Kind == STATE_TOKEN {
			if state_token.Str == "let" {
				token.StateType = LET
			} else if state_token.Str == "set" {
				token.StateType = SET
			}
			stk := token
			my_get_token(&token)
			//获取变量类型
			if token.Kind == TOKEN_TYPE_TOKEN {
				stk.TokenType = getTokenType(token.Str)
				if stk.TokenType == ERRORTYPE {
					fmt.Println("error type : ", token.Str)
					os.Exit(1)
				}

				my_get_token(&token)

				//变量后续是否为赋值操作
				if token.Kind == ASS_OPERATOR_TOKEN {
					value = parse_expression()
					value = getValue(stk.TokenType, value, minus_flages)
					if lvNum > 0 {
						//局部变量还未写入数据库
						localParam, _ := localParamList[lvNum]
						stk.Value = value
						localParam[stk.Str] = stk
						localParamList[lvNum] = localParam
						if stk.StateType == SET {
							key := strconv.Itoa(lvNum) + "_" + stk.Str
							dbParamList[key] = stk
						}
					} else {
						if _, ok := arrayParamList[stk.Str]; ok {
							fmt.Println("error the variable is existed : ", stk.Str)
							os.Exit(1)
						} else if _, ok := paramList[stk.Str]; ok {
							fmt.Println("error the variable is existed : ", stk.Str)
							os.Exit(1)
						} else {
							stk.Value = value
							paramList[stk.Str] = stk
							if stk.StateType == SET {
								dbParamList[stk.Str] = stk
							}
						}
					}
				} else {
					unget_token(&token)
					paramList[stk.Str] = stk
					return value
				}
			} else if token.Kind == LEFT_BRACKET_TOKEN { //变量后续是否为数组标签[
				var arr ArrayListToken
				value = parse_expression()
				boundary := []int{int(value.(float64))}
				arr.Str = stk.Str
				arr.Dimension = 1
				if _, ok := paramList[stk.Str]; ok {
					fmt.Println("error the variable is existed : ", stk.Str)
					os.Exit(1)
				} else if _, ok := arrayParamList[stk.Str]; ok {
					fmt.Println("error the variable is existed : ", stk.Str)
					os.Exit(1)
				}
				for {
					my_get_token(&token)
					if token.Kind != RIGHT_BRACKET_TOKEN {
						fmt.Println("error Miss ] ")
						os.Exit(1)
					}
					my_get_token(&token)
					if token.Kind == LEFT_BRACKET_TOKEN {
						value = parse_expression()
						boundary = append(boundary, int(value.(float64)))
						arr.Dimension = arr.Dimension + 1
					} else {
						if token.Kind == TOKEN_TYPE_TOKEN {
							stk.TokenType = getTokenType(token.Str)
							if stk.TokenType == ERRORTYPE {
								fmt.Println("error type : ", token.Str)
								os.Exit(1)
							}
						} else {
							unget_token(&token)
							fmt.Println("error Miss Type ")
						}
						break
					}
				}
				arr.Boundary = boundary
				arr.TokenType = stk.TokenType
				arr.StateType = stk.StateType

				my_get_token(&token)
				//数组后续是否为赋值操作
				if token.Kind == ASS_OPERATOR_TOKEN {
					arrValue := []rune(st_line[st_line_pos : len(st_line)-1])
					fmt.Println("------", string(arrValue), st_line_pos, "------")
					arrValue = []rune(strings.TrimLeft(string(arrValue[:]), " "))
					var spliStr string = "" //数组分割字符
					for i, v := range arrValue {
						if v != '{' {
							fmt.Println("------", i, "------")
							if i != arr.Dimension {
								fmt.Println("error Dimension error!")
								os.Exit(1)
							}
							for ; i > 1; i-- {
								spliStr = spliStr + "}"
							}
							spliStr = spliStr + ","
							break
						}
					}
					fmt.Println("------", spliStr, "------")
					strs := strings.Split(string(arrValue), spliStr)
					fmt.Println("------", strs, len(strs), "------")
					if len(strs) == arr.Boundary[0] {
						for i := 1; ; i++ {
							if spliStr == "," {
								break
							}
							spliStr = string([]rune(spliStr)[1:])
							fmt.Println("------", spliStr, "------")
							strs = strings.Split(strs[0], spliStr)
							fmt.Println("------", strs, len(strs), "------")
							if len(strs) != arr.Boundary[i] {
								fmt.Println("error Boundary error!")
								os.Exit(1)
							}
						}
					}
					valueStr := strings.Replace(string(arrValue), "{", "", -1)
					valueStr = strings.Replace(valueStr, "}", "", -1)
					for _, v := range strings.Split(valueStr, ",") {
						if arr.TokenType >= INT8 && arr.TokenType <= FLOAT64 {
							val, _ := strconv.ParseFloat(v, 64)
							value = float64(val)
							value = getValue(arr.TokenType, value, 0)
						} else if arr.TokenType == BOOL {
							if v == "true" {
								value = true
							} else if v == "false" {
								value = false
							}
						} else if arr.TokenType == STRING || arr.TokenType == CHAR {
							value = string([]rune(v)[1 : len(v)-1])
							value = getValue(arr.TokenType, value, 0)
						}
						arr.Value = append(arr.Value, value)
					}

					fmt.Println("------", valueStr, arr, st_line_pos, "------")
				} else {
					unget_token(&token)
					var num int = 1
					for _, v := range arr.Boundary {
						num = num * v
					}
					for index := 0; index < num; index++ {
						if arr.TokenType >= INT8 && arr.TokenType <= FLOAT64 {
							value = float64(0)
							value = getValue(arr.TokenType, value, 0)
						} else if arr.TokenType == BOOL {
							value = false
						} else if arr.TokenType == STRING || arr.TokenType == CHAR {
							value = ""
							value = getValue(arr.TokenType, value, 0)
						}
						arr.Value = append(arr.Value, value)
					}
				}
				fmt.Println("------", arr, st_line_pos, "------")
				arrayParamList[arr.Str] = arr
			}
		}
	}

	/*获取变量，返回value*/
	if token.Kind == STATE_TOKEN {
		var tk Token
		var arr ArrayListToken
		var isArr int = 0
		num := lvNum
		//判断是否已经声明
		for {
			if num == 0 {
				if ltk, ok := paramList[token.Str]; ok {
					tk = ltk
					break
				} else if ltk, ok := arrayParamList[token.Str]; ok {
					arr = ltk
					isArr = 1
					break
				} else {
					fmt.Println("an undeclared variable : ", token.Str)
					os.Exit(1)
				}
				break
			}
			localParam, _ := localParamList[num]
			if ltk, ok := localParam[token.Str]; ok {
				tk = ltk
				break
			}
			num = num - 1
		}

		my_get_token(&token)
		//变量后续是否为赋值操作
		if token.Kind == ASS_OPERATOR_TOKEN {
			if isArr == 0 {
				value = parse_expression()
				//value = getValue(tk.tokenType, value, minus_flages)
				tk.Value = getValue(tk.TokenType, value, minus_flages)
				if num == 0 {
					paramList[tk.Str] = tk
					if tk.StateType == SET {
						dbParamList[tk.Str] = tk
					}
				} else {
					localParam, _ := localParamList[num]
					//局部变量暂时没有写入数据库
					localParam[token.Str] = tk
					localParamList[num] = localParam
					if tk.StateType == SET {
						key := strconv.Itoa(lvNum) + "_" + tk.Str
						dbParamList[key] = tk
					}
				}
			} else if isArr == 1 {
				arrValue := []rune(st_line[st_line_pos+1 : len(st_line)-1])
				fmt.Println("------", string(arrValue), st_line_pos, "------")
				var spliStr string = "" //数组分割字符
				for i, v := range arrValue {
					if v != '{' {
						if i != arr.Dimension {
							fmt.Println("error Dimension error!")
							os.Exit(1)
						}
						for ; i > 1; i-- {
							spliStr = spliStr + "}"
						}
						spliStr = spliStr + ","
						break
					}
				}
				strs := strings.Split(string(arrValue), spliStr)
				if len(strs) == arr.Boundary[0] {
					for i := 1; ; i++ {
						if spliStr == "," {
							break
						}
						spliStr = string([]rune(spliStr)[1:])
						strs = strings.Split(strs[0], spliStr)
						if len(strs) != arr.Boundary[i] {
							fmt.Println("error Boundary error!")
							os.Exit(1)
						}
					}
				}
				valueStr := strings.Replace(string(arrValue), "{", "", -1)
				valueStr = strings.Replace(valueStr, "}", "", -1)
				for i, v := range strings.Split(valueStr, ",") {
					if arr.TokenType >= INT8 && arr.TokenType <= FLOAT64 {
						val, _ := strconv.ParseFloat(v, 64)
						value = float64(val)
						value = getValue(arr.TokenType, value, 0)
					} else if arr.TokenType == BOOL {
						if v == "true" {
							value = true
						} else if v == "false" {
							value = false
						}
					} else if arr.TokenType == STRING || arr.TokenType == CHAR {
						value = string([]rune(v)[1 : len(v)-1])
						value = getValue(arr.TokenType, value, 0)
					}
					arr.Value[i] = value
				}

				fmt.Println("------", valueStr, arr, st_line_pos, "------")
			}
		} else if token.Kind == LEFT_BRACKET_TOKEN {
			value = parse_expression()
			boundary := []int{int(value.(float64))}
			for {
				my_get_token(&token)
				if token.Kind != RIGHT_BRACKET_TOKEN {
					fmt.Println("error Miss ] ")
					os.Exit(1)
				}
				my_get_token(&token)
				if token.Kind == LEFT_BRACKET_TOKEN {
					value = parse_expression()
					boundary = append(boundary, int(value.(float64)))
				} else {
					unget_token(&token)
					break
				}
			}
			if token.Kind == ASS_OPERATOR_TOKEN {
				my_get_token(&token)
				value = parse_expression()

				var stlen int = 0
				var cup int = 1
				for i, v1 := range boundary {
					cup = v1
					for j := i + 1; j < arr.Dimension; j++ {
						cup = cup * arr.Boundary[j]
					}
					stlen = stlen + cup
				}
				fmt.Println("arr num : ", stlen)
				if len(boundary) == arr.Dimension {
					arr.Value[stlen] = value
				} else if len(boundary) < arr.Dimension {
					valueLen := 1
					for i := arr.Dimension - len(boundary); i > 0; i-- {
						valueLen = valueLen * arr.Boundary[arr.Dimension-i]
					}
					//fmt.Println("*********************", value, "*************************")
					j := 0
					for i := stlen; i < (stlen + valueLen - 1); i++ {
						//fmt.Println("*********************", i, j, "*************************")
						arr.Value[i] = value.([]interface{})[j]
						j++
					}
					//fmt.Println("*********************", value, "*************************")
				} else if len(boundary) > arr.Dimension {
					fmt.Println("error to many []")
					os.Exit(1)
				}
			} else {
				//arrlen := len(arr.Value)
				var stlen int = 0
				var cup int = 1
				for i, v1 := range boundary {
					cup = v1
					for j := i + 1; j < arr.Dimension; j++ {
						cup = cup * arr.Boundary[j]
					}
					stlen = stlen + cup
				}
				fmt.Println("arr num : ", stlen)
				if len(boundary) == arr.Dimension {
					value = arr.Value[stlen]
				} else if len(boundary) < arr.Dimension {
					valueLen := 1
					for i := arr.Dimension - len(boundary); i > 0; i-- {
						valueLen = valueLen * arr.Boundary[arr.Dimension-i]
					}
					fmt.Println("*********************", valueLen, "*************************")
					value = arr.Value[stlen : stlen+valueLen-1]
					fmt.Println("*********************", value, "*************************")
				} else if len(boundary) > arr.Dimension {
					fmt.Println("error to many []")
					os.Exit(1)
				}
			}
		} else {
			unget_token(&token)
			if isArr == 1 {
				value = arr.Value
				fmt.Println("Value : ", value)
			} else if isArr == 0 {
				value = tk.Value
				fmt.Println("tk.Value : ", tk.Value)
			}
		}

	}
	//如果是else
	if token.Kind == ELSE_TOKEN {
		value = true
		if else_type == 2 && if_type == 3 {
			lvNum = lvNum + 1
			localParam := make(map[string]Token)
			localParamList[lvNum] = localParam
			my_get_token(&token)
			if token.Kind == LEFT_BRACES_TOKEN {
				if inBody == 1 {
					getCodeInBody(&fortoken)
				} else {
					getCode()
				}
				else_type = 0
				if_type = 0
			} else if token.Kind == IF_TOKEN {
				else_type = 0
			}
			delete(localParamList, lvNum)
			lvNum = lvNum - 1
		} else if else_type == 2 && if_type == 0 {
			fmt.Println("Miss if error")
			os.Exit(1)
		} else if else_type == 0 {
			if inBody == 1 {
				skipCodeInBody(&fortoken)
			} else {
				skipCode()
			}
		}
	}
	//如果是if
	if token.Kind == IF_TOKEN {
		if_type = 1
		my_get_token(&token)
		//fmt.Println("if token", token.Str)
		if token.Kind == LEFT_PAREN_TOKEN {
			value = parse_logic_expression()
			my_get_token(&token)
			if token.Kind != RIGHT_PAREN_TOKEN {
				fmt.Println("missing ')' error.")
				os.Exit(1)
			}
			if value.(bool) {
				lvNum = lvNum + 1
				localParam := make(map[string]Token)
				localParamList[lvNum] = localParam
				//fmt.Println("if true")
				if_type = 2
				my_get_token(&token)
				if token.Kind == LEFT_BRACES_TOKEN {
					if inBody == 1 {
						getCodeInBody(&fortoken)
					} else {
						getCode()
					}
				}
				if_type = 0
				delete(localParamList, lvNum)
				lvNum = lvNum - 1
			} else {
				if_type = 3
				else_type = 2
				//fmt.Println("if false")
				if inBody == 1 {
					skipCodeInBody(&fortoken)
				} else {
					skipCode()
				}
			}
		} else {
			fmt.Println("missing '(' error.")
			os.Exit(1)
		}
	}
	//如果是for
	if token.Kind == FOR_TOKEN {

		forNum++
		var forToken ForToken
		var bodyCode [][]rune
		//fmt.Println("for")
		value = 0.0
		condition := getCondition()
		forToken.Condition1 = condition[0]
		forToken.Condition2 = condition[1]
		forToken.Condition3 = condition[2]

		if inBody == 1 {
			i := forNum - 1
			ftoken, _ := forTokenList[i]
			bodyCode = getBodyCodeInBody(&ftoken)
		} else {
			bodyCode = getBodyCode()
		}

		forToken.Body = bodyCode
		forTokenList[forNum] = forToken
		runFor(&forToken, forNum)
		inBody = 0

	}
	//如果是常量
	if token.Kind == NUMBER_TOKEN {
		value = token.Value
	} else if token.Kind == LEFT_PAREN_TOKEN {
		value = parse_expression()
		my_get_token(&token)
		if token.Kind != RIGHT_PAREN_TOKEN {
			fmt.Println("missing ')' error.")
			os.Exit(1)
		}
	} else if token.Kind == CHAR_TOKEN || token.Kind == STRING_TOKEN {
		value = string([]rune(token.Str)[1 : len(token.Str)-1])
		fmt.Println(token.Str, " ----------- ", value.(string))

	} else if token.Kind == BOOL_TOKEN {
		if token.Str == "true" {
			value = true
		} else if token.Str == "false" {
			value = false
		}
	} else {
		unget_token(&token)
	}
	//fmt.Println(reflect.ValueOf(value), "************")
	if reflect.TypeOf(value).String() == "float32" || reflect.TypeOf(value).String() == "float64" {
		value = reflect.ValueOf(value).Float()
		if minus_flages == 1 {
			value = -reflect.ValueOf(value).Float()
		}
	} else if reflect.TypeOf(value).String() == "int8" || reflect.TypeOf(value).String() == "int16" || reflect.TypeOf(value).String() == "int32" || reflect.TypeOf(value).String() == "int64" {
		value = reflect.ValueOf(value).Int()
		if minus_flages == 1 {
			value = -reflect.ValueOf(value).Int()
		}
		value = float64(value.(int64))
	} else if reflect.TypeOf(value).String() == "uint8" || reflect.TypeOf(value).String() == "uint16" || reflect.TypeOf(value).String() == "uint32" || reflect.TypeOf(value).String() == "uint64" {
		value = reflect.ValueOf(value).Uint()
		if minus_flages == 1 {
			value = -reflect.ValueOf(value).Uint()
		}
		value = float64(value.(uint64))
	}
	return value
}

func parse_term() interface{} {
	var v1 interface{}
	var v2 interface{}
	var token Token

	v1 = parse_primary_expression()
	for {
		my_get_token(&token)
		if token.Kind != DIV_OPERATOR_TOKEN && token.Kind != MUL_OPERATOR_TOKEN && token.Kind != MOD_OPERATOR_TOKEN {
			unget_token(&token)
			break
		}
		v2 = parse_primary_expression()

		if reflect.TypeOf(v1).String() == reflect.TypeOf(v2).String() {
			if token.Kind == MUL_OPERATOR_TOKEN {
				if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
					v1 = reflect.ValueOf(v1).Float() * reflect.ValueOf(v2).Float()
				} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
					v1 = reflect.ValueOf(v1).Int() * reflect.ValueOf(v2).Int()
				} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
					v1 = reflect.ValueOf(v1).Uint() * reflect.ValueOf(v2).Uint()
				} else if reflect.TypeOf(v1).String() == "rune" {
					v1 = v1.(rune) * v2.(rune)
				} else {
					fmt.Println("These Type can not add ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				}
			} else if token.Kind == DIV_OPERATOR_TOKEN {
				if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
					v1 = reflect.ValueOf(v1).Float() / reflect.ValueOf(v2).Float()
				} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
					v1 = reflect.ValueOf(v1).Int() / reflect.ValueOf(v2).Int()
				} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
					v1 = reflect.ValueOf(v1).Uint() / reflect.ValueOf(v2).Uint()
				} else if reflect.TypeOf(v1).String() == "rune" {
					v1 = v1.(rune) / v2.(rune)
				} else {
					fmt.Println("These Type can not sub ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				}
			} else if token.Kind == MOD_OPERATOR_TOKEN {
				//fmt.Println("MOD_OPERATOR_TOKEN")
				if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
					v1 = int64(reflect.ValueOf(v1).Float()) % int64(reflect.ValueOf(v2).Float())
				} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
					v1 = reflect.ValueOf(v1).Int() % reflect.ValueOf(v2).Int()
				} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
					v1 = reflect.ValueOf(v1).Uint() % reflect.ValueOf(v2).Uint()
				} else {
					fmt.Println("These Type can not mod ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				}
				v1 = float64(v1.(int64))
			} else {
				unget_token(&token)
			}
		} else {
			if token.Kind == MOD_OPERATOR_TOKEN {
				if !strings.Contains(reflect.ValueOf(v1).String(), ".") && !strings.Contains(reflect.ValueOf(v2).String(), ".") {
					if reflect.TypeOf(v1).String() == "float64" {
						v1 = int64(v1.(float64))
					} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
						v1 = reflect.ValueOf(v1).Int()
					} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
						v1 = reflect.ValueOf(v1).Uint()
						v1 = int64(v1.(uint64))
					}
					if reflect.TypeOf(v2).String() == "float64" {
						v2 = int64(v2.(float64))
					} else if reflect.TypeOf(v2).String() == "int8" || reflect.TypeOf(v2).String() == "int16" || reflect.TypeOf(v2).String() == "int32" || reflect.TypeOf(v2).String() == "int64" {
						v2 = reflect.ValueOf(v2).Int()
					} else if reflect.TypeOf(v2).String() == "uint8" || reflect.TypeOf(v2).String() == "uint16" || reflect.TypeOf(v2).String() == "uint32" || reflect.TypeOf(v2).String() == "uint64" {
						v2 = reflect.ValueOf(v2).Uint()
						v2 = int64(v2.(uint64))
					}
					v1 = v1.(int64) % v2.(int64)
					v1 = float64(v1.(int64))
				} else {
					fmt.Println("These Type can not mod ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				}
			} else {
				fmt.Println("Type inconsistency ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
			}
		}
	}
	//fmt.Println("v1...", v1)
	return v1
}

func parse_expression() interface{} {
	var v1 interface{}
	var v2 interface{}
	var token Token

	v1 = parse_term()
	for {
		my_get_token(&token)
		if token.Kind != ADD_OPERATOR_TOKEN && token.Kind != SUB_OPERATOR_TOKEN {
			unget_token(&token)
			break
		}
		v2 = parse_term()
		if reflect.TypeOf(v1).String() == reflect.TypeOf(v2).String() {
			if token.Kind == ADD_OPERATOR_TOKEN {
				//fmt.Println("======", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
					v1 = reflect.ValueOf(v1).Float() + reflect.ValueOf(v2).Float()
				} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
					v1 = reflect.ValueOf(v1).Int() + reflect.ValueOf(v2).Int()
				} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
					v1 = reflect.ValueOf(v1).Uint() + reflect.ValueOf(v2).Uint()
				} else if reflect.TypeOf(v1).String() == "rune" {
					v1 = v1.(rune) + v2.(rune)
				} else if reflect.TypeOf(v1).String() == "string" {
					v1 = v1.(string) + v2.(string)
				} else {
					fmt.Println("These Type can not add ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				}
			} else if token.Kind == SUB_OPERATOR_TOKEN {
				if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
					v1 = reflect.ValueOf(v1).Float() - reflect.ValueOf(v2).Float()
				} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
					v1 = reflect.ValueOf(v1).Int() - reflect.ValueOf(v2).Int()
				} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
					v1 = reflect.ValueOf(v1).Uint() - reflect.ValueOf(v2).Uint()
				} else if reflect.TypeOf(v1).String() == "rune" {
					v1 = v1.(rune) - v2.(rune)
				} else {
					fmt.Println("These Type can not sub ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
				}
			} else {
				unget_token(&token)
			}
		} else {
			fmt.Println("Type inconsistency ", reflect.TypeOf(v1).String(), " : ", reflect.TypeOf(v2).String())
		}
	}
	return v1
}

func parse_relation_expression() interface{} {
	var v1 interface{}
	var v2 interface{}
	var token Token

	v1 = parse_expression()
	for {
		my_get_token(&token)
		if token.Kind != EQ_TOKEN && token.Kind != GE_TOKEN && token.Kind != GT_TOKEN && token.Kind != LT_TOKEN && token.Kind != LE_TOKEN && token.Kind != NE_TOKEN {
			unget_token(&token)
			break
		}
		v2 = parse_expression()
		if token.Kind == EQ_TOKEN {
			v1 = (v1 == v2)
			//fmt.Println(v1.(bool))
		} else if token.Kind == GE_TOKEN {
			//fmt.Println(reflect.ValueOf(v1), "GEGEGE...", reflect.ValueOf(v2))
			if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
				v1 = (reflect.ValueOf(v1).Float() >= reflect.ValueOf(v2).Float())
			} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
				v1 = (reflect.ValueOf(v1).Int() >= reflect.ValueOf(v2).Int())
			} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
				v1 = (reflect.ValueOf(v1).Uint() >= reflect.ValueOf(v2).Uint())
			} else if reflect.TypeOf(v1).String() == "rune" {
				v1 = (v1.(rune) >= v2.(rune))
			} else if reflect.TypeOf(v1).String() == "string" {
				v1 = (v1.(string) >= v2.(string))
			} else {
				fmt.Println("These Type can not >= between ", reflect.TypeOf(v1).String(), " and ", reflect.TypeOf(v2).String())
			}
			//fmt.Println(v1.(bool))
		} else if token.Kind == GT_TOKEN {
			//fmt.Println(reflect.ValueOf(v1), "GEGEGE...", reflect.ValueOf(v2))
			if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
				v1 = (reflect.ValueOf(v1).Float() > reflect.ValueOf(v2).Float())
			} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
				v1 = (reflect.ValueOf(v1).Int() > reflect.ValueOf(v2).Int())
			} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
				v1 = (reflect.ValueOf(v1).Uint() > reflect.ValueOf(v2).Uint())
			} else if reflect.TypeOf(v1).String() == "rune" {
				v1 = (v1.(rune) > v2.(rune))
			} else if reflect.TypeOf(v1).String() == "string" {
				v1 = (v1.(string) > v2.(string))
			} else {
				fmt.Println("These Type can not >= between ", reflect.TypeOf(v1).String(), " and ", reflect.TypeOf(v2).String())
			}
			//fmt.Println(v1.(bool))
		} else if token.Kind == LE_TOKEN {
			//fmt.Println(reflect.ValueOf(v1), "GEGEGE...", reflect.ValueOf(v2))
			if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
				v1 = (reflect.ValueOf(v1).Float() <= reflect.ValueOf(v2).Float())
			} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
				v1 = (reflect.ValueOf(v1).Int() <= reflect.ValueOf(v2).Int())
			} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
				v1 = (reflect.ValueOf(v1).Uint() <= reflect.ValueOf(v2).Uint())
			} else if reflect.TypeOf(v1).String() == "rune" {
				v1 = (v1.(rune) <= v2.(rune))
			} else if reflect.TypeOf(v1).String() == "string" {
				v1 = (v1.(string) <= v2.(string))
			} else {
				fmt.Println("These Type can not >= between ", reflect.TypeOf(v1).String(), " and ", reflect.TypeOf(v2).String())
			}
			//fmt.Println(v1.(bool))
		} else if token.Kind == LT_TOKEN {
			//fmt.Println(reflect.ValueOf(v1), "GEGEGE...", reflect.ValueOf(v2))
			if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
				v1 = (reflect.ValueOf(v1).Float() < reflect.ValueOf(v2).Float())
			} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
				v1 = (reflect.ValueOf(v1).Int() < reflect.ValueOf(v2).Int())
			} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
				v1 = (reflect.ValueOf(v1).Uint() < reflect.ValueOf(v2).Uint())
			} else if reflect.TypeOf(v1).String() == "rune" {
				v1 = (v1.(rune) < v2.(rune))
			} else if reflect.TypeOf(v1).String() == "string" {
				v1 = (v1.(string) < v2.(string))
			} else {
				fmt.Println("These Type can not >= between ", reflect.TypeOf(v1).String(), " and ", reflect.TypeOf(v2).String())
			}
			//fmt.Println(v1.(bool))
		} else if token.Kind == NE_TOKEN {
			//fmt.Println(reflect.ValueOf(v1), "GEGEGE...", reflect.ValueOf(v2))
			if reflect.TypeOf(v1).String() == "float32" || reflect.TypeOf(v1).String() == "float64" {
				v1 = (reflect.ValueOf(v1).Float() != reflect.ValueOf(v2).Float())
			} else if reflect.TypeOf(v1).String() == "int8" || reflect.TypeOf(v1).String() == "int16" || reflect.TypeOf(v1).String() == "int32" || reflect.TypeOf(v1).String() == "int64" {
				v1 = (reflect.ValueOf(v1).Int() != reflect.ValueOf(v2).Int())
			} else if reflect.TypeOf(v1).String() == "uint8" || reflect.TypeOf(v1).String() == "uint16" || reflect.TypeOf(v1).String() == "uint32" || reflect.TypeOf(v1).String() == "uint64" {
				v1 = (reflect.ValueOf(v1).Uint() != reflect.ValueOf(v2).Uint())
			} else if reflect.TypeOf(v1).String() == "rune" {
				v1 = (v1.(rune) != v2.(rune))
			} else if reflect.TypeOf(v1).String() == "string" {
				v1 = (v1.(string) != v2.(string))
			} else {
				fmt.Println("These Type can not >= between ", reflect.TypeOf(v1).String(), " and ", reflect.TypeOf(v2).String())
			}
			//fmt.Println(v1.(bool))
		}
	}
	return v1
}

func parse_logic_expression() interface{} {
	var v1 interface{}
	var v2 interface{}
	var token Token
	v1 = parse_relation_expression()
	for {
		my_get_token(&token)
		if token.Kind != LOGICAL_AND_TOKEN && token.Kind != LOGICAL_OR_TOKEN {
			unget_token(&token)
			break
		}
		v2 = parse_relation_expression()
		if token.Kind == LOGICAL_AND_TOKEN {
			//fmt.Println(v1.(bool), "ANDAND", v2.(bool))
			v1 = (v1.(bool) && v2.(bool))
		} else if token.Kind == LOGICAL_OR_TOKEN {
			//fmt.Println(v1.(bool), "OROROR", v2.(bool))
			v1 = (v1.(bool) || v2.(bool))
		}
	}
	return v1
}

//获取变量类型
func getTokenType(str string) TokenType {
	if str == "int8" {
		return INT8
	} else if str == "int16" {
		return INT16
	} else if str == "int32" {
		return INT32
	} else if str == "int64" {
		return INT64
	} else if str == "uint8" {
		return UINT8
	} else if str == "uint16" {
		return UINT16
	} else if str == "uint32" {
		return UINT32
	} else if str == "uint64" {
		return UINT64
	} else if str == "bool" {
		return BOOL
	} else if str == "float32" {
		return FLOAT32
	} else if str == "float64" {
		return FLOAT64
	} else if str == "string" {
		return STRING
	} else if str == "rune" {
		return CHAR
	} else {
		return ERRORTYPE
	}
}

//获取value类型
func getValue(t TokenType, value interface{}, minus_flages int) interface{} {
	if t == INT8 {
		if minus_flages == 1 {
			value = -int8(value.(float64))
		} else {
			value = int8(value.(float64))
		}
	} else if t == INT16 {
		if minus_flages == 1 {
			value = -int16(value.(float64))
		} else {
			value = int16(value.(float64))
		}
	} else if t == INT32 {
		if minus_flages == 1 {
			value = -int32(value.(float64))
		} else {
			value = int32(value.(float64))
		}
	} else if t == INT64 {
		if minus_flages == 1 {
			value = -int64(value.(float64))
		} else {
			value = int64(value.(float64))
		}
	} else if t == FLOAT32 {
		if minus_flages == 1 {
			value = -float32(value.(float64))
		} else {
			value = float32(value.(float64))
		}
	} else if t == FLOAT64 {
		if minus_flages == 1 {
			value = -value.(float64)
		} else {
			value = value.(float64)
		}
	} else if t == UINT8 {
		if minus_flages == 1 {
			value = uint8(-value.(float64))
		} else {
			value = uint8(value.(float64))
		}
	} else if t == UINT16 {
		if minus_flages == 1 {
			value = uint16(-value.(float64))
		} else {
			value = uint16(value.(float64))
		}
	} else if t == UINT32 {
		if minus_flages == 1 {
			value = uint32(-value.(float64))
		} else {
			value = uint32(value.(float64))
		}
	} else if t == UINT64 {
		if minus_flages == 1 {
			value = uint64(-value.(float64))
		} else {
			value = uint64(value.(float64))
		}
	} else if t == CHAR {
		r1 := []rune(value.(string))
		//暂时没有实现转义字符
		value = r1[0]
	} else if t == STRING {
		value = value.(string)
	} else if t == BOOL {
		value = value.(bool)
	}
	return value
}
func getCode() {
	var braces_num int = 1
	for {
		input, _, c := inputReader.ReadLine()
		if c == io.EOF {
			fmt.Println("if error")
			fmt.Println("missing '}' error.")
			os.Exit(1)
		}
		//fmt.Println(len(input), string(input), "get")
		if len(input) == 0 || strings.Replace(string(input), " ", "", -1) == "" { //跳过空行
			continue
		} else {
			line := string(input) + "\n"
			//fmt.Println("get", line)
			if strings.Contains(string(input), "}") {
				braces_num--
				if strings.Contains(string(input), "else") && braces_num == 0 {
					set_line([]rune(strings.TrimSpace(string(input)))[1:])
					//fmt.Println("bak", string([]rune(strings.TrimSpace(string(input)))[1:]))
					stl_bak = 1
					break
				}
			}

			if braces_num == 0 || strings.Replace(string(input), " ", "", -1) == "}" {
				break
			}
			set_line([]rune(line))
			parse_line()
			//fmt.Println(">>", reflect.ValueOf(value))
		}
	}
}

func getCodeInBody(forToken *ForToken) {
	var braces_num int = 1
	for {

		input := forToken.Body[bodyLine]
		bodyLine++
		//fmt.Println(len(input), string(input), "get")
		if len(input) == 0 || strings.Replace(string(input), " ", "", -1) == "" { //跳过空行
			continue
		} else {
			line := string(input)
			//fmt.Println("get", line)
			if strings.Contains(string(input), "}") {
				braces_num--
				if strings.Contains(string(input), "else") && braces_num == 0 {
					set_line([]rune(strings.TrimSpace(string(input)))[1:])
					//fmt.Println("bak", string([]rune(strings.TrimSpace(string(input)))[1:]))
					stl_bak = 1
					break
				}
			}

			if braces_num == 0 || strings.Replace(string(input), " ", "", -1) == "}" {
				break
			}
			set_line([]rune(line))
			parse_line()
			//fmt.Println(">>", reflect.ValueOf(value))
		}
	}
}

//跳过{...}
func skipCode() {
	var braces_num int = 1
	for {
		input, _, c := inputReader.ReadLine()
		if c == io.EOF {
			fmt.Println("if error")
			fmt.Println("missing '}' error.")
			os.Exit(1)
		}
		if len(input) == 0 || strings.Replace(string(input), " ", "", -1) == "" { //跳过空行
			continue
		} else {
			//fmt.Println("sk", line)
			if strings.Contains(string(input), "}") {
				braces_num--
				if strings.Contains(string(input), "else") && braces_num == 0 {
					set_line([]rune(strings.TrimSpace(string(input)))[1:])
					//fmt.Println("bak", string([]rune(strings.TrimSpace(string(input)))[1:]))
					stl_bak = 1
					break
				}
			}
			if strings.Contains(string(input), "{") {
				braces_num++
			}
			if braces_num == 0 {
				break
			}
		}
	}
}

func skipCodeInBody(forToken *ForToken) {
	var braces_num int = 1
	for {
		//fmt.Println(bodyLine, string(forToken.Body[bodyLine]))
		input := forToken.Body[bodyLine]
		bodyLine++
		if len(input) == 0 || strings.Replace(string(input), " ", "", -1) == "" { //跳过空行
			continue
		} else {
			//fmt.Println("sk", line, braces_num)
			if strings.Contains(string(input), "}") {
				braces_num--
				if strings.Contains(string(input), "else") && braces_num == 0 {
					set_line([]rune(strings.TrimSpace(string(input)))[1:])
					fmt.Println("bak", string([]rune(strings.TrimSpace(string(input)))[1:]))
					stl_bak = 1
					break
				}
			}
			if strings.Contains(string(input), "{") {
				braces_num++
			}
			if braces_num == 0 {
				break
			}
		}
	}
}

//获取for条件语句
func getCondition() [3][]rune {
	var conditions [3][]rune
	str_line := string(st_line)
	for_condition_line := string(st_line[strings.Index(str_line, "(")+1 : strings.LastIndex(str_line, ")")])
	condition := strings.Split(for_condition_line, ";")
	for i, v := range condition {
		//condition
		conditions[i] = []rune(v + "\n")
		//fmt.Println("for ", v, i)
	}
	return conditions
}

//获取{...}
func getBodyCode() [][]rune {
	var braces_num int = 1
	var forBody [][]rune
	for {
		input, _, c := inputReader.ReadLine()
		if c == io.EOF {
			fmt.Println("missing '}' error.")
			os.Exit(1)
		}
		if len(input) == 0 || strings.Replace(string(input), " ", "", -1) == "" { //跳过空行
			continue
		} else {
			line := string(input) + "\n"
			//fmt.Println("sk", line)
			if strings.Contains(string(input), "}") {
				braces_num--
			}
			if strings.Contains(string(input), "{") {
				braces_num++
			}
			if braces_num == 0 {
				break
			}
			forBody = append(forBody, []rune(line))
		}
	}
	//fmt.Println(forBody[0][0])
	return forBody
}
func getBodyCodeInBody(forToken *ForToken) [][]rune {
	var braces_num int = 1
	var forBody [][]rune
	lineNum := bodyLine
	for {
		input := forToken.Body[lineNum]
		lineNum++
		if len(input) == 0 || strings.Replace(string(input), " ", "", -1) == "" { //跳过空行
			continue
		} else {
			line := string(input) + "\n"
			//fmt.Println("sk", line)
			if strings.Contains(string(input), "}") {
				braces_num--
			}
			if strings.Contains(string(input), "{") {
				braces_num++
			}
			if braces_num == 0 {
				break
			}
			forBody = append(forBody, []rune(line))
		}
	}
	//fmt.Println(forBody[0][0])
	return forBody
}

func runFor(forToken *ForToken, num int) {
	lvNum = lvNum + 1
	localParam := make(map[string]Token)
	localParamList[lvNum] = localParam
	fmt.Println("condition1", string(forToken.Condition1))
	set_line(forToken.Condition1)
	parse_line()
	for {
		inBody = 1
		set_line(forToken.Condition2)
		value := parse_line()
		fmt.Println("condition2", string(forToken.Condition2))
		if value.(bool) {
			bodyLine = 0
			for i := 0; i < len(forToken.Body); i++ {
				var v []rune
				fmt.Println("forNum :", forNum, i, "bodyLine", bodyLine, "Body :", string(forToken.Body[i]))
				if stl_bak == 0 {
					if t, ok := forTokenList[num+1]; ok {
						delete(forTokenList, num+1)
						i = i + len(t.Body)
					}
					v = forToken.Body[i]
					bodyLine++
					//fmt.Println("for ----", string(v), strings.Replace(string(v), " ", "", -1))
					if strings.TrimSpace(string(v)) != "}" {
						set_line(v)
						parse_line()
					}
				} else {
					stl_bak = 0
					i = bodyLine - 1
					parse_line()
					//fmt.Println(">>", reflect.ValueOf(value))
				}

			}
			set_line(forToken.Condition3)
			value3 := parse_line()
			fmt.Println("condition3", string(forToken.Condition3), value3)
		} else {
			break
		}
	}
	forNum--
	if forNum == 1 {
		inBody = 0
	}
	delete(localParamList, lvNum)
	lvNum = lvNum - 1
}

func excutes() {
	paramList = make(map[string]Token)               //变量列表
	arrayParamList = make(map[string]ArrayListToken) //数组变量列表
	dbParamList = make(map[string]Token)             //数据库变量列表
	forTokenList = make(map[int]ForToken)            //for列表
	localParamList = make(map[int]map[string]Token)  //局部变量列表

	for {
		//fmt.Println("please input:")
		if stl_bak == 0 {
			input, _, c := inputReader.ReadLine()
			if c == io.EOF {
				break
			}
			if len(input) == 0 { //跳过空行
				continue
			} else {
				line := string(input) + "\n"
				//fmt.Println(line)
				set_line([]rune(line))
				parse_line()
				//fmt.Println(">>", reflect.ValueOf(value))
			}
		} else {
			stl_bak = 0
			parse_line()
			//fmt.Println(">>", reflect.ValueOf(value))
		}

	}
	for k, v := range dbParamList {
		fmt.Println(k, " : ", v.Value)
		TokenAsBytes, _ := json.Marshal(v)
		//fmt.Println(string(TokenAsBytes), v)
		db.(store.Store).Set([]byte(string(k)), TokenAsBytes)
	}
}

func parse_line() interface{} {
	var value interface{}

	st_look_ahead_token_exists = 0
	value = parse_logic_expression()
	return value
}

func main() {
	db, _ = store.NewDb("simple.db")
	var err error
	//fi, err = os.Open("D:\\work\\TongJi\\go_work\\src\\src\\github.com\\fate\\mycalc2\\test.fate")
	fi, err = os.Open("D:\\0_chenyao\\git\\src\\github.com\\tjfoc\\tjfoc\\simple\\test.fate")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer fi.Close()
	inputReader = bufio.NewReader(fi)
	excutes()

	ToeknBytes, _ := db.(store.Store).Get([]byte("a"))
	t := Token{}
	json.Unmarshal(ToeknBytes, &t)
	fmt.Println("a : ", t.Value)

	ToeknBytes2, _ := db.(store.Store).Get([]byte("b"))
	t2 := Token{}
	json.Unmarshal(ToeknBytes2, &t2)
	fmt.Println("b : ", t2.Value)
}
