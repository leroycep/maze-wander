package main

import (
	"encoding/csv"
	"github.com/nsf/termbox-go"
	"io"
	"os"
	"time"
)

func checkCollision(tiles [][]rune, x, y int) bool {
	if x < 0 || y < 0 || y > len(tiles) || x > len(tiles[0]) {
		return true
	}
	switch tiles[y][x] {
	case '.':
		return false
	default:
		return true
	}
}

func dialogBox(words string) {
	for i := 1; i < 80; i++ {
		termbox.SetCell(i, 30, '-', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(i, 36, '-', termbox.ColorWhite, termbox.ColorDefault)
	}
	for i := 31; i < 36; i++ {
		termbox.SetCell(0, i, '|', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(80, i, '|', termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.SetCell(0, 30, '+', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(80, 30, '+', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(0, 36, '+', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(80, 36, '+', termbox.ColorWhite, termbox.ColorDefault)

	for idx, char := range words {
		termbox.SetCell(1+idx, 31, char, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.SetCursor(1+len(words), 31)
}

func typeDialog(dialogOutput chan string, text string) {
	for idx, _ := range text {
		dialogOutput <- text[:idx+1]
		time.Sleep(50 * time.Millisecond)
	}
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	var playerx, playery int = 20, 15
	var tiles [][]rune

	// Stuff for the dialog box
	var dialogDisplay string = ""

	eventQueue := make(chan termbox.Event) // So that we can have async keyboard
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	dialogQueue := make(chan string)
	go typeDialog(dialogQueue, "Somebody wanted to have a bad time...")

	var tilesFile io.Reader
	tilesFile, err = os.Open("assets/map.csv")
	if err != nil {
		panic(err)
	}

	csvReader := csv.NewReader(tilesFile)
	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}

	tiles = make([][]rune, len(records))
	for y, row := range records {
		tileRow := make([]rune, len(row))
		for x, _ := range row {
			switch row[x] {
			case "-1":
				tileRow[x] = ' '
			case "178":
				tileRow[x] = '+'
			case "210":
				tileRow[x] = '-'
			case "59":
				tileRow[x] = '|'
			case "4":
				playerx = x
				playery = y
				fallthrough
			case "226":
				tileRow[x] = '.'
			}
		}
		tiles[y] = tileRow
	}

loop:
	for {
		select { // Multiplex between channells
		case ev := <-eventQueue: // Handle termbox events
			if ev.Type == termbox.EventKey {
				// Handle keyboard input
				nextx, nexty := playerx, playery
				switch {
				case ev.Key == termbox.KeyEsc:
					break loop
				case ev.Key == termbox.KeyArrowLeft, ev.Ch == 'a':
					nextx -= 1
				case ev.Key == termbox.KeyArrowRight, ev.Ch == 'd':
					nextx += 1
				case ev.Key == termbox.KeyArrowUp, ev.Ch == 'w':
					nexty -= 1
				case ev.Key == termbox.KeyArrowDown, ev.Ch == 's':
					nexty += 1
				}
				if !checkCollision(tiles, nextx, nexty) {
					playerx, playery = nextx, nexty
				}
			}
		case text := <-dialogQueue:
			dialogDisplay = text
		default: // Do main loop
			termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
			for y, row := range tiles {
				for x, tile := range row {
					termbox.SetCell(x, y, tile, termbox.ColorGreen, termbox.ColorDefault)
				}
			}
			termbox.SetCell(playerx, playery, '@', termbox.ColorYellow, termbox.ColorDefault)
			dialogBox(dialogDisplay)
			termbox.Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}
}
