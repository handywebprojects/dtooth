package main

// https://blog.kowalczyk.info/article/wOYk/advanced-command-execution-in-go-with-osexec.html

import (
	"fmt"
	"io"
	"bufio"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

func readStdout(r io.Reader)() {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
        fmt.Println("line:*** ", scanner.Text(), " ***")
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    } else {
		fmt.Println("stdout closed ok")
	}
}

func readStderr(r io.Reader)() {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
        fmt.Println("error:*** ", scanner.Text(), " ***")
    }

    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    } else {
		fmt.Println("stderr closed ok")
	}
}

func main() {
	cmd := exec.Command("ls", "-l")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("stockfish.exe")
	}

	stdinOut, _ := cmd.StdinPipe()
	fmt.Print(stdinOut)
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {		
		readStdout(stdoutIn)
		wg.Done()
	}()

	go func() {		
		readStderr(stderrIn)
		wg.Done()
	}()

	go func(){
		time.Sleep(1e9)
		io.WriteString(stdinOut, "uci\n")
		time.Sleep(1e9)
		io.WriteString(stdinOut, "quit\n")
	}()

	wg.Wait()

	err = cmd.Wait()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}		
}
