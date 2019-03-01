package main

import(
	"fmt"
	"os"
	"time"
	"strconv"

	"github.com/handywebprojects/abb"
)

func boardtest(){
	board := abb.Board{}
	board.Setfromfen(abb.START_FEN)
	board.Makealgebmove("e2e4")	
	fmt.Println(board.Tostring())
}

func main(){
	fmt.Println("Auto Book Builder", abb.Client)
	//ars := abb.Getanalysisroots()
	//ar := ars[0]	
	//abb.Delallpositions("defaultatomic")	
	ar := abb.Analysisroot{abb.START_FEN, 20, 15, "default", "atomic"}
	fmt.Println("analysis root", ar)	
	b := abb.NewBook(ar.Bookname, ar.Bookvariantkey, ar.Fen)			

	numcycles := 1
	numcyclesstr, hasnumcycles := os.LookupEnv("NUMCYCLES")
	if hasnumcycles{
		numcycles, _ = strconv.Atoi(numcyclesstr)
	}

	fmt.Println("build book", b.Fullname(), "cycles", numcycles)
	time.Sleep(10 * time.Second)

	for cycle := 0; cycle < numcycles; cycle++{
		fmt.Println("build cycle", cycle, "of", numcycles)
		time.Sleep(10 * time.Second)
		for i := 0; i < 50; i++ {
			fmt.Println("cycle", i)
			b.Addone(ar.Depth, ar.Enginedepth)
			fmt.Println("position cache size", len(b.Poscache))
		}	
		b.Minimaxout(ar.Depth)
		time.Sleep(5 * time.Minute)
	}
}

