package main

import (
	"fmt"
	"os"
	"strconv"
	"unicode"
)

var st_line []rune
var st_line_pos int

type LexerStatus int

const (
	INITIAL_STATUS LexerStatus = iota
	IN_INT_PART_STATUS
	DOT_STATUS
	IN_FRAC_PART_STATUS

	CHAR_PART_STATUS
	STRING_PART_STATUS

	FIRST_PARAM_STATUS
	FOLLOW_PARAM_STATUS
)

func getToken(token *Token) {
	var out_pos int = 0
	var status LexerStatus = INITIAL_STATUS
	var current_char rune
	var next_char rune
	token.Str = ""
	token.Kind = BAD_TOKEN
	for {
		// if st_line[st_line_pos] == '\000' {
		// 	break
		// }
		current_char = st_line[st_line_pos]
		//fmt.Println("current_char---(", string(current_char), ")")
		if (status == IN_INT_PART_STATUS || status == IN_FRAC_PART_STATUS) && !unicode.IsDigit(current_char) && current_char != '.' {
			token.Kind = NUMBER_TOKEN

			value, _ := strconv.ParseFloat(token.Str, 64)
			token.Value = float64(value)
			//token.tokenType = FLOAT64

			//fmt.Println("current_char---(", token.Str, ")")
			return
		}

		if (status == FIRST_PARAM_STATUS || status == FOLLOW_PARAM_STATUS) && !unicode.IsDigit(current_char) && current_char != '_' && !unicode.IsLetter(current_char) {
			//fmt.Println("current_str---(", token.Str, ")")
			if IsKeyWord(token.Str) {
				token.Kind = TOKEN_TYPE_TOKEN
			} else if IsStatement(token.Str) {
				token.Kind = STATE_TYPE_TOKEN
			} else if IsBool(token.Str) {
				token.Kind = BOOL_TOKEN
			} else if IsFlowType(token.Str) {
				if token.Str == "if" {
					token.Kind = IF_TOKEN
				} else if token.Str == "else" {
					token.Kind = ELSE_TOKEN
				} else if token.Str == "for" {
					token.Kind = FOR_TOKEN
				}
			} else {
				token.Kind = STATE_TOKEN
			}
			return
		}
		//跳过空格
		if unicode.IsSpace(current_char) {
			//fmt.Println("current_char-------(", string(current_char), ")")
			if current_char == '\n' {
				token.Kind = END_OF_LINE_TOKEN
				return
			}
			if status != CHAR_PART_STATUS && status != STRING_PART_STATUS {
				st_line_pos++
				continue
			}
		}
		if out_pos >= MAX_TOKENSIZE-1 {
			fmt.Println("token too long.")
			os.Exit(1)
		}

		token.Str += string(st_line[st_line_pos])
		st_line_pos++
		out_pos++
		if status != CHAR_PART_STATUS && status != STRING_PART_STATUS {
			if current_char == '+' {
				token.Kind = ADD_OPERATOR_TOKEN
				return
			} else if current_char == '-' {
				token.Kind = SUB_OPERATOR_TOKEN
				return
			} else if current_char == '*' {
				token.Kind = MUL_OPERATOR_TOKEN
				return
			} else if current_char == '/' {
				token.Kind = DIV_OPERATOR_TOKEN
				return
			} else if current_char == '=' {
				next_char = st_line[st_line_pos]
				if next_char == '=' { //判断下一个标识
					token.Kind = EQ_TOKEN
					st_line_pos++
					out_pos++
					token.Str = "=="
					return
				}
				token.Kind = ASS_OPERATOR_TOKEN
				return
			} else if current_char == '>' {
				next_char = st_line[st_line_pos]
				if next_char == '=' { //判断下一个标识
					token.Kind = GE_TOKEN
					st_line_pos++
					out_pos++
					token.Str = ">="
					return
				}
				token.Kind = GT_TOKEN
				return
			} else if current_char == '<' {
				next_char = st_line[st_line_pos]
				if next_char == '=' { //判断下一个标识
					token.Kind = LE_TOKEN
					st_line_pos++
					out_pos++
					token.Str = ">="
					return
				}
				token.Kind = LT_TOKEN
				return
			} else if current_char == '!' {
				next_char = st_line[st_line_pos]
				if next_char == '=' { //判断下一个标识
					token.Kind = NE_TOKEN
					st_line_pos++
					out_pos++
					token.Str = "!="
					return
				}
				token.Kind = BAD_TOKEN
				return
			} else if current_char == '|' {
				next_char = st_line[st_line_pos]
				if next_char == '|' { //判断下一个标识
					token.Kind = LOGICAL_OR_TOKEN
					st_line_pos++
					out_pos++
					token.Str = "||"
					return
				}
				token.Kind = BAD_TOKEN
				return
			} else if current_char == '&' {
				next_char = st_line[st_line_pos]
				if next_char == '&' { //判断下一个标识
					token.Kind = LOGICAL_AND_TOKEN
					st_line_pos++
					out_pos++
					token.Str = "&&"
					return
				}
				token.Kind = BAD_TOKEN
				return
			} else if current_char == '%' {
				token.Kind = MOD_OPERATOR_TOKEN
				return
			} else if unicode.IsDigit(current_char) {
				if status == INITIAL_STATUS {
					status = IN_INT_PART_STATUS
				} else if status == DOT_STATUS {
					status = IN_FRAC_PART_STATUS
				} else if status == FIRST_PARAM_STATUS {
					status = FOLLOW_PARAM_STATUS
				}
			} else if current_char == '.' {
				if status == IN_INT_PART_STATUS {
					status = DOT_STATUS
				} else {
					fmt.Println("syntax error.")
					os.Exit(1)
				}
			} else if current_char == '\'' {
				token.Kind = CHAR_SIGN_TOKEN
				status = CHAR_PART_STATUS
			} else if current_char == '"' {
				token.Kind = STRING_SIGN_TOKEN
				status = STRING_PART_STATUS
			} else if unicode.IsLetter(current_char) {
				if status == INITIAL_STATUS {
					status = FIRST_PARAM_STATUS
				}
			} else if current_char == '_' {
				if status == FIRST_PARAM_STATUS {
					status = FOLLOW_PARAM_STATUS
				} else {
					fmt.Println("error! a variable must begin with an alphabet")
					os.Exit(1)
				}
			} else if current_char == '(' {
				token.Kind = LEFT_PAREN_TOKEN
				return
			} else if current_char == ')' {
				token.Kind = RIGHT_PAREN_TOKEN
				return
			} else if current_char == '{' {
				token.Kind = LEFT_BRACES_TOKEN
				return
			} else if current_char == '}' {
				token.Kind = RIGHT_BRACES_TOKEN
				return
			} else if current_char == '[' {
				token.Kind = LEFT_BRACKET_TOKEN
				return
			} else if current_char == ']' {
				token.Kind = RIGHT_BRACKET_TOKEN
				return
			} else {
				fmt.Println("bad character(", current_char, ")")
				os.Exit(1)
			}
		} else if status == CHAR_PART_STATUS || status == STRING_PART_STATUS {
			if current_char == '\'' {
				token.Kind = CHAR_TOKEN
				return
			} else if current_char == '"' {
				token.Kind = STRING_TOKEN
				return
			}
			//fmt.Println("----current_char---(", token.str, ")")
		}

	}
}

func set_line(line []rune) {
	st_line = line
	st_line_pos = 0
}

func IsKeyWord(str string) bool {
	for _, keyword := range KeyWords {
		if keyword == str {
			return true
		}
	}
	return false
}

func IsStatement(str string) bool {
	for _, statementword := range StatementWords {
		if statementword == str {
			return true
		}
	}
	return false
}

func IsFlowType(str string) bool {
	for _, flowword := range FlowWords {
		if flowword == str {
			return true
		}
	}
	return false
}

func IsBool(str string) bool {
	if str == "true" || str == "false" {
		return true
	} else {
		return false
	}

}

// func parse_line(buf []rune) {
// 	var token Token

// 	set_line(buf)

// 	for {
// 		getToken(&token)
// 		if token.kind == END_OF_LINE_TOKEN {
// 			break
// 		} else {
// 			fmt.Println("kind...", token.kind, "str...", token.str)
// 		}
// 	}
// }
// func main() {
// 	fi, err := os.Open("D:\\0_chenyao\\git\\src\\fate\\mycalc2\\test.fate")
// 	if err != nil {
// 		fmt.Printf("Error: %s\n", err)
// 		return
// 	}
// 	defer fi.Close()
// 	inputReader := bufio.NewReader(fi)
// 	for {
// 		//fmt.Println("please input:")
// 		input, _, c := inputReader.ReadLine()
// 		if c == io.EOF {
// 			break
// 		}
// 		if len(input) == 0 { //跳过空行
// 			continue
// 		} else {
// 			line := string(input) + "\n"
// 			fmt.Println(line)
// 			set_line([]rune(line))
// 			parse_line([]rune(line))
// 			//fmt.Println(">>", reflect.ValueOf(value))
// 		}
// 	}

// }
