package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
)

func getWord() string {
	wordList := []string{"a", "about", "all", "also", "and", "as", "at", "be", "because", "but", "by", "can", "come", "could", "day", "do", "even", "find", "first", "for", "from", "get", "give", "go", "have", "he", "her", "here", "him", "his", "how", "I", "if", "in", "into", "it", "its", "just", "know", "like", "look", "make", "man", "many", "me", "more", "my", "new", "no", "not", "now", "of", "on", "one", "only", "or", "other", "our", "out", "people", "say", "see", "she", "so", "some", "take", "tell", "than", "that", "the", "their", "them", "then", "there", "these", "they", "thing", "think", "this", "those", "time", "to", "two", "up", "use", "very", "want", "way", "we", "well", "what", "when", "which", "who", "will", "with", "would", "year", "you", "your"}

	return wordList[rand.Intn(100)]
}

func client(wg *sync.WaitGroup) {

	// connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:9889")
	if err != nil {
		fmt.Println(err)
		wg.Done()
		return
	}

	key := getWord()
	value := getWord()
	setMessage := "set " + key + " 5\r\n" + value + "\r\n"
	getMessage := "get " + key + "\r\n"
	// send to server
	fmt.Fprintf(conn, "%s", setMessage)
	fmt.Fprintf(conn, "%s", getMessage)
	reader := bufio.NewReader(conn)
	for {
		// wait for reply
		message, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
		}
		fmt.Print("Message from server: " + message + "\n")
		if strings.TrimSpace(message) == "END" {
			break
		}
	}
	conn.Close()
	wg.Done()
}

func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go client(&wg)
	}
	wg.Wait()
}
