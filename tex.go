package main

import (
	"fmt"
	"main/com"
	"time"
)

// PRE: verdad
// POST: IsPrime devuelve verdad si n es primo y falso en caso contrario
func IsPrime(n int) (foundDivisor bool) {
	foundDivisor = false
	for i := 2; (i < n) && !foundDivisor; i++ {
		foundDivisor = (n%i == 0)
	}
	return !foundDivisor
}

// PRE: interval.A < interval.B
// POST: FindPrimes devuelve todos los números primos comprendidos en el
// 		intervalo [interval.A, interval.B]
func FindPrimes(interval com.TPInterval) float64 {
	var primes []int
	start := time.Now()
	for i := interval.A; i <= interval.B; i++ {
		if IsPrime(i) {
			primes = append(primes, i)
		}
	}
	end := time.Now()
	val := end.Sub(start)
	fmt.Println(val.Seconds())
	return val.Seconds()
}

func texFindPrimes() {
	interval := com.TPInterval{1000, 70000}
	var sum float64
	sum = 0
	for i := 0; i < 25; i++ {
		sum += FindPrimes(interval)
	}
	fmt.Print("Media: ")
	fmt.Println(sum / 25)
}

func trivialFunc() {

}

func texGorutine() {
	var val time.Duration

	for i := 0; i < 5; i++ {
		start := time.Now()
		for i := 0; i < 5; i++ {
			go trivialFunc()
		}
		end := time.Now()
		val += end.Sub(start)
		time.Sleep(3)
	}
	fmt.Print("Media: ")
	fmt.Println(val / 25)
}

func main() {
	texGorutine()
}
