package abb

import (
	"fmt"	
	"math"
	"math/rand"
	"time"
	"sort"
	"strings"

	"github.com/dtooth/dragontoothmg"
	"github.com/uci/uci"
)

const INF_SCORE = 10000
const MATE_SCORE = 9000

const Engname = "engines/stockfish9"

var Eng, _ = uci.NewEngine(Engname)

func init(){
	seed := time.Now().Unix()
	fmt.Println("initializing seed", seed)
	rand.Seed(seed)
	fmt.Println("engine", Engname, Eng)
}

func chance(percent int) bool{
	r := rand.Intn(100)
	return r < percent
}

type MultipvItem struct{
	Algeb string
	Score int64
	Eval int64
	Depth int64
}

func (mi MultipvItem) Serialize() map[string]interface{}{
	return map[string]interface{}{
		"algeb": mi.Algeb,
		"score": mi.Score,
		"eval": mi.Eval,
		"depth": mi.Depth,
	}
}

func MultipvItemFromdata(data map[string]interface{}) MultipvItem{
	return MultipvItem{
		data["algeb"].(string),
		data["score"].(int64),
		data["eval"].(int64),
		data["depth"].(int64),
	}
}

type Position struct{
	Fen string
	Zobristkeyhex string
	Moves map[string]MultipvItem
	Docid string
}

type Movelist struct{
	items []MultipvItem
}

func (m *Movelist) Len() int{
	return len(m.items)
}

func (m *Movelist) Swap(i, j int){
	m.items[i], m.items[j] = m.items[j], m.items[i]
}

func (m *Movelist) Less(i, j int) bool{
	return m.items[i].Eval > m.items[j].Eval
}

func (p Position) Getmovelist() Movelist{
	movelist := make([]MultipvItem, 0)
	for _, move := range(p.Moves){
		movelist = append(movelist, move)
	}
	ml := Movelist{movelist}
	sort.Sort(&ml)
	return ml
}

func NewPosition(fen string) Position{
	docid := Fen2docid(fen)
	p := Position{fen, Fen2zobristkeyhex(fen), make(map[string]MultipvItem), docid}
	return p
}

func (p *Position) SetMove(mi MultipvItem){
	p.Moves[mi.Algeb] = mi
}

func (p Position) Serialize() map[string]interface{}{
	m := make(map[string]interface{})
	for algeb, move := range(p.Moves){
		m[algeb] = move.Serialize()
	}
	return m
}

func Analyze(fen string, depth int) Position {
	Eng.SetOptions(uci.Options{
		Hash:128,
		Threads:1,
		MultiPV:250,
	})

	Eng.SetFEN(fen)
	
	resultOpts := uci.HighestDepthOnly
	results, _ := Eng.GoDepth(depth, resultOpts)

	moves := results.Results
	p := NewPosition(fen)
	for _, move := range(moves){		
		score := int64(move.Score)
		depth := int64(move.Depth)
		if move.Mate{
			if score < 0{
				score = -INF_SCORE + score
			}else{
				score = INF_SCORE - score
			}
		}else if math.Abs(float64(score)) > MATE_SCORE{
			if score < 0{
				score = -MATE_SCORE
			}else{
				score = MATE_SCORE
			}
		}
		mi := MultipvItem{move.BestMoves[0], score, score, depth}
		p.SetMove(mi)
	}

	return p
}

func Fen2zobristkey(fen string) uint64{
	board := dragontoothmg.ParseFen(fen)
	return board.Hash()
}

func Fen2zobristkeyhex(fen string) string{
	return fmt.Sprintf("%#016x\n", Fen2zobristkey(fen))[2:]
}

func Fen2docid(fen string) string{	
	parts := strings.Split(fen, " ")
	rawfenparts := strings.Split(parts[0], "/")
	rawfen := strings.Join(rawfenparts, "")
	//docid := rawfen + parts[1] + parts[2] + parts[3]
	// ignore ep, TODO: include ep when there is ep capture
	docid := rawfen + parts[1] + parts[2] + "-"
	return docid
}

func MakeAlgebmove(algeb string, fen string) string{
	move, _ := dragontoothmg.ParseMove(algeb)
	board := dragontoothmg.ParseFen(fen)
	board.Apply(move)
	return board.ToFen()
}

func (b Book) SelectRecursive(fen string, depth int64, maxdepth int64, line []string) string{
	fmt.Println("selecting fen", depth, line, fen)
	if depth > maxdepth{
		fmt.Println("max depth exceeded")
		return ""
	}
	if b.Hasfen(fen){
		mli := b.Getmovesbyfen(fen)
		maxmoves := 3
		if len(mli) < 3{
			maxmoves = len(mli)
		}
		sel := rand.Intn(maxmoves)
		selmove := mli[sel]
		newfen := MakeAlgebmove(selmove.Algeb, fen)
		return b.SelectRecursive(newfen, depth + 1, maxdepth, append(line, selmove.Algeb))
	}else{
		fmt.Println("selected", fen)
		return fen
	}
}

func (b Book) Select(maxdepth int64) string{
	return b.SelectRecursive(b.Rootfen, 0, maxdepth, []string{})
}

func (b Book) Fullname() string{
	return fmt.Sprintf("[Book %s %s]", b.Name, b.Variantkey)
}

func (b Book) Addone(maxdepth int64, enginedepth int64) string{
	fmt.Println("add one to", b.Fullname())
	fen := b.Select(maxdepth)
	if fen == "" {
		fmt.Println("add one failed")
		return ""
	}else{
		fmt.Println("analyzing", fen)
		p := Analyze(fen, int(enginedepth))
		fmt.Println("storing", fen, p.Docid)
		b.StorePosition(p)
		return fen
	}	
}