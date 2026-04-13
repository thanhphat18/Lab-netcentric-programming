package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Student struct {
	ID         int
	StudyHours int
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== Go Concurrency Lab Menu ===")
		fmt.Println("1. Three Counters")
		fmt.Println("2. Website fetcher")
		fmt.Println("3. Calculator with Channels")
		fmt.Println("4. Message Queue")
		fmt.Println("5. File Downloader")
		fmt.Println("6. Number Processor")
		fmt.Println("7. Search Race")
		fmt.Println("8. Library Simulation")
		fmt.Println("0. Exit")

		choice := readInt(reader, "Choose option: ")

		switch choice {
		case 1:
			task11ThreeCounters()
		case 2:
			task12WebsiteFetcher()
		case 3:
			task21CalculatorWithChannels()
		case 4:
			task22MessageQueue()
		case 5:
			task31FileDownloader()
		case 6:
			task32NumberProcessor()
		case 7:
			task41SearchRace()
		case 8:
			task51LibrarySimulation()
		case 0:
			fmt.Println("Goodbye!")
			return
		default:
			fmt.Println("Invalid option!")
		}
	}
}

// part 1
// task 1.1
func task11ThreeCounters() {
	fmt.Println("\n=== Part 1 - Task 1.1: Three Counters ===")

	go counter("A", 3)
	go counter("B", 4)
	go counter("C", 5)

	time.Sleep(2 * time.Second)
	fmt.Println("All done!")
}

func counter(name string, max int) {
	for i := 1; i <= max; i++ {
		fmt.Printf("Counter %s: %d\n", name, i)
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Printf("Counter %s finished!\n", name)
}

func readLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func readInt(reader *bufio.Reader, prompt string) int {
	for {
		input := readLine(reader, prompt)
		value, err := strconv.Atoi(input)
		if err == nil {
			return value
		}
		fmt.Println("Please enter a valid integer.")
	}
}

// task 1.2
func task12WebsiteFetcher() {
	fmt.Println("\n=== Part 1 - Task 1.2: Website Fetcher ===")
	start := time.Now()

	go fetchWebsite("Google.com", 200)
	go fetchWebsite("Facebook.com", 400)
	go fetchWebsite("Amazon.com", 300)
	go fetchWebsite("Twitter.com", 150)

	time.Sleep(500 * time.Millisecond)
	fmt.Printf("\nCompleted in: %s\n", time.Since(start))
}

func fetchWebsite(name string, delayMs int) {
	fmt.Printf("Fetching %s...\n", name)
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	fmt.Printf("✓ Got data from %s\n", name)
}

// part 2
// task 1.1
func task21CalculatorWithChannels() {
	fmt.Println("\n=== Part 2 - Task 2.1: Calculator with Channels ===")
	numbers := []int{10, 20, 30, 40, 50}
	fmt.Println("Numbers:", numbers)

	sumChan := make(chan int)
	avgChan := make(chan float64)

	start := time.Now()

	go calculateSum(numbers, sumChan)
	go calculateAverage(numbers, avgChan)

	sum := <-sumChan
	average := <-avgChan

	fmt.Printf("Sum: %d\n", sum)
	fmt.Printf("Average: %.1f\n", average)
	fmt.Printf("Time: %s\n", time.Since(start))
}

func calculateSum(numbers []int, result chan int) {
	sum := 0
	for _, num := range numbers {
		sum += num
		time.Sleep(100 * time.Millisecond)
	}
	result <- sum
}

func calculateAverage(numbers []int, result chan float64) {
	sum := 0
	for _, num := range numbers {
		sum += num
		time.Sleep(100 * time.Millisecond)
	}

	average := 0.0
	if len(numbers) > 0 {
		average = float64(sum) / float64(len(numbers))
	}

	result <- average
}

// task 1.2
func task22MessageQueue() {
	fmt.Println("\n=== Part 2 - Task 2.2: Message Queue ===")

	messages := make(chan string, 10)

	go sender("Alice", messages, 3)
	go sender("Bob", messages, 2)
	go sender("Charlie", messages, 4)

	for i := 0; i < 9; i++ {
		msg := <-messages
		fmt.Println(msg)
	}

	fmt.Println("\nAll messages received!")
}

func sender(name string, messages chan string, count int) {
	for i := 1; i <= count; i++ {
		msg := fmt.Sprintf("Message %d from %s", i, name)
		messages <- msg
		time.Sleep(150 * time.Millisecond)
	}
}

// part 3
// task 3.1
func task31FileDownloader() {
	fmt.Println("\n=== Part 3 - Task 3.1: File Downloader ===")
	start := time.Now()

	files := map[string]int{
		"video.mp4":   8,
		"song.mp3":    4,
		"photo.jpg":   2,
		"doc.pdf":     5,
		"archive.zip": 6,
	}

	var wg sync.WaitGroup

	for filename, size := range files {
		wg.Add(1)
		go downloadFile(filename, size, &wg)
	}

	wg.Wait()
	fmt.Printf("\n✓ All downloads complete! (%s)\n", time.Since(start))
}

func downloadFile(filename string, sizeMB int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Downloading %s (%dMB)...\n", filename, sizeMB)
	time.Sleep(time.Duration(sizeMB*100) * time.Millisecond)
	fmt.Printf("✓ %s complete!\n", filename)
}

// task 3.2
func task32NumberProcessor() {
	fmt.Println("\n=== Part 3 - Task 3.2: Number Processor ===")
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	fmt.Println("Numbers:", numbers)

	evenChan := make(chan []int)
	oddChan := make(chan []int)
	squareChan := make(chan []int)

	var wg sync.WaitGroup
	start := time.Now()

	wg.Add(3)
	go findEvens(numbers, evenChan, &wg)
	go findOdds(numbers, oddChan, &wg)
	go findSquares(numbers, squareChan, &wg)

	go func() {
		wg.Wait()
		close(evenChan)
		close(oddChan)
		close(squareChan)
	}()

	evens := <-evenChan
	odds := <-oddChan
	squares := <-squareChan

	fmt.Println("Evens:", evens)
	fmt.Println("Odds:", odds)
	fmt.Println("Squares:", squares)
	fmt.Printf("\nTime: %s\n", time.Since(start))
}

func findEvens(numbers []int, result chan []int, wg *sync.WaitGroup) {
	defer wg.Done()

	var evens []int
	for _, num := range numbers {
		if num%2 == 0 {
			evens = append(evens, num)
		}
		time.Sleep(50 * time.Millisecond)
	}
	result <- evens
}

func findOdds(numbers []int, result chan []int, wg *sync.WaitGroup) {
	defer wg.Done()

	var odds []int
	for _, num := range numbers {
		if num%2 != 0 {
			odds = append(odds, num)
		}
		time.Sleep(50 * time.Millisecond)
	}
	result <- odds
}

func findSquares(numbers []int, result chan []int, wg *sync.WaitGroup) {
	defer wg.Done()

	var squares []int
	for _, num := range numbers {
		squares = append(squares, num*num)
		time.Sleep(50 * time.Millisecond)
	}
	result <- squares
}

// part 4
// 4.1
func task41SearchRace() {
	fmt.Println("\n=== Part 4 - Task 4.1: Search Race ===")
	query := "golang concurrency"

	chA := make(chan string)
	chB := make(chan string)

	start := time.Now()

	go searchEngineA(query, chA)
	go searchEngineB(query, chB)

	select {
	case result := <-chA:
		fmt.Printf("Engine A won! (%s)\n", time.Since(start))
		fmt.Println(result)
	case result := <-chB:
		fmt.Printf("Engine B won! (%s)\n", time.Since(start))
		fmt.Println(result)
	}
}

func searchEngineA(query string, ch chan string) {
	time.Sleep(300 * time.Millisecond)
	result := fmt.Sprintf("Results from Engine A for '%s'", query)
	ch <- result
}

func searchEngineB(query string, ch chan string) {
	time.Sleep(200 * time.Millisecond)
	result := fmt.Sprintf("Results from Engine B for '%s'", query)
	ch <- result
}

// Part 5:
// task 5.1
func task51LibrarySimulation() {
	fmt.Println("\n=== Part 5 - Task 5.1: Library Simulation ===")
	fmt.Println("Library capacity: 30 students")
	fmt.Println("Total students today: 100")
	fmt.Println("Simulation: 1 second = 1 hour\n")

	library := make(chan bool, 30)
	var wg sync.WaitGroup
	start := time.Now()

	students := make([]Student, 100)
	for i := 0; i < 100; i++ {
		students[i] = Student{
			ID:         i + 1,
			StudyHours: rand.Intn(4) + 1,
		}
	}

	for _, s := range students {
		wg.Add(1)
		go student(s.ID, s.StudyHours, library, &wg)
	}

	wg.Wait()
	totalOpenTime := time.Since(start)

	fmt.Println("\n=== Simulation Complete ===")
	fmt.Printf("Total students served: %d\n", len(students))
	fmt.Printf("Library was open for: %.0f hours\n", totalOpenTime.Seconds())
}

func student(id int, studyHours int, library chan bool, wg *sync.WaitGroup) {
	defer wg.Done()

	if len(library) == cap(library) {
		fmt.Printf("Student %d waiting. (library full)\n", id)
	}

	library <- true
	fmt.Printf("Student %d entered library, will study for %d hours\n", id, studyHours)

	time.Sleep(time.Duration(studyHours) * time.Second)

	<-library
	fmt.Printf("Student %d left library after %d hours\n", id, studyHours)
}
