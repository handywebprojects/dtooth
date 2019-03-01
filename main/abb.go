package main

import(
	"fmt"
	"time"

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
	ar := abb.Analysisroot{abb.START_FEN, int64(abb.Envint("ANALYSISDEPTH", 20)), int64(abb.Envint("ENGINEDEPTH", 20)), "default", "atomic"}
	fmt.Println("analysis root", ar)	
	fmt.Println("analysis widths", abb.Envint("WIDTH0", abb.DEFAULT_WIDTH0), abb.Envint("WIDTH1", abb.DEFAULT_WIDTH1), abb.Envint("WIDTH2", abb.DEFAULT_WIDTH2))	
	b := abb.NewBook(ar.Bookname, ar.Bookvariantkey, ar.Fen)				

	b.Synccache()

	numcycles := abb.Envint("NUMCYCLES", 1)

	fmt.Println("build book", b.Fullname(), "cycles", numcycles)
	//time.Sleep(10 * time.Second)

	for cycle := 0; cycle < numcycles; cycle++{
		fmt.Println("build cycle", cycle, "of", numcycles)
		//time.Sleep(10 * time.Second)
		for i := 0; i < 5; i++ {
			fmt.Println("cycle", i)
			b.Addone(ar.Depth, ar.Enginedepth)
			fmt.Println("position cache size", len(b.Poscache))
		}	
		b.Minimaxout(ar.Depth)
		time.Sleep(5 * time.Minute)
	}
}

