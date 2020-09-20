package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"testing"
)

func getWord() string {
	wordList := []string{"a", "about", "all", "also", "and", "as", "at", "be", "because", "but", "by", "can", "come", "could", "day", "do", "even", "find", "first", "for", "from", "get", "give", "go", "have", "he", "her", "here", "him", "his", "how", "I", "if", "in", "into", "it", "its", "just", "know", "like", "look", "make", "man", "many", "me", "more", "my", "new", "no", "not", "now", "of", "on", "one", "only", "or", "other", "our", "out", "people", "say", "see", "she", "so", "some", "take", "tell", "than", "that", "the", "their", "them", "then", "there", "these", "they", "thing", "think", "this", "those", "time", "to", "two", "up", "use", "very", "want", "way", "we", "well", "what", "when", "which", "who", "will", "with", "would", "year", "you", "your"}

	return wordList[rand.Intn(100)]
}

func setCommand(conn net.Conn, key string, value string) {
	setMessage := fmt.Sprintf("set %s %d\r\n%s\r\n", key, len([]byte(value)), value)
	// send to server
	fmt.Fprintf(conn, "%s", setMessage)
}

func getCommand(conn net.Conn, key string) {
	getMessage := fmt.Sprintf("get %s\r\n", key)
	fmt.Fprintf(conn, "%s", getMessage)
}

func clientSetGet(wg *sync.WaitGroup, t *testing.T) {
	// connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:9889")
	if err != nil {
		fmt.Println(err)
		wg.Done()
		return
	}
	key := getWord()
	value := getWord()
	setCommand(conn, key, value)
	getCommand(conn, key)
	reader := bufio.NewReader(conn)
	// Read next 4 messages from server
	for i := 0; i < 4; i++ {
		// wait for reply
		message, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Println(err)
		}
		if i == 2 && strings.TrimSpace(message) == "" {
			t.Errorf("Got %s, Expected %s", message, value)
		}
		// fmt.Print("Message from server: " + message + "\n")
		if strings.TrimSpace(message) == "END" {
			break
		}
	}
	conn.Close()
	wg.Done()
}

func TestSetCommandStored(t *testing.T) {
	// connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:9889")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Set message
	key := getWord()
	value := getWord()
	setCommand(conn, key, value)
	// wait for reply
	message, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println(err)
	}
	message = strings.TrimSpace(message)
	if message != "STORED" {
		t.Errorf("Got %s, Expected %s", message, "STORED")
	}
	conn.Close()
}

func TestGetCommandFound(t *testing.T) {
	// connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:9889")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Get command
	getCommand(conn, "unit")
	// wait for reply
	reader := bufio.NewReader(conn)
	for i := 0; i < 3; i++ {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if i == 1 && message != "test" {
			t.Errorf("Got %s, Expected %s", message, "test")
		} else if i == 2 && message != "END" {
			t.Errorf("Got %s, Expected %s", message, "END")
		}
	}
	conn.Close()
}

func TestGetCommandEmpty(t *testing.T) {
	// connect to server
	conn, err := net.Dial("tcp", "127.0.0.1:9889")
	if err != nil {
		fmt.Println(err)
		return
	}
	// Get command with new key
	getCommand(conn, "NewKey")
	// wait for reply
	reader := bufio.NewReader(conn)
	for i := 0; i < 3; i++ {
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)
		if i == 1 && message != "" {
			t.Errorf("Got %s, Expected %s", message, "")
		} else if i == 2 && message != "END" {
			t.Errorf("Got %s, Expected %s", message, "END")
		}
	}
	conn.Close()
}

func TestGetSetConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go clientSetGet(&wg, t)
	}
	wg.Wait()
}

func main() {
	// var wg sync.WaitGroup
	// for i := 0; i < 1000; i++ {
	// 	wg.Add(1)
	// 	go clientSetGet(&wg, nil)
	// }
	// wg.Wait()
}
