package abb

import (
	"fmt"	
	"context"	
	"time"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"cloud.google.com/go/firestore"
)

const START_FEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"

var ctx = context.Background()

var opt = option.WithCredentialsFile("firebase/fbsacckey.json")
var app, apperr = firebase.NewApp(context.Background(), nil, opt)
var Client, clienterr = app.Firestore(ctx)
var analysisrootscollection = Client.Collection("analysisroots")
var bookscollection = Client.Collection("books")

type Book struct{
	Name string
	Variantkey string
	Id string
	Rootfen string
	Positions *firestore.CollectionRef
	Poscache map[string]Position
}

func (b Book) Store(){
	bookscollection.Doc(b.Id).Set(ctx, map[string]interface{}{
		"fen": b.Name,
		"variantkey": b.Variantkey,
		"id": b.Id,
		"rootfen": b.Rootfen,
		"positions": b.Positions,
	})
}

func PositionFromdocsnapshot(ds *firestore.DocumentSnapshot) Position{
	data := ds.Data()
	fen := data["fen"].(string)
	zobristkeyhex := Fen2zobristkeyhex(fen)
	docid := Fen2docid(fen)
	movesdata := data["moves"].(map[string]interface{})
	moves := make(map[string]MultipvItem)
	for _, movedata := range(movesdata){
		moves[movedata.(map[string]interface{})["algeb"].(string)] = MultipvItemFromdata(movedata.(map[string]interface{}))
	}
	return Position{fen, zobristkeyhex, moves, docid}
}

func (b *Book) Hasfen(fen string) bool{	
	docid := Fen2docid(fen)
	_, ok := b.Poscache[docid]
	if ok{
		return true
	}
	snapshot, _ := b.Positions.Doc(docid).Get(ctx)
	if snapshot.Exists(){
		b.Poscache[docid] = PositionFromdocsnapshot(snapshot)
		return true
	}else{
		return false
	}
}

func (b Book) Getcachedpositionbyfen(fen string) Position{
	docid := Fen2docid(fen)
	return b.Poscache[docid]
}

func (b Book) Getmovelistbyfen(fen string) Movelist{
	p := b.Getcachedpositionbyfen(fen)
	ml := p.Getmovelist()
	return ml
}

func (b Book) Getmovesbyfen(fen string) []MultipvItem{	
	return b.Getmovelistbyfen(fen).items
}

func NewBook(name string, variantkey string, rootfen string) Book{	
	id := name + variantkey
	docref := bookscollection.Doc(id)
	positionsref := docref.Collection("positions")
	b := Book{
		name,
		variantkey,
		id,
		rootfen,
		positionsref,
		make(map[string]Position),
	}	
	b.Store()
	return b
}

type Analysisroot struct{
	Fen string
	Depth int64	
	Enginedepth int64
	Bookname string
	Bookvariantkey string	
	Numcycles int
	Batchsize int
	Cutoff int64
	Width0 int
	Width1 int
	Width2 int
}

func Addanalysisroot(id string, ar Analysisroot){
	analysisrootscollection.Doc(id).Set(ctx, map[string]interface{}{
		"fen": ar.Fen,
		"depth":  ar.Depth,
		"enginedepth":  ar.Enginedepth,
		"bookname": ar.Bookname,
		"bookvariantkey": ar.Bookvariantkey,
	})
}

func Delallpositions(bookname string){
	iter := bookscollection.Doc(bookname).Collection("positions").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		fmt.Println("deleting", doc.Ref.ID)
		doc.Ref.Delete(ctx)
	}
}

func (b Book) Getallpositions() []Position{
	positions := make([]Position, 0)
	iter := bookscollection.Doc(b.Id).Collection("positions").Documents(ctx)
	for {
		doc, err := iter.Next()				
		if err == iterator.Done {
				break
		}
		if err != nil {
				log.Fatalf("Failed to iterate: %v", err)
		}
		p := PositionFromdocsnapshot(doc)
		positions = append(positions, p)
	}
	return positions
}

func (b *Book) Synccache(){
	start := time.Now()
	fmt.Println("syncing cache", b.Fullname())
	ps := b.Getallpositions()
	numpos := len(ps)
	b.Poscache = make(map[string]Position)
	for _, p := range ps{
		b.Poscache[p.Docid] = p
	}
	elapsed := time.Since(start)
	fmt.Println("syncing cache done", b.Fullname(), "number of positions", numpos, "took", elapsed, "rate", float32(numpos) / float32(elapsed) * 1e9, "pos/sec")
}

func (b *Book) Uploadcache(){
	start := time.Now()
	fmt.Println("uploading cache", b.Fullname())
	for _, p := range b.Poscache{
		b.StorePosition(p)
	}
	numpos := len(b.Poscache)
	elapsed := time.Since(start)
	fmt.Println("uploading cache done", b.Fullname(), "number of positions", numpos, "took", elapsed, "rate", float32(numpos) / float32(elapsed) * 1e9, "pos/sec")
}

func Getanalysisroots() []Analysisroot{
	iter := analysisrootscollection.Documents(ctx)
	ars := make([]Analysisroot, 0)
	for {
			doc, err := iter.Next()
			if err == iterator.Done {
					break
			}
			if err != nil {
					log.Fatalf("Failed to iterate: %v", err)
			}
			data := doc.Data()			
			fen := data["fen"].(string)
			depth := data["depth"].(int64)						
			enginedepth := data["enginedepth"].(int64)	
			bookname := data["bookname"].(string)
			bookvariantkey := data["bookvariantkey"].(string)	
			numcycles := data["numcycles"].(int)	
			batchsize := data["batchsize"].(int)	
			cutoff := data["cutoff"].(int64)	
			width0 := data["width0"].(int)	
			width1 := data["width1"].(int)	
			width2 := data["width2"].(int)	
			ars = append(ars, Analysisroot{fen, depth, enginedepth, bookname, bookvariantkey, numcycles, batchsize, cutoff, width0, width1, width2})
			//fmt.Println(fen, depth, doc.Ref.ID)
			//doc.Ref.Delete(ctx)
	}
	if len(ars) == 0 {
		fmt.Println("no analysis roots found, creating one")
		sr := Analysisroot{START_FEN, DEFAULT_ANALYSISDEPTH, DEFAULT_ENGINEDEPTH, DEFAULT_BOOKNAME, DEFAULT_VARIANTKEY, DEFAULT_NUMCYCLES, DEFAULT_BATCHSIZE, DEFAULT_CUTOFF, DEFAULT_WIDTH0, DEFAULT_WIDTH1, DEFAULT_WIDTH2}		
		Addanalysisroot(Fen2docid(START_FEN), sr)				
		ars = append(ars, sr)
	}
	return ars
}

func (b *Book) StorePosition(p Position){	
	b.Positions.Doc(p.Docid).Set(ctx, map[string]interface{}{
		"fen": p.Fen,
		"moves": p.Serialize(),
	})
	b.Poscache[p.Docid] = p
}