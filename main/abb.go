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

	ar := abb.Analysisroot{
		abb.Envstr("ANALYSISROOT", abb.START_FEN),
		int64(abb.Envint("ANALYSISDEPTH", abb.DEFAULT_ANALYSISDEPTH)),
		int64(abb.Envint("ENGINEDEPTH", abb.DEFAULT_ENGINEDEPTH)),
		abb.Envstr("BOOKNAME", abb.DEFAULT_BOOKNAME),
		abb.Envstr("VARIANTKEY", abb.DEFAULT_VARIANTKEY),
		abb.Envint("NUMCYCLES", abb.DEFAULT_NUMCYCLES),
		abb.Envint("BATCHSIZE", abb.DEFAULT_BATCHSIZE),
		int64(abb.Envint("CUTOFF", abb.DEFAULT_CUTOFF)),
		abb.Envint("WIDTH0", abb.DEFAULT_WIDTH0),
		abb.Envint("WIDTH1", abb.DEFAULT_WIDTH1),
		abb.Envint("WIDTH2", abb.DEFAULT_WIDTH2),
	}
	
	fmt.Println("analysis root", ar)		

	b := abb.NewBook(ar.Bookname, ar.Bookvariantkey, ar.Fen)				

	b.Synccache()

	time.Sleep(10 * time.Second)
	for cycle := 0; cycle < ar.Numcycles; cycle++{
		fmt.Println("build cycle", cycle + 1, "of", ar.Numcycles)
		time.Sleep(10 * time.Second)
		for i := 0; i < ar.Batchsize; i++ {
			fmt.Println("build cycle", cycle + 1, "batch", i)
			time.Sleep(10 * time.Second)
			b.Addone(ar)
			fmt.Println("position cache size", len(b.Poscache))
		}	
		b.Minimaxout(ar)
		time.Sleep(5 * time.Minute)
	}
}
