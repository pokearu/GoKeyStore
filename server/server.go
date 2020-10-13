package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var storePath string = "./store.json"

var mu sync.Mutex

func loadJSON() map[string]string {
	// read JSON file
	data, err := ioutil.ReadFile(storePath)
	if err != nil {
		fmt.Print(err)
	}
	var keyStore map[string]string
	if err := json.Unmarshal(data, &keyStore); err != nil {
		panic(err)
	}
	return keyStore
}

func getJSON() map[string]string {
	mu.Lock()
	keyStore := loadJSON()
	mu.Unlock()
	return keyStore
}

func setJSON(key string, value string) error {
	// Locking for write
	mu.Lock()
	keyStore := loadJSON()
	keyStore[key] = value
	updatedStore, _ := json.Marshal(keyStore)
	err := ioutil.WriteFile(storePath, updatedStore, os.ModePerm)
	if err != nil {
		fmt.Print(err)
		mu.Unlock()
		return err
	}
	mu.Unlock()
	return nil
}

func appendJSON(key string, value string) error {
	// Locking for write
	mu.Lock()
	keyStore := loadJSON()
	if _, ok := keyStore[string(key)]; ok {
		keyStore[key] += " " + value
	} else {
		keyStore[key] = value
	}
	updatedStore, _ := json.Marshal(keyStore)
	err := ioutil.WriteFile(storePath, updatedStore, os.ModePerm)
	if err != nil {
		fmt.Print(err)
		mu.Unlock()
		return err
	}
	mu.Unlock()
	return nil
}

func deleteJSON(key string) error {
	// Locking for write
	mu.Lock()
	keyStore := loadJSON()
	if _, ok := keyStore[string(key)]; ok {
		delete(keyStore, key)
	} else {
		return fmt.Errorf("NOT_FOUND")
	}
	updatedStore, _ := json.Marshal(keyStore)
	err := ioutil.WriteFile(storePath, updatedStore, os.ModePerm)
	if err != nil {
		fmt.Print(err)
		mu.Unlock()
		return err
	}
	mu.Unlock()
	return nil
}

func processMessage(conn net.Conn) {
	log.Println("Started a connection!")
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				log.Println("Ended a connection")
				break
			}
			log.Println(err)
			break
		}
		// Parse command by Space
		command := strings.Fields(message)
		if command[0] == "get" {
			key := command[1]
			keyStore := getJSON()
			if val, ok := keyStore[string(key)]; ok {
				size := len([]byte(strings.TrimSpace(val)))
				res := fmt.Sprintf("VALUE %s %d\r\n%s\r\nEND\r\n", key, size, val)
				conn.Write([]byte(string(res)))
			} else {
				res := fmt.Sprintf("VALUE %s %d\r\n%s\r\nEND\r\n", key, 0, "")
				conn.Write([]byte(string(res)))
			}
		} else if command[0] == "set" {
			//Store value in file
			key := command[1]
			valueSize := strings.TrimSpace(command[2])
			value, _ := reader.ReadString('\n')
			computedValueSize := len([]byte(strings.TrimSpace(value)))
			if size, _ := strconv.Atoi(valueSize); computedValueSize != size {
				conn.Write([]byte(string("CLIENT_ERROR Value size does not match\r\n")))
				continue
			}
			log.Printf("SET %s", key)
			err := setJSON(key, strings.TrimSpace(value))
			if err != nil {
				conn.Write([]byte(string("NOT-STORED\r\n")))
			} else {
				conn.Write([]byte(string("STORED\r\n")))
			}
		} else if command[0] == "append" {
			//Store value in file
			key := command[1]
			valueSize := strings.TrimSpace(command[2])
			value, _ := reader.ReadString('\n')
			computedValueSize := len([]byte(strings.TrimSpace(value)))
			if size, _ := strconv.Atoi(valueSize); computedValueSize != size {
				conn.Write([]byte(string("CLIENT_ERROR Value size does not match\r\n")))
				continue
			}
			log.Printf("APPEND %s", key)
			err := appendJSON(key, strings.TrimSpace(value))
			if err != nil {
				conn.Write([]byte(string("NOT-STORED\r\n")))
			} else {
				conn.Write([]byte(string("STORED\r\n")))
			}
		} else if command[0] == "delete" {
			key := command[1]
			log.Printf("DELETE %s", key)
			err := deleteJSON(key)
			if err != nil {
				if err.Error() == "NOT_FOUND" {
					conn.Write([]byte(string("NOT_FOUND\r\n")))
				} else {
					conn.Write([]byte(string("NOT-DELETED\r\n")))
				}
			} else {
				conn.Write([]byte(string("DELETED\r\n")))
			}
		} else {
			conn.Write([]byte(string("CLIENT_ERROR Command not supported\r\n")))
		}
	}
	conn.Close()
}

func storeExists() bool {
	if _, err := os.Stat(storePath); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		log.Println(err)
		return false
	}
}

func server() {
	log.Println("Start server...")
	// Listen on port 9889
	ln, err := net.Listen("tcp", ":9248")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		// Accept connection
		conn, _ := ln.Accept()
		go processMessage(conn)
	}
}

func main() {
	if !storeExists() {
		err := ioutil.WriteFile(storePath, []byte(`{"unit":"test"}`), 0644)
		if err != nil {
			panic(err)
		}
	}
	server()
}
