/*
Copyright Suzhou Tongji Fintech Research Institute 2018 All Rights Reserved.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
//line yacc.y:2
package fate

import __yyfmt__ "fmt"

//line yacc.y:2
//line yacc.y:5
type FateSymType struct {
	yys  int
	val  int
	sval string
}

const EOF = 57346
const SET = 57347
const GET = 57348
const DATA = 57349

var FateToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"EOF",
	"SET",
	"GET",
	"DATA",
	"';'",
}
var FateStatenames = [...]string{}

const FateEofCode = 1
const FateErrCode = 2
const FateInitialStackSize = 16

//line yacc.y:25
//line yacctab:1
var FateExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const FatePrivate = 57344

const FateLast = 12

var FateAct = [...]int{

	6, 10, 4, 5, 3, 8, 7, 2, 1, 0,
	0, 9,
}
var FatePact = [...]int{

	-3, -1000, -8, -1000, -1, -2, -3, -6, -1000, -1000,
	-1000,
}
var FatePgo = [...]int{

	0, 8, 7, 4,
}
var FateR1 = [...]int{

	0, 1, 2, 2, 3, 3, 3,
}
var FateR2 = [...]int{

	0, 1, 1, 3, 0, 3, 2,
}
var FateChk = [...]int{

	-1000, -1, -2, -3, 5, 6, 8, 7, 7, -3,
	7,
}
var FateDef = [...]int{

	4, -2, 1, 2, 0, 0, 4, 0, 6, 3,
	5,
}
var FateTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 8,
}
var FateTok2 = [...]int{

	2, 3, 4, 5, 6, 7,
}
var FateTok3 = [...]int{
	0,
}

var FateErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	FateDebug        = 0
	FateErrorVerbose = false
)

type FateLexer interface {
	Lex(lval *FateSymType) int
	Error(s string)
}

type FateParser interface {
	Parse(FateLexer) int
	Lookahead() int
}

type FateParserImpl struct {
	lval  FateSymType
	stack [FateInitialStackSize]FateSymType
	char  int
}

func (p *FateParserImpl) Lookahead() int {
	return p.char
}

func FateNewParser() FateParser {
	return &FateParserImpl{}
}

const FateFlag = -1000

func FateTokname(c int) string {
	if c >= 1 && c-1 < len(FateToknames) {
		if FateToknames[c-1] != "" {
			return FateToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func FateStatname(s int) string {
	if s >= 0 && s < len(FateStatenames) {
		if FateStatenames[s] != "" {
			return FateStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func FateErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !FateErrorVerbose {
		return "syntax error"
	}

	for _, e := range FateErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + FateTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := FatePact[state]
	for tok := TOKSTART; tok-1 < len(FateToknames); tok++ {
		if n := base + tok; n >= 0 && n < FateLast && FateChk[FateAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if FateDef[state] == -2 {
		i := 0
		for FateExca[i] != -1 || FateExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; FateExca[i] >= 0; i += 2 {
			tok := FateExca[i]
			if tok < TOKSTART || FateExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if FateExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += FateTokname(tok)
	}
	return res
}

func Fatelex1(lex FateLexer, lval *FateSymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = FateTok1[0]
		goto out
	}
	if char < len(FateTok1) {
		token = FateTok1[char]
		goto out
	}
	if char >= FatePrivate {
		if char < FatePrivate+len(FateTok2) {
			token = FateTok2[char-FatePrivate]
			goto out
		}
	}
	for i := 0; i < len(FateTok3); i += 2 {
		token = FateTok3[i+0]
		if token == char {
			token = FateTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = FateTok2[1] /* unknown char */
	}
	if FateDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", FateTokname(token), uint(char))
	}
	return char, token
}

func FateParse(Fatelex FateLexer) int {
	return FateNewParser().Parse(Fatelex)
}

func (Fatercvr *FateParserImpl) Parse(Fatelex FateLexer) int {
	var Faten int
	var FateVAL FateSymType
	var FateDollar []FateSymType
	_ = FateDollar // silence set and not used
	FateS := Fatercvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	Fatestate := 0
	Fatercvr.char = -1
	Fatetoken := -1 // Fatercvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		Fatestate = -1
		Fatercvr.char = -1
		Fatetoken = -1
	}()
	Fatep := -1
	goto Fatestack

ret0:
	return 0

ret1:
	return 1

Fatestack:
	/* put a state and value onto the stack */
	if FateDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", FateTokname(Fatetoken), FateStatname(Fatestate))
	}

	Fatep++
	if Fatep >= len(FateS) {
		nyys := make([]FateSymType, len(FateS)*2)
		copy(nyys, FateS)
		FateS = nyys
	}
	FateS[Fatep] = FateVAL
	FateS[Fatep].yys = Fatestate

Fatenewstate:
	Faten = FatePact[Fatestate]
	if Faten <= FateFlag {
		goto Fatedefault /* simple state */
	}
	if Fatercvr.char < 0 {
		Fatercvr.char, Fatetoken = Fatelex1(Fatelex, &Fatercvr.lval)
	}
	Faten += Fatetoken
	if Faten < 0 || Faten >= FateLast {
		goto Fatedefault
	}
	Faten = FateAct[Faten]
	if FateChk[Faten] == Fatetoken { /* valid shift */
		Fatercvr.char = -1
		Fatetoken = -1
		FateVAL = Fatercvr.lval
		Fatestate = Faten
		if Errflag > 0 {
			Errflag--
		}
		goto Fatestack
	}

Fatedefault:
	/* default state action */
	Faten = FateDef[Fatestate]
	if Faten == -2 {
		if Fatercvr.char < 0 {
			Fatercvr.char, Fatetoken = Fatelex1(Fatelex, &Fatercvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if FateExca[xi+0] == -1 && FateExca[xi+1] == Fatestate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			Faten = FateExca[xi+0]
			if Faten < 0 || Faten == Fatetoken {
				break
			}
		}
		Faten = FateExca[xi+1]
		if Faten < 0 {
			goto ret0
		}
	}
	if Faten == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			Fatelex.Error(FateErrorMessage(Fatestate, Fatetoken))
			Nerrs++
			if FateDebug >= 1 {
				__yyfmt__.Printf("%s", FateStatname(Fatestate))
				__yyfmt__.Printf(" saw %s\n", FateTokname(Fatetoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for Fatep >= 0 {
				Faten = FatePact[FateS[Fatep].yys] + FateErrCode
				if Faten >= 0 && Faten < FateLast {
					Fatestate = FateAct[Faten] /* simulate a shift of "error" */
					if FateChk[Fatestate] == FateErrCode {
						goto Fatestack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if FateDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", FateS[Fatep].yys)
				}
				Fatep--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if FateDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", FateTokname(Fatetoken))
			}
			if Fatetoken == FateEofCode {
				goto ret1
			}
			Fatercvr.char = -1
			Fatetoken = -1
			goto Fatenewstate /* try again in the same state */
		}
	}

	/* reduction by production Faten */
	if FateDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", Faten, FateStatname(Fatestate))
	}

	Fatent := Faten
	Fatept := Fatep
	_ = Fatept // guard against "declared and not used"

	Fatep -= FateR2[Faten]
	// Fatep is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if Fatep+1 >= len(FateS) {
		nyys := make([]FateSymType, len(FateS)*2)
		copy(nyys, FateS)
		FateS = nyys
	}
	FateVAL = FateS[Fatep+1]

	/* consult goto table to find next state */
	Faten = FateR1[Faten]
	Fateg := FatePgo[Faten]
	Fatej := Fateg + FateS[Fatep].yys + 1

	if Fatej >= FateLast {
		Fatestate = FateAct[Fateg]
	} else {
		Fatestate = FateAct[Fatej]
		if FateChk[Fatestate] != -Faten {
			Fatestate = FateAct[Fateg]
		}
	}
	// dummy call; replaced with literal code
	switch Fatent {

	case 5:
		FateDollar = FateS[Fatept-3 : Fatept+1]
		//line yacc.y:21
		{
			ft.Set(FateDollar[2].sval, FateDollar[3].sval)
		}
	case 6:
		FateDollar = FateS[Fatept-2 : Fatept+1]
		//line yacc.y:22
		{
			ft.Get(FateDollar[2].sval)
		}
	}
	goto Fatestack /* stack new state and value */
}
