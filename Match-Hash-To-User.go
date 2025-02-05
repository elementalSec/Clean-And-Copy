package main

import (
        "bufio"
        "fmt"
        "os"
        "regexp"
        "strings"
        "sync"
)

// Configurable settings
const (
        hashFilePath      = "/home/elemental/Pass-Hashes/hashcat.potfile" // Path to cracked hashes
        outputFilePath    = "Cracked"                          // Output file
        passwordMinLength = 5                                  // Filter: Only save passwords longer than X chars
        numWorkers        = 10                                 // Number of concurrent workers for processing NTDS
)

func main() {
        // Ensure correct usage
        if len(os.Args) < 2 {
                fmt.Println("Usage: ./match-user-to-pass NTDS-DUMP")
                os.Exit(1)
        }

        ntdsFile := os.Args[1]

        // Compile regex for NTLM hash validation
        ntlmRegex := regexp.MustCompile(`(?i)^[0-9a-f]{32}$`)

        // Load hashes from hashcat.potfile into a map
        hashes := make(map[string]string)

        hashFileHandle, err := os.Open(hashFilePath)
        if err != nil {
                fmt.Println("Error opening hashcat.potfile:", err)
                os.Exit(1)
        }
        defer hashFileHandle.Close()

        scanner := bufio.NewScanner(hashFileHandle)
        for scanner.Scan() {
                line := strings.TrimSpace(scanner.Text())
                parts := strings.SplitN(line, ":", 2) // Ensure only two parts: NTLM hash & password

                if len(parts) == 2 && ntlmRegex.MatchString(parts[0]) {
                        password := strings.TrimSpace(parts[1])
                        if len(password) >= passwordMinLength { // Only store passwords that meet length criteria
                                hashes[parts[0]] = password
                        }
                }
        }

        if err := scanner.Err(); err != nil {
                fmt.Println("Error reading hashcat.potfile:", err)
                os.Exit(1)
        }

        // Open NTDS dump
        ntdsHandle, err := os.Open(ntdsFile)
        if err != nil {
                fmt.Println("Error opening NTDS-DUMP file:", err)
                os.Exit(1)
        }
        defer ntdsHandle.Close()

        // Open output file for writing results
        outputHandle, err := os.Create(outputFilePath)
        if err != nil {
                fmt.Println("Error creating output file:", err)
                os.Exit(1)
        }
        defer outputHandle.Close()

        writer := bufio.NewWriter(outputHandle)

        // Concurrent processing with a worker pool
        lines := make(chan string, 100) // Channel to pass NTDS lines
        results := make(chan string, 100) // Channel for matched results
        var wg sync.WaitGroup

        // Worker function
        for i := 0; i < numWorkers; i++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for line := range lines {
                                parts := strings.Split(line, ":")
                                if len(parts) != 7 {
                                        continue
                                }
                                username := parts[0]
                                ntlmHash := parts[3]

                                if password, exists := hashes[ntlmHash]; exists {
                                        results <- fmt.Sprintf("%s:%s\n", username, password)
                                }
                        }
                }()
        }

        // Start a separate goroutine to write results
        go func() {
                for result := range results {
                        writer.WriteString(result)
                }
                writer.Flush()
        }()

        // Read NTDS file line by line and distribute work
        scanner = bufio.NewScanner(ntdsHandle)
        for scanner.Scan() {
                lines <- scanner.Text()
        }
        close(lines) // Signal workers that no more data is coming

        // Wait for workers to finish processing
        wg.Wait()
        close(results) // Signal the result writer to finish

        if err := scanner.Err(); err != nil {
                fmt.Println("Error reading NTDS-DUMP file:", err)
                os.Exit(1)
        }

        fmt.Println("Matching complete! Results saved to:", outputFilePath)
}
