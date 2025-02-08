package main

import (
        "bufio"
        "encoding/hex"
        "fmt"
        "os"
        "regexp"
        "strings"
        "sync"
)

// Configurable settings
const (
        passwordMinLength = 1  // Filter: Only save passwords longer than X chars
        numWorkers        = 10 // Number of concurrent workers for processing NTDS
)

func decodeHex(hexStr string) (string, error) {
        decoded, err := hex.DecodeString(hexStr)
        if err != nil {
                return "", err
        }
        return string(decoded), nil
}

func main() {
        // Ensure correct usage
        if len(os.Args) < 4 {
                fmt.Println("Usage: ./match-hash-to-user NTDS-DUMP HASHCAT-POTFILE OUTPUT-FILE")
                os.Exit(1)
        }

        ntdsFile := os.Args[1]
        hashFilePath := os.Args[2]
        outputFilePath := os.Args[3]

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
                parts := strings.SplitN(line, ":", 2)
                if len(parts) == 2 && ntlmRegex.MatchString(parts[0]) {
                        password := strings.TrimSpace(parts[1])
                        if strings.HasPrefix(strings.ToUpper(password), "$HEX[") {
                                hexValue := strings.TrimSuffix(strings.TrimPrefix(password, "$HEX["), "]")
                                decoded, err := decodeHex(hexValue)
                                if err == nil {
                                        password = decoded
                                }
                        }
                        if len(password) >= passwordMinLength {
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
        lines := make(chan string, 100)
        results := make(chan string, 100)
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
        close(lines)

        // Wait for workers to finish processing
        wg.Wait()
        close(results)

        if err := scanner.Err(); err != nil {
                fmt.Println("Error reading NTDS-DUMP file:", err)
                os.Exit(1)
        }

        fmt.Println("Matching complete! Results saved to:", outputFilePath)
}

        fmt.Println("Matching complete! Results saved to:", outputFilePath)
}
