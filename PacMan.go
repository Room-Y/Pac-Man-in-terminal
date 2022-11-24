package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/danicat/simpleansi"
)

type sprite struct {
	row int
	col int
}

type Config struct {
	Player   string `json:"player"`
	Ghost    string `json:"ghost"`
	Wall     string `json:"wall"`
	Dot      string `json:"dot"`
	Pill     string `json:"pill"`
	Death    string `json:"death"`
	Space    string `json:"space"`
	UseEmoji bool   `json:"use_emoji"`
}

var cfg Config

var maze []string

var player sprite
var ghosts []*sprite

var score int
var numDots int
var lives = 1
var err_or_end_tag string

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

	// load json config
	err = loadConfig("config.json")
	if err != nil {
		log.Println("failed to load configuration:", err)
		return
	}

	// process input (async)
	input := make(chan string)
	go func(ch chan<- string) {
		for {
			inn, err := readInput()
			if err != nil {
				log.Println("error reading input:", err)
				err_or_end_tag = "err"
			}
			ch <- inn
		}
	}(input)

	// game loop
	for {
		// process movement
		select {
		case inp := <-input:
			if inp == "ESC" {
				err_or_end_tag = "ESC"
			}
			movePlayer(inp)
		default:
		}

		moveGhosts()

		// process collisions
		for _, g := range ghosts {
			if player == *g {
				lives--
			}
		}

		// update screen
		printScreen()

		// check game over
		if err_or_end_tag != "" {
			log.Println("err_or_end_tag:", err_or_end_tag)
			break
		}
		if numDots == 0 {
			log.Println("numDots:", numDots)
			log.Println("You Win!")
			break
		}
		if lives <= 0 {
			moveCursor(player.row, player.col)
			fmt.Print(cfg.Death)
			moveCursor(len(maze)+2, 0)
			log.Println("lives:", lives)
			log.Println("Your lives over!")
			break
		}

		// repeat
		time.Sleep(300 * time.Millisecond)

		// Temp: break infinite loop
		// fmt.Println("Hello, Pac Go!")
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

	for row, line := range maze {
		for col, char := range line {
			switch char {
			case 'P':
				player = sprite{row, col}
			case 'G':
				ghosts = append(ghosts, &sprite{row, col})
			case '.':
				numDots++
			}
		}
	}

	return nil
}

func printScreen() {
	simpleansi.ClearScreen()
	for _, line := range maze {
		for _, chr := range line {
			switch chr {
			case '#':
				fmt.Print(simpleansi.WithBlueBackground(cfg.Wall))
			case '.':
				fmt.Print(cfg.Dot)
			case 'X':
				fmt.Print(cfg.Pill)
			default:
				fmt.Print(cfg.Space)
			}
		}
		fmt.Println()
	}

	// print player
	moveCursor(player.row, player.col)
	fmt.Print(cfg.Player)

	// print ghosts
	for _, g := range ghosts {
		moveCursor(g.row, g.col)
		fmt.Print(cfg.Ghost)
	}

	//move cursor outside of maze drawing area
	moveCursor(len(maze)+1, 0)
	fmt.Println("Score:", score, "\tLives:", lives)
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
	} else if cnt >= 3 {
		if buffer[0] == 0x1b && buffer[1] == '[' {
			switch buffer[2] {
			case 'A':
				return "UP", nil
			case 'B':
				return "DOWN", nil
			case 'C':
				return "RIGHT", nil
			case 'D':
				return "LEFT", nil
			}
		}
	}

	return "", nil
}

func makeMove(oldRow, oldCol int, dir string) (newRow, newCol int) {
	newRow, newCol = oldRow, oldCol

	switch dir {
	case "UP":
		newRow = newRow - 1
		if newRow < 0 {
			newRow = len(maze) - 1
		}
	case "DOWN":
		newRow = newRow + 1
		if newRow == len(maze) {
			newRow = 0
		}
	case "RIGHT":
		newCol = newCol + 1
		if newCol == len(maze[0]) {
			newCol = 0
		}
	case "LEFT":
		newCol = newCol - 1
		if newCol < 0 {
			newCol = len(maze[0]) - 1
		}
	}

	if maze[newRow][newCol] == '#' {
		newRow, newCol = oldRow, oldCol
	}

	return
}

func movePlayer(dir string) {
	player.row, player.col = makeMove(player.row, player.col, dir)

	removeDot := func(row, col int) {
		maze[player.row] = maze[player.row][0:player.col] + " " + maze[player.row][player.col+1:]
	}

	switch maze[player.row][player.col] {
	case '.':
		numDots--
		score++
		removeDot(player.row, player.col)
	case 'X':
		score += 10
		removeDot(player.row, player.col)
	}
}

func drawDirection() string {
	dir := rand.Intn(4)
	move := map[int]string{
		0: "UP",
		1: "DOWN",
		2: "RIGHT",
		3: "LEFT",
	}
	return move[dir]
}

func moveGhosts() {
	for _, g := range ghosts {
		dir := drawDirection()
		g.row, g.col = makeMove(g.row, g.col, dir)
	}
}

func loadConfig(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := json.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	return nil
}

func moveCursor(row, col int) {
	if cfg.UseEmoji {
		simpleansi.MoveCursor(row, col*2)
	} else {
		simpleansi.MoveCursor(row, col)
	}
}
