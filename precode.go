package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

func Generator(ctx context.Context, ch chan<- int64, fn func(int64)) {
	defer close(ch)
	i := int64(0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ch <- i
			fn(i)
			i++

		}
	}

}

func Worker(in <-chan int64, out chan<- int64) {
	// 2. Функция Worker
	defer close(out)
	for num := range in {
		out <- num
		time.Sleep(1 * time.Millisecond)
	}

}

func main() {
	chIn := make(chan int64)
	// 3. Создание контекста
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	var inputSum int64
	var inputCount int64
	var mu sync.Mutex

	go Generator(ctx, chIn, func(i int64) {

		mu.Lock()
		defer mu.Unlock()
		inputSum += i
		inputCount++
	})

	const NumOut = 10
	outs := make([]chan int64, NumOut)
	for i := 0; i < NumOut; i++ {
		outs[i] = make(chan int64)
		go Worker(chIn, outs[i])
	}
	amounts := make([]int64, NumOut)
	chOut := make(chan int64, NumOut)

	var wg sync.WaitGroup

	// 4. Собираем числа из каналов outs
	for i, out := range outs {
		wg.Add(1)
		go func(in <-chan int64, i int64) {
			defer wg.Done()
			for num := range in {
				chOut <- num
				amounts[i]++
			}

		}(out, int64(i))
	}

	go func() {
		wg.Wait()
		close(chOut)
	}()

	var count int64
	var sum int64

	// 5. Читаем числа из результирующего канала
	for num := range chOut {
		count++
		sum += num
	}

	fmt.Println("Количество чисел", inputCount, count)
	fmt.Println("Сумма чисел", inputSum, sum)
	fmt.Println("Разбивка по каналам", amounts)

	// проверка результатов
	if inputSum != sum {
		log.Fatalf("Ошибка: суммы чисел не равны: %d != %d\n", inputSum, sum)
	}
	if inputCount != count {
		log.Fatalf("Ошибка: количество чисел не равно: %d != %d\n", inputCount, count)
	}
	for _, v := range amounts {
		inputCount -= v
	}
	if inputCount != 0 {
		log.Fatalf("Ошибка: разделение чисел по каналам неверное\n")
	}
}
