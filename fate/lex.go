package fate

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/tjfoc/tjfoc/core/miscellaneous"
	"github.com/tjfoc/tjfoc/core/store"
)

const (
	ROLLSIZE = 2
	BUFFSIZE = 1024
)

var ft FateLex // 全局变量

type FateLex struct {
	pos        int
	lim        int
	roll       int    // 回滚字符数
	rollBuffer []int  // 回滚字符
	buffer     []byte // 缓冲区
	fd         io.Reader
	lock       sync.Mutex
	db         store.Store
	records    map[string]string
}

func (f *FateLex) ungetc() {
	f.roll++
}

// 获取下一个字符
func (f *FateLex) nextc() int {
	var c int

	if f.roll > 0 {
		f.roll--
		return f.rollBuffer[f.roll]
	}
	if f.pos == f.lim {
		if n, _ := f.fd.Read(f.buffer); n == 0 {
			c = 0 // EOF
		} else {
			f.lim = n
			f.pos = 1
			c = int(f.buffer[0])
		}
	} else {
		c = int(f.buffer[f.pos])
		f.pos++
	}
	f.rollBuffer[1] = f.rollBuffer[0]
	f.rollBuffer[0] = c
	return c
}

// 跳过空白符和注释
func (f *FateLex) skipSpace() {
	var c int

	for {
		if c = f.nextc(); !isSpace(c) {
			f.ungetc()
			break
		}
	}
	if c == '#' {
		for {
			if c = f.nextc(); c == '\n' {
				f.ungetc()
				break
			} else if c == 0 {
				f.ungetc()
				return
			}
		}
	}
}

func (f *FateLex) Lex(lval *FateSymType) int {
	f.skipSpace()
	buf := []byte{}
	for {
		c := f.nextc()
		if isAlnum(c) {
			buf = append(buf, byte(c))
		} else if c == '"' {
			for {
				c := f.nextc()
				switch c {
				case 0, ';':
					return 0
				case '"':
					lval.sval = string(miscellaneous.Dup(buf))
					typ := wordType(buf)
					buf = []byte{}
					return typ
				default:
					buf = append(buf, byte(c))
				}
			}
		} else if len(buf) > 0 {
			f.ungetc()
			lval.sval = string(miscellaneous.Dup(buf))
			typ := wordType(buf)
			buf = []byte{}
			return typ
		} else {
			switch c {
			case 0, ';':
				return c
			}
		}
	}
}

func (f *FateLex) Error(s string) {
	log.Fatal(fmt.Errorf("Lex: %s", s))
}

func (f *FateLex) Run() map[string]string {
	defer ft.lock.Unlock()
	ft.records = make(map[string]string)
	FateParse(&ft)
	records := ft.records
	for k, v := range ft.records {
		ft.db.Set([]byte(k), []byte(v))
	}
	ft.records = make(map[string]string)
	return records
}

func (f *FateLex) Set(k, v string) {
	f.records[k] = v
}

func (f *FateLex) Get(k string) {
	if v, err := ft.db.Get([]byte(k)); err != nil {
		f.records[k] = string(v)
	}
}

func New(fd io.Reader, db store.Store) *FateLex {
	ft.lock.Lock()
	ft.fd = fd
	ft.db = db
	ft.buffer = make([]byte, BUFFSIZE)
	ft.rollBuffer = make([]int, ROLLSIZE)
	return &ft
}
