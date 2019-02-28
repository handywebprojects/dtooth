package main

import(
	"fmt"

	"github.com/handywebprojects/abb"
)

func main(){
	fmt.Println("Auto Book Builder", abb.Client)
	ars := abb.Getanalysisroots()
	ar := ars[0]	
	fmt.Println("analysis root", ar)
	//abb.Delallpositions("defaultstandard")
	b := abb.NewBook(ar.Bookname, ar.Bookvariantkey, ar.Fen)	
	for i := 0; i < 10; i++ {
		fmt.Println("cycle", i)
		b.Addone(ar.Depth, ar.Enginedepth)
		fmt.Println("position cache size", len(b.Poscache))
	}	
	b.Minimaxout(ar.Depth)
}

