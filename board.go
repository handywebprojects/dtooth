package abb

import (
	"fmt"	
	"math"
	"math/rand"
	"time"
	"sort"
	"strings"
	"strconv"

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

func Analyze(fen string, depth int, variantkey string) Position {
	Eng.SetOptions(uci.Options{
		UCI_Variant:variantkey,
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
				score = -INF_SCORE - score
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
	fmt.Println(p)

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

func MakeAlgebmoveStandard(algeb string, fen string) string{
	move, _ := dragontoothmg.ParseMove(algeb)
	board := dragontoothmg.ParseFen(fen)
	board.Apply(move)
	return board.ToFen()
}

type Piece struct{
	Kind string
	Color int
}

type Board struct{
	Rep []Piece
	Turnfen string
	Castlefen string
	Epfen string
}

func (b Board) Tostring() string{
	rows := make([]string, 0)
	for i := 0; i < 8; i++{
		row := make([]string, 0)
		for j := 0; j < 8; j++{
			p := b.Rep[i*8+j]
			row = append(row, fmt.Sprintf("%s%d", p.Kind, p.Color))
		}		
		rows = append(rows, strings.Join(row, " "))
	}
	posrep := strings.Join(rows, "\n-----------------------\n")
	posrep += fmt.Sprintf("\n\n%s %s %s", b.Turnfen, b.Castlefen, b.Epfen)
	return posrep
}

func (b *Board) Setfromfen(fen string){
	b.Rep = make([]Piece, 64)
	fenparts := strings.Split(fen, " ")
	rawfen := fenparts[0]
	b.Turnfen = fenparts[1]
	b.Castlefen = fenparts[2]
	b.Epfen = fenparts[3]
	rawfenrows := strings.Split(rawfen, "/")
	cnt := 0
	for _, row := range(rawfenrows){
		cs := strings.Split(row, "")
		for _, c := range cs{
			if ( c >= "0" ) && ( c <= "9"){
				sc, _ := strconv.Atoi(c)
				for j := 0; j < sc; j++{
					b.Rep[cnt] = Piece{"-", 0}
					cnt++
				}
			}else{
				if ( c >= "A" ) && ( c <= "Z"){
					b.Rep[cnt] = Piece{strings.ToLower(c), 1}
					cnt++
				}else{
					b.Rep[cnt] = Piece{c, 0}
					cnt++
				}
			}
		}		
	}
}

func (b Board) Tofen() string{	
	buff := ""	
	scnt := 0
	for cnt :=0; cnt < 64; {
		p := b.Rep[cnt+scnt]				
		if p.Kind == "-"{
			scnt++
		}else{
			if(scnt>0){
				buff+=fmt.Sprintf("%d", scnt)
				cnt+=scnt
				scnt = 0
			}
			if p.Color == 1{
				buff+=strings.ToUpper(p.Kind)				
			}else{
				buff+=p.Kind
			}
			cnt++			
		}			
		if (scnt > 0) && (((cnt+scnt)%8) == 0){
			buff+=fmt.Sprintf("%d", scnt)
			cnt+=scnt
			scnt = 0
		}
		if(((cnt+scnt)%8)==0)&&(cnt<64){
			buff+="/"
		}
	}	
	buff += " " + b.Turnfen + " " + b.Castlefen + " " + b.Epfen + " 0 1"
	return buff
}

func Sqindeces(sq string) (int, int){	
	return int(sq[0]) - 97, int(56 - sq[1])
}

func index(i int, j int) int{
	return j*8 + i
}

func ijok(i int, j int) bool{
	if (i<0) || (i>7) || (j<0) || (j>7){
		return false
	}
	return true
}

func ijalgeb(i int, j int) string{
	return fmt.Sprintf("%c%c", 97+i, 56-j)
}

func (b *Board) Makealgebmove(algeb string){
	fromi, fromj := Sqindeces(algeb[0:2])
	toi, toj := Sqindeces(algeb[2:4])	
	fromindex := index(fromi, fromj)
	toindex := index(toi, toj)
	fromp := b.Rep[fromindex]	
	top := b.Rep[toindex]	
	if fromp.Kind == "p"{
		if ( fromj - toj ) == 2{
			if ijok(toi-1, toj){
				tp := b.Rep[index(toi-1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 0 ){
					b.Epfen = ijalgeb(toi, toj+1)
				}				
			}
			if ijok(toi+1, toj){
				tp := b.Rep[index(toi+1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 0 ){
					b.Epfen = ijalgeb(toi, toj+1)
				}				
			}
		}
		if ( toj - fromj ) == 2{
			if ijok(toi-1, toj){
				tp := b.Rep[index(toi-1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 1 ){
					b.Epfen = ijalgeb(toi, toj-1)
				}				
			}
			if ijok(toi+1, toj){
				tp := b.Rep[index(toi+1, toj)]
				if ( tp.Kind == "p" ) && ( tp.Color == 1 ){
					b.Epfen = ijalgeb(toi, toj-1)
				}				
			}
		}
	}
	b.Rep[fromindex] = Piece{"-", 0}
	b.Rep[toindex] = fromp
	cK := false
	cQ := false
	ck := false
	cq := false
	for i:=0;i<len(b.Castlefen);i++{
		cp := b.Castlefen[i:i+1]		
		if cp == "K"{
			cK = true
		}
		if cp == "Q"{
			cQ = true
		}
		if cp == "k"{
			ck = true
		}
		if cp == "q"{
			cq = true
		}
	}	
	if b.Turnfen == "w"{
		b.Turnfen = "b"
	}else{
		b.Turnfen = "w"
	}	
	if fromp.Kind == "k"{
		if algeb == "e1g1"{
			b.Rep[63] = Piece{"-", 0}
			b.Rep[61] = Piece{"r", 1}
			cK = false
			cQ = false
		}
		if algeb == "e1c1"{
			b.Rep[56] = Piece{"-", 0}
			b.Rep[59] = Piece{"r", 1}
			cK = false
			cQ = false
		}
		if algeb == "e8g8"{
			b.Rep[7] = Piece{"-", 0}
			b.Rep[5] = Piece{"r", 0}
			ck = false
			cq = false
		}
		if algeb == "e8c8"{
			b.Rep[0] = Piece{"-", 0}
			b.Rep[3] = Piece{"r", 0}
			ck = false
			cq = false
		}
	}		
	if len(algeb) == 5{
		b.Rep[toindex] = Piece{algeb[4:5], fromp.Color}
	}
	if top.Kind != "-"{
		b.Rep[toindex]=Piece{"-", 0}
		for di:=-1;di<2;di++{
			for dj:=-1;dj<2;dj++{
				if!((di==0)&&(dj==0)){
					ni := toi+di
					nj := toj+dj
					if ijok(ni, nj){
						cp := b.Rep[index(ni, nj)]
						if (cp.Kind != "-")&&(cp.Kind != "p"){
							b.Rep[index(ni, nj)] = Piece{"-", 0}
						}
					}
				}
			}
		}
	}
	if b.Rep[63].Kind == "-"{
		cK = false
	}
	if b.Rep[56].Kind == "-"{
		cQ = false
	}
	if b.Rep[7].Kind == "-"{
		ck = false
	}
	if b.Rep[0].Kind == "-"{
		cq = false
	}
	b.Castlefen = ""
	if cK{
		b.Castlefen+="K"
	}
	if cQ{
		b.Castlefen+="Q"
	}
	if ck{
		b.Castlefen+="k"
	}
	if cq{
		b.Castlefen+="q"
	}
	if b.Castlefen==""{
		b.Castlefen="-"
	}
}

func MakeAlgebmoveAtomic(algeb string, fen string) string{
	b := Board{}
	b.Setfromfen(fen)	
	b.Makealgebmove(algeb)	
	newfen := b.Tofen()	
	return newfen
}

func (b Book) MakeAlgebmove(algeb string, fen string) string{
	if b.Variantkey == "atomic"{
		return MakeAlgebmoveAtomic(algeb, fen)
	}
	return MakeAlgebmoveStandard(algeb, fen)
}

func (b Book) SelectRecursive(fen string, depth int64, maxdepth int64, line []string) string{
	fmt.Println("selecting fen", depth, line, fen)
	if depth > maxdepth{
		fmt.Println("max depth exceeded")
		return ""
	}
	if b.Hasfen(fen){
		mli := b.Getmovesbyfen(fen)
		maxmoves := 2
		if depth == 0{
			maxmoves = 10
		}
		if depth == 1{
			maxmoves = 3
		}
		if depth == 2{
			maxmoves = 3
		}
		if len(mli) < 3{
			maxmoves = len(mli)
		}
		sel := rand.Intn(maxmoves)
		selmove := mli[sel]
		newfen := b.MakeAlgebmove(selmove.Algeb, fen)
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
		p := Analyze(fen, int(enginedepth), b.Variantkey)
		fmt.Println("storing", fen, p.Docid)
		b.StorePosition(p)
		return fen
	}	
}

func (b *Book) Minimaxrecursive(fen string, line []string, docids []string, depth int64, maxdepth int64, seldepth int64, nodes int64) (int64, int64, int64){
	//fmt.Println("minimax", fen, line, docids, depth, maxdepth)
	max := int64(-INF_SCORE)
	if depth > maxdepth{		
		return 2 * max, seldepth, nodes
	}
	docid := Fen2docid(fen)
	p, ok := b.Poscache[docid]
	if !ok{		
		return 2 * max, seldepth, nodes
	}
	if depth > seldepth{
		seldepth = depth
	}
	nodes += 1
	for algeb, mi := range p.Moves{		
		newfen := b.MakeAlgebmove(algeb, fen)
		value, newseldepth, newnodes := b.Minimaxrecursive(newfen, append(line, algeb), append(docids, docid), depth + 1, maxdepth, seldepth, nodes)
		seldepth = newseldepth
		nodes = newnodes
		if value < -INF_SCORE{
			value = mi.Score
		}	
		p.Moves[algeb] = MultipvItem{algeb, mi.Score, value, mi.Depth}
		if depth == 0{
			fmt.Println(algeb, mi.Score, value)	
		}		
		if value > max{
			max = value
		}
	}
	return -max, seldepth, nodes
}

func (b *Book) Minimaxout(maxdepth int64){
	start := time.Now()
	fmt.Println("minimaxing out", b.Fullname())
	b.Synccache()
	value, seldepth, nodes := b.Minimaxrecursive(b.Rootfen, []string{}, []string{}, 0, maxdepth, 0, 0)
	fmt.Println("minimax done", -value, seldepth, nodes)
	elapsed := time.Since(start)
	fmt.Println("minimaxing done", b.Fullname(), "took", elapsed, "rate", float32(nodes) / float32(elapsed) * 1e9)
	b.Uploadcache()		
}