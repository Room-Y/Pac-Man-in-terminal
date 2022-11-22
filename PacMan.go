package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/danicat/simpleansi"
)

var maze []string

func main() {
	// initialize game
	initialise() //启用stty的 cbreak 模式关闭 echo
	defer cleanup()

	// load resources
	err := loadMaze("maze01.txt") //载入地图信息
	if err != nil {
		log.Println("failed to load maze:", err)
		return
	}

	// game loop
	for {
		// update screen
		printScreen()

		// process input
		input, err := readInput()
		if err != nil {
			log.Print("error reading input:", err)
			break
		}

		if input == "ESC" {
			break
		}

		// process movement

		// process collisions

		// check game over

		// Temp: break infinite loop
		// fmt.Println("Hello, Pac Go!")
		break
	}
}

func loadMaze(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		maze = append(maze, line)
	}

	return nil
}

func printScreen() {
	simpleansi.ClearScreen()
	for _, line := range maze {
		fmt.Println(line)
	}
}

func initialise() {
	cbTerm := exec.Command("stty", "cbreak", "-echo")
	cbTerm.Stdin = os.Stdin

	err := cbTerm.Run()
	if err != nil {
		log.Fatalln("unable to activate cbreak mode:", err)
	}
}

func cleanup() {
	cookedTerm := exec.Command("stty", "-cbreak", "echo")
	cookedTerm.Stdin = os.Stdin

	err := cookedTerm.Run()
	if err != nil {
		log.Fatalln("unable to restore cbreak mode:", err)
	}
}

func readInput() (string, error) {
	buffer := make([]byte, 100)

	cnt, err := os.Stdin.Read(buffer)
	if err != nil {
		return "", err
	}

	if cnt == 1 && buffer[0] == 0x1b {
		return "ESC", nil
	}

	return "", nil
}
