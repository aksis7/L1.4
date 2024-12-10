package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Чтение количества воркеров из аргументов командной строки
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <number_of_workers>")
		return
	}
	numWorkers, err := strconv.Atoi(os.Args[1])
	if err != nil || numWorkers <= 0 {
		fmt.Println("Please provide a valid positive integer for the number of workers.")
		return
	}

	// Создаем канал для передачи данных
	dataChannel := make(chan int)

	//Механизм единого контроля завершения (Context)
	// распространяется на все воркеры , которые слушают ctx.Done()

	// Создаем контекст с возможностью отмены
	ctx, cancel := context.WithCancel(context.Background())

	// Запуск горутины для обработки сигнала завершения (Ctrl+C)
	go handleInterrupt(cancel)

	// Запуск воркеров
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, &wg, i, dataChannel)
	}

	// Главный поток, записывающий данные в канал
	go producer(ctx, dataChannel)

	// Ожидаем завершения всех воркеров
	wg.Wait()
	fmt.Println("All workers stopped. Program exiting.")
}

// Горутинa для генерации данных
func producer(ctx context.Context, dataChannel chan int) {
	defer close(dataChannel)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Producer stopping...")
			return
		default:
			data := rand.Intn(100) // Генерируем случайное число
			dataChannel <- data
			time.Sleep(500 * time.Millisecond) // Эмуляция работы
		}
	}
}

// Воркеры, которые читают данные из канала
func worker(ctx context.Context, wg *sync.WaitGroup, id int, dataChannel chan int) {
	defer wg.Done()
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Worker %d stopping...\n", id)
			return
		case data, ok := <-dataChannel:
			if !ok {
				// Канал закрыт
				fmt.Printf("Worker %d detected closed channel. Exiting...\n", id)
				return
			}
			fmt.Printf("Worker %d received data: %d\n", id, data)
		}
	}
}

// Обработка сигнала Ctrl+C
func handleInterrupt(cancel context.CancelFunc) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)

	<-signalChannel // Ждем сигнал
	fmt.Println("Interrupt received. Stopping program...")
	//передаем сигнал завершения всем зависимым горутинам через ctx
	cancel()
}
