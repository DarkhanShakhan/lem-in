package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	filename := os.Args[1]
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Fatal("Error: invalid input file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	colony := getColony(scanner)
	if len(colony.rooms[colony.end]) == 0 {
		log.Fatal("Error: no tunnel to end room")
	}
	direct := false            //flag to show the direct tunnel from start to end
	startTunnels := []string{} //re-collects all possbile tunnels for start exluding "direct" tunnel
	for _, s := range colony.rooms[colony.start] {
		if s != colony.end {
			startTunnels = append(startTunnels, s)
		} else {
			direct = true //set flag to true for future use
		}
	}
	colony.rooms[colony.start] = startTunnels

	r := allTunnels([]string{}, colony.start, colony.end, colony.rooms)
	tunnels := [][]Tunnel{} //get all possible tunnel combinations
	for i := 0; i < len(r); i++ {
		temp := []Tunnel{}
		if direct { //if direct tunnels exists
			temp = append(temp, Tunnel{path: []string{colony.start, colony.end}, path_len: 2})
		}
		for _, s := range r[i][1:] {
			temp = append(temp, Tunnel{path: s, path_len: len(s)})
		}
		tunnels = append(tunnels, temp)
	}
	min := math.MaxInt
	for _, t := range tunnels { //find optimal tunnel, taking into account the number of ants
		sort.Slice(t, func(i, j int) bool { return t[i].path_len < t[j].path_len })
		distributeAnts(colony.n, t)
		if t[0].path_len+t[0].ants < min {
			colony.tunnel = t
			min = t[0].path_len + t[0].ants
		}
	}
	colony.printMigation()
}

func (c *Colony) printMigation() {
	ants := map[string][]string{}
	n := 1
	c.initMove(ants, &n)
	printMove(ants)
	for len(ants) > 0 {
		c.initMove(ants, &n)
		printMove(ants)
	}
}

func (c *Colony) initMove(ants map[string][]string, n *int) {
	for i := 0; i < len(c.tunnel); i++ {
		if c.tunnel[i].ants > 0 {
			ants[fmt.Sprintf("L%d", n)] = c.tunnel[i].path[1:]
			*n++
			c.tunnel[i].ants--
		}
	}
}

func printMove(ants map[string][]string) {
	for ant, path := range ants {
		if len(path) > 0 {
			fmt.Printf("%s-%s ", ant, path[0])
			ants[ant] = path[1:]
		} else {
			delete(ants, ant)
		}
	}
	if len(ants) != 0 {
		fmt.Println()
	}
}

func allTunnels(curr []string, start, end string, rooms map[string][]string) [][][]string {
	curr_rooms := copyRooms(rooms)
	if len(curr) > 0 {
		for _, r := range curr[1 : len(curr)-1] {
			delete(curr_rooms, r)
		}
	}
	pathes := findAllPaths(start, end, curr_rooms, []string{})
	if len(pathes) == 0 {
		return [][][]string{{curr}}
	}
	res := [][][]string{}
	for _, p := range pathes {
		s := allTunnels(p, start, end, curr_rooms)
		for _, r := range s {
			res = append(res, append([][]string{curr}, r...))
		}
	}
	return res
}

func copyRooms(rooms map[string][]string) map[string][]string {
	res := map[string][]string{}
	for k, v := range rooms {
		res[k] = v
	}
	return res
}

func findAllPaths(curr, end string, rooms map[string][]string, path []string) [][]string {
	if curr == end {
		return [][]string{append(path, curr)}
	}
	var res [][]string
	for _, room := range rooms[curr] {
		if inPath(room, path) {
			continue
		}
		res = append(res, findAllPaths(room, end, rooms, append(path, curr))...)
	}
	return res
}

func distributeAnts(n int, tunnels []Tunnel) {
	for n > 0 {
		for i := len(tunnels) - 1; i >= 0; i-- {
			if i == 0 {
				tunnels[0].ants++
				n--
				break
			}
			if tunnels[i].ants+tunnels[i].path_len < tunnels[i-1].ants+tunnels[i-1].path_len {
				tunnels[i].ants++
				n--
				break
			}
		}
	}
}

type Tunnel struct {
	path     []string
	path_len int
	ants     int
}

type Colony struct {
	n          int
	start, end string
	rooms      map[string][]string
	tunnel     []Tunnel
}

//construct colony from file
func getColony(scanner *bufio.Scanner) Colony {
	col := Colony{rooms: map[string][]string{}}
	var err error
	scanner.Scan()
	col.n, err = strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal("Error: invalid input format") //
	}
	if col.n == 0 {
		log.Fatal("Error: number of ants is zero")
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "##start") {
			scanner.Scan()
			col.start = getRoomName(scanner.Text())
			col.rooms[col.start] = []string{}
			continue
		}
		if strings.HasPrefix(line, "##end") {
			scanner.Scan()
			col.end = getRoomName(scanner.Text())
			col.rooms[col.end] = []string{}
			continue
		}
		if isRoomName(line) {
			col.rooms[getRoomName(line)] = []string{}
			continue
		}
		if isTunnel(col.rooms, line) {
			r := strings.Split(line, "-")
			col.rooms[r[0]] = append(col.rooms[r[0]], r[1])
			col.rooms[r[1]] = append(col.rooms[r[1]], r[0])
			continue
		}
	}
	return col
}

//checks whether the line is tunnel between two rooms
func isTunnel(rooms map[string][]string, line string) bool {
	temp := strings.Split(line, "-")
	if len(temp) != 2 {
		return false
	}
	if _, ok := rooms[temp[0]]; !ok {
		return false
	}
	if _, ok := rooms[temp[1]]; !ok {
		return false
	}
	return true
}

//checks whether the line is a room
func isRoomName(line string) bool {
	temp := strings.Split(line, " ")
	if line[0] == 'L' || line[0] == '#' {
		return false
	}
	if len(temp) != 3 {
		return false
	}
	if _, err := strconv.Atoi(temp[1]); err != nil {
		return false
	}
	if _, err := strconv.Atoi(temp[2]); err != nil {
		return false
	}
	return true
}

//gets the room name from the line
func getRoomName(line string) string {
	return strings.Split(line, " ")[0]
}

func inPath(target string, path []string) bool {
	for _, p := range path {
		if p == target {
			return true
		}
	}
	return false
}
