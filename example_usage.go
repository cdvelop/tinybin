package tinybin

import (
	"bytes"
	"fmt"
	"log"
)

// ExampleUsage demonstrates the new TinyBin constructor pattern API
func ExampleUsage() {
	// Basic usage - no logging
	tb := New()

	type User struct {
		ID   int    `binary:"id"`
		Name string `binary:"name"`
		Age  int    `binary:"age"`
	}

	user := User{ID: 1, Name: "Alice", Age: 30}

	// Encode
	data, err := tb.Encode(user)
	if err != nil {
		log.Fatal("Encode failed:", err)
	}

	fmt.Printf("Encoded data length: %d bytes\n", len(data))

	// Decode
	var decoded User
	err = tb.Decode(data, &decoded)
	if err != nil {
		log.Fatal("Decode failed:", err)
	}

	fmt.Printf("Decoded: %+v\n", decoded)
}

// ExampleWithCustomLogging demonstrates custom logging
func ExampleWithCustomLogging() {
	tb := New(func(msg ...any) {
		fmt.Printf("[DEBUG] TinyBin: %v\n", msg)
	})

	type Product struct {
		ID    uint64  `binary:"id"`
		Name  string  `binary:"name"`
		Price float64 `binary:"price"`
	}

	product := Product{ID: 123, Name: "Laptop", Price: 999.99}

	data, err := tb.Encode(product)
	if err != nil {
		log.Fatal("Encode failed:", err)
	}

	var decoded Product
	err = tb.Decode(data, &decoded)
	if err != nil {
		log.Fatal("Decode failed:", err)
	}

	fmt.Printf("Product: %+v\n", decoded)
}

// ExampleMultipleInstances demonstrates instance isolation
func ExampleMultipleInstances() {
	// Create multiple instances - each completely isolated
	httpTB := New()
	grpcTB := New()
	kafkaTB := New()

	type Message struct {
		ID      int    `binary:"id"`
		Content string `binary:"content"`
	}

	msg := Message{ID: 1, Content: "Hello World"}

	// Each instance can be used independently
	httpData, _ := httpTB.Encode(msg)
	grpcData, _ := grpcTB.Encode(msg)
	kafkaData, _ := kafkaTB.Encode(msg)

	// All three byte slices are identical (same encoding logic)
	// but each instance maintains its own cache and pools
	fmt.Printf("HTTP data length: %d\n", len(httpData))
	fmt.Printf("gRPC data length: %d\n", len(grpcData))
	fmt.Printf("Kafka data length: %d\n", len(kafkaData))
}

// ExampleWithWriter demonstrates encoding to specific writers
func ExampleWithWriter() {
	tb := New()

	type Config struct {
		Host string `binary:"host"`
		Port int    `binary:"port"`
	}

	config := Config{Host: "localhost", Port: 8080}

	// Encode to a buffer
	var buffer bytes.Buffer
	err := tb.EncodeTo(config, &buffer)
	if err != nil {
		log.Fatal("EncodeTo failed:", err)
	}

	fmt.Printf("Buffer contains %d bytes\n", buffer.Len())
}

// ExampleConcurrentUsage demonstrates thread-safe concurrent usage
func ExampleConcurrentUsage() {
	tb := New()

	type Counter struct {
		Value int `binary:"value"`
	}

	// Each goroutine can safely use the same instance
	for i := 0; i < 10; i++ {
		go func(id int) {
			counter := Counter{Value: id}

			data, err := tb.Encode(counter)
			if err != nil {
				log.Printf("Goroutine %d: Encode failed: %v", id, err)
				return
			}

			var decoded Counter
			err = tb.Decode(data, &decoded)
			if err != nil {
				log.Printf("Goroutine %d: Decode failed: %v", id, err)
				return
			}

			fmt.Printf("Goroutine %d: %d\n", id, decoded.Value)
		}(i)
	}
}
