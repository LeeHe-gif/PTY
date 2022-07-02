package pty

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-colorable"
)

type DataProtocol struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}

func (pty *Pty) HandleStdIO() {
	go pty.handleStdIn()
	pty.handleStdOut()
}

func (pty *Pty) handleStdIn() {
	var err error
	var protocol DataProtocol
	var bufferText string
	inputReader := bufio.NewReader(os.Stdin)
	for {
		bufferText, _ = inputReader.ReadString('\n')
		err = json.Unmarshal([]byte(bufferText), &protocol)
		if err != nil {
			fmt.Printf("[MCSMANAGER-TTY] Unmarshall json err:%v\n,original data:%#v\n", err, bufferText)
			continue
		}
		switch protocol.Type {
		case 1:
			pty.StdIn.Write([]byte(protocol.Data))
		case 2:
			resizeWindow(pty, protocol.Data)
		case 3:
			pty.StdIn.Write([]byte{3})
		default:
		}
	}
}

func (pty *Pty) handleStdOut() {
	var err error
	var n int
	buf := make([]byte, 4*2048)
	reader := bufio.NewReader(pty.StdOut)
	stdout := colorable.NewColorableStdout()
	for {
		n, err = reader.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Printf("[MCSMANAGER-TTY] Failed to read from pty master: %v\n", err)
			continue
		} else if err == io.EOF {
			pty.Close()
			os.Exit(-1)
		}
		stdout.Write(buf[:n])
	}
}

// Set the PTY window size based on the text
func resizeWindow(pty *Pty, sizeText string) {
	arr := strings.Split(sizeText, ",")
	cols, err1 := strconv.Atoi(arr[0])
	rows, err2 := strconv.Atoi(arr[1])
	if err1 != nil || err2 != nil {
		fmt.Printf("[MCSMANAGER-TTY] Failed to set window size: %v\n%v\n", err1, err2)
		return
	}
	pty.Setsize(uint32(cols), uint32(rows))
}