//Chris de la Iglesia

package main

import (
	"strings"
	"fmt"
	"os"
	"bufio"
	"io/ioutil"
	"sync"
	"sync/atomic"
)

const DATALENGTH = 1000000
var (
	data []int32
	prog []string
	barriers []sync.WaitGroup
	total sync.WaitGroup
	in *bufio.Reader
)


func main() {
	var err error
	if len(os.Args) < 1 {
		fmt.Println("Usage: gokkake <.bk file>")
		return
	}

	bytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	prog = strings.Split(string(bytes),"\n")

	data = make([]int32, DATALENGTH)

	most := 0
	for i := 0; i < len(prog); i++ {
		//This is dumb but golang has no 'max' for integer vals
		if most < len(prog[i]) { 
			most = len(prog[i])
		}
	}
	barriers = make([]sync.WaitGroup, most)

	//fmt.Println(prog)
	
	runtime.GOMAXPROCS(runtime.NumCPU())
	in = bufio.NewReader(os.Stdin)
	spawnThread(0, DATALENGTH/2)
	total.Wait()
}

func run(line int, dataptr int) {
	pc := 0
	var err error
	for pc < len(prog[line]) {
		switch op := prog[line][pc]; op {
		case '+':
			atomic.AddInt32(&data[dataptr], 1)
		case '-':
			atomic.AddInt32(&data[dataptr], -1)
		case '>':
			dataptr++
		case '<':
			dataptr--
		case '[':
			if data[dataptr] == 0 {
				if pc = matchForwards(prog[line], pc); pc == -1 {
					panic(0)
				}
			}
		case ']':
			if data[dataptr] != 0 {
				if pc = matchBackwards(prog[line], pc); pc == -1 {
					panic(0)
				}
			}
		case ',':
			var b byte
			if b, err = in.ReadByte(); err != nil {
				panic(err)
			}
			data[dataptr] = int32(b)
		case '.':
			fmt.Printf("%c",data[dataptr])
		case 'v':
			spawnThread(line+1, dataptr)
		case '^':
			spawnThread(line-1, dataptr)
		case '*':
			spawnThread(line, dataptr)
		case '|':
			barriers[pc].Done()
			barriers[pc].Wait()
			barriers[pc].Add(1)
		default:
		}
		pc++
	}
	total.Done()
}

func spawnThread(line int, dataptr int) {
	if line < 0 {
		line += len(prog)
	}
	if line >= len(prog) {
		line -= len(prog)
	}
	for i := 0; i < len(prog[line]); i++ {
		if prog[line][i] == '|' {
			barriers[i].Add(1)
		}
	}
	total.Add(1)
	go run(line, dataptr)
}

func matchBackwards(src string, i int) int {
	depth := 0
	r := i
	for r > 0 {
		end := strings.LastIndex(src[:r], "]")
		start := strings.LastIndex(src[:r],"[")
		if start > end {
			if depth == 0 {
				return start
			} else {
				r = start
				depth--
			}
		} else {
			r = end
			depth++
		}
	}
	return -1
}

func matchForwards(src string, i int) int {
	depth := 0
	r := i
	for r < len(src) {
		end := strings.Index(src[r+1:], "]")
		start := strings.Index(src[r+1:],"[")
		if end == -1 {
			return -1
		}
		if end < start {
			if depth == 0 {
				return end
			} else {
				r = end
				depth--
			}
		} else {
			r = start
			depth++
		}
	}
	return -1
}