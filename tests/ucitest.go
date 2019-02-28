package main

import (
	"fmt"
	"log"
	"github.com/uci/uci"
)

func main() {
	eng, err := uci.NewEngine("stockfish.exe")
	if err != nil {
		log.Fatal(err)
	}
	
	// set some engine options
	eng.SetOptions(uci.Options{
		Hash:128,
		Threads:1,
		MultiPV:1,
	})

	// set the starting position
	eng.SetFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	
	// set some result filter options
	resultOpts := uci.HighestDepthOnly | uci.IncludeUpperbounds | uci.IncludeLowerbounds
	results, _ := eng.Go(10, "a2a4 c2c3 h2h4", 100000, resultOpts)

	// print it (String() goes to pretty JSON for now)
	fmt.Println(results)
}