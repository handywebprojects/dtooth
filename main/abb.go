package main

import(
	"fmt"
	"time"

	"github.com/handywebprojects/abb"
)

func main(){
	fmt.Println("Auto Book Builder", abb.Client)
	ars := abb.Getanalysisroots()
	ar := ars[0]	
	fmt.Println("analysis root", ar)
	//abb.Delallpositions("defaultstandard")
	b := abb.NewBook(ar.Bookname, ar.Bookvariantkey, ar.Fen)
	start := time.Now()
	fmt.Println("getting all positions from the book")
	ps := b.Getallpositions()
	numpos := len(ps)
	elapsed := time.Since(start)
	fmt.Println("done, number of positions", numpos, "took", elapsed, "rate", float32(numpos) / float32(elapsed) * 1e9, "pos/sec")
	for i := 0; i < 10; i++ {
		fmt.Println("cycle", i)
		b.Addone(ar.Depth, ar.Enginedepth)
		fmt.Println("position cache size", len(b.Poscache))
	}	
}

