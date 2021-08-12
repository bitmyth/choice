// go install url.go
// open url in browser

package main

import (
    "bufio"
    "flag"
    "github.com/atotto/clipboard"
    "github.com/mattn/go-runewidth"
    "github.com/nsf/termbox-go"
    "log"
    "os"
    "os/exec"
    "strings"
    "unicode/utf8"
)

const edit_box_width = 30

type IBox struct {
    text          [][]byte
    filteredText  [][]byte
    scrollY       int
    cursorXOffset int
    cursorYOffset int
    width         int
    height        int
    filter        []rune
}

const coldef = termbox.ColorDefault

var file string

func main() {
    flag.StringVar(&file, "f", "/Users/gsh/url", "file path")
    flag.Parse()

    err := termbox.Init()
    if err != nil {
        panic(err)
    }
    defer termbox.Close()
    termbox.SetInputMode(termbox.InputEsc)

    termbox.Clear(coldef, coldef)

    w, h := termbox.Size()

    box := IBox{width: w, height: h}

    file, err := os.Open(file)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        box.text = append(box.text, scanner.Bytes())
        //fmt.Println(scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

    for {
        box.refresh()
        switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventResize:
            box.width = ev.Width
            box.height = ev.Height
            box.refresh()

        case termbox.EventKey:
            switch ev.Key {
            case termbox.KeyEsc:
                return
            case termbox.KeyCtrlC:
                return

            case termbox.KeyCtrlK:
                box.moveCursorUp()

            case termbox.KeyCtrlJ:
                box.moveCursorDown()

            case termbox.KeyCtrlF:
                box.pageDown()

            case termbox.KeyCtrlB:
                box.pageUp()

            case termbox.KeyEnter:
                box.executeHook()

            case termbox.KeyBackspace2, termbox.KeyBackspace, termbox.KeyDelete, termbox.KeyCtrlD:
                box.deleteFilter()
            default:
                if ev.Ch != 0 {
                    box.appendFilter(ev.Ch)
                }
            }
        }
    }
}

func (box *IBox) deleteFilter() {
    rightBound := len(box.filter) - 1 - 1
    if rightBound > 0 {
        box.filter = box.filter[0:rightBound]
    } else {
        box.filter = box.filter[:0]
    }
}

func (box *IBox) appendFilter(char rune) {
    box.filter = append(box.filter, char)
    box.scrollY = 0
    box.cursorYOffset = 0
}

func (box *IBox) executeHook() {
    selected := string(box.filteredText[box.scrollY+box.cursorYOffset])
    // Copy selected text to clipboard
    clipboard.WriteAll(selected)

    firstAppear := strings.Index(selected, "http")
    if firstAppear == 0 {
        cmd := exec.Command("open", selected)
        _ = cmd.Run()
        return
    } else {
        cmd := exec.Command("/bin/sh", "-c", selected)
        buf, err := cmd.Output()
        if err != nil {
            println(string(buf))
            println(err.Error())
        }
    }
}

func (box *IBox) pageUp() {
    scrollY := box.scrollY - box.height
    if scrollY < 0 {
        scrollY = 0
        box.cursorYOffset = 0
    }
    box.scrollY = scrollY
}

func (box *IBox) pageDown() {
    scrollY := box.scrollY + box.height
    if scrollY >= len(box.filteredText) {
        scrollY = box.scrollY
        box.cursorYOffset = (len(box.filteredText)-scrollY)%box.height - 1
    }
    box.scrollY = scrollY
}

func (box *IBox) moveCursorUp() {
    if box.cursorYOffset > 0 {
        box.cursorYOffset--
    } else if box.cursorYOffset == 0 && box.scrollY > 0 {
        box.scrollY--
    }
}

func (box *IBox) moveCursorDown() {
    // hit the bottom of whole document
    if box.scrollY+box.cursorYOffset == len(box.filteredText)-1 {
        return
    }
    // hit the bottom of screen
    if box.cursorYOffset > box.height-2 {
        box.scrollY++
    } else {
        box.cursorYOffset++
    }
}

func (box *IBox) refresh() {
    y := 0

    // filter text
    if len(box.filter) > 0 {
        box.filteredText = [][]byte{}
        for _, line := range box.text {
            if strings.Contains(string(line), string(box.filter)) {
                box.filteredText = append(box.filteredText, line)
            }
        }
    } else {
        box.filteredText = box.text
        //box.filteredText = make([][]byte, len(box.text))
        //copy(box.filteredText, box.text)
    }

    // loop through all lines
    for index, line := range box.filteredText[box.scrollY:] {
        bg := termbox.ColorDefault
        fg := termbox.ColorDefault

        if box.cursorYOffset == index {
            bg = termbox.ColorBlue
            fg = termbox.ColorWhite
        }

        text := line

        // print line
        x := 0
        for _, c := range string(line) {
            termbox.SetCell(x, y, c, fg, bg)
            x += runewidth.RuneWidth(c)

        }
        // Right padding with blanks
        for x < box.width {
            for remain := box.width - x; remain > 0; remain-- {
                termbox.SetCell(x, y, ' ', fg, bg)
                x++
            }
        }

        for {
            break
            // Deprecated  will causing blanks sit between chinese characters
            r, size := utf8.DecodeRune(text)
            termbox.SetCell(x, y, r, fg, bg)

            x += size
            text = text[size:]
        }

        y++
    }

    if y < box.height {
        for remain := box.height - y; remain > 0; remain-- {
            for x := 0; x < box.width; x++ {
                termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
            }
            y++
        }
    }

    if len(box.filter) > 0 {
        box.print(0, box.height-1, termbox.ColorWhite, termbox.ColorBlack, string(box.filter))
    }
    _ = termbox.Flush()
}

func (box *IBox) print(x, y int, fg, bg termbox.Attribute, msg string) {
    for _, c := range msg {
        termbox.SetCell(x, y, c, fg, bg)
        x += runewidth.RuneWidth(c)
    }
}
