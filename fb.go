package abb

import (
	"fmt"	
	"context"	
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
			ars = append(ars, Analysisroot{fen, depth, enginedepth, bookname, bookvariantkey})
			//fmt.Println(fen, depth, doc.Ref.ID)
			//doc.Ref.Delete(ctx)
	}
	if len(ars) == 0 {
		fmt.Println("no analysis roots found, creating one")
		sr := Analysisroot{START_FEN, 10, 14, "default", "standard"}		
		Addanalysisroot(Fen2docid(START_FEN), sr)				
		ars = append(ars, sr)
	}
	return ars
}

func (b *Book) StorePosition(p Position){	
	b.Positions.Doc(p.Docid).Set(ctx, map[string]interface{}{
		"fen": p.Fen,
		"moves":  p.Serialize(),
	})
	b.Poscache[p.Docid] = p
}