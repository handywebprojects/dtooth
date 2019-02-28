package main

import (
	"fmt"	

	"github.com/dtooth/dragontoothmg"
)

func main(){
	// Read a position from a FEN string
	board := dragontoothmg.ParseFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	// Generate all legal moves
	//moveList := board.GenerateLegalMoves()
	moveList := []dragontoothmg.Move{}
	move1, _ := dragontoothmg.ParseMove("e2e4")
	move2, _ := dragontoothmg.ParseMove("d2d4")
	moveList = append(moveList, move1)
	moveList = append(moveList, move2)
	// For every legal move
	for _, currMove := range moveList {
		// Apply it to the board
		unapplyFunc := board.Apply(currMove)
		// Print the move, the new position, and the hash of the new position
		fmt.Println("Moved to:", &currMove) // Reference converts Move to string automatically
		fmt.Println("New position is:", board.ToFen())
		zk := board.Hash()
		zkh := fmt.Sprintf("%#016x\n", zk)[2:]
		fmt.Println(zkh)
		//fmt.Println("This new position has Zobrist hash:", zk)
		// Unapply the move
		unapplyFunc()
	}
}