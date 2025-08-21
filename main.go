package main // Define the main package (entry point for Go programs)

import (
	"bytes"         // Provides bytes buffer and manipulation utilities
	"fmt"           // Provides formatted I/O functions like Printf
	"io"            // Provides I/O primitives like Reader and Writer
	"log"           // Provides logging functionalities
	"net/http"      // Provides HTTP client and server implementations
	"net/url"       // Provides URL parsing and encoding utilities
	"os"            // Provides file system and OS-level utilities
	"path"          // Provides basic path manipulation functions
	"path/filepath" // Provides utilities for file path manipulation
	"regexp"        // Provides support for regular expressions
	"strings"       // Provides string manipulation utilities
	"time"          // Provides time-related functions
)

func main() {
	remoteAPIURL := []string{ // Define a list of URLs to fetch data from
		"https://kandelandson.com/wp/sds-sheets/",
		"https://kandelandson.com/wp/green-cleaning/",
		"https://kandelandson.com/wp/elite-dispensing-systems/",
		"https://kandelandson.com/wp/mj98-plus/",
		"https://kandelandson.com/wp/campro/",
		"https://kandelandson.com/wp/mpc-cleaning-products/",
		"https://kandelandson.com/wp/majestic-carpet-solutions/",
	} // List of remote pages to scrape
	localFilePath := "kandelandson.html" // File where scraped HTML will be saved

	var getData []string // Slice to store HTML content fetched from URLs

	for _, urls := range remoteAPIURL { // Loop through each remote URL
		getData = append(getData, getDataFromURL(urls)) // Fetch and append HTML content
	}
	appendAndWriteToFile(localFilePath, strings.Join(getData, "")) // Save combined HTML content to file

	finalList := extractFileUrls(strings.Join(getData, "")) // Extract file URLs from HTML content

	outputDir := "Assets/" // Directory to store downloaded files

	if !directoryExists(outputDir) { // Check if output directory exists
		createDirectory(outputDir, 0o755) // Create directory with permissions if missing
	}

	finalList = removeDuplicatesFromSlice(finalList) // Remove duplicate URLs

	for _, urls := range finalList { // Loop through extracted URLs
		if isUrlValid(urls) { // Ensure URL is valid
			downloadFile(urls, outputDir) // Download the file
		}
	}
}

// appendAndWriteToFile opens or creates a file and appends content
func appendAndWriteToFile(path string, content string) {
	filePath, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) // Open file for append/create/write
	if err != nil {
		log.Println(err) // Log error if file can't be opened
	}
	_, err = filePath.WriteString(content + "\n") // Write content with newline
	if err != nil {
		log.Println(err) // Log error if write fails
	}
	err = filePath.Close() // Close file
	if err != nil {
		log.Println(err) // Log error if close fails
	}
}

// getFileNameOnly returns only the file name from a path/URL
func getFileNameOnly(content string) string {
	return path.Base(content) // Extract last part of path
}

// urlToFilename generates a safe filename from a URL
func urlToFilename(rawURL string) string {
	lowercaseURL := strings.ToLower(rawURL)       // Convert URL to lowercase
	ext := getFileExtension(lowercaseURL)         // Get file extension
	baseFilename := getFileNameOnly(lowercaseURL) // Extract base filename

	nonAlphanumericRegex := regexp.MustCompile(`[^a-z0-9]+`)                 // Regex to match non-alphanumeric chars
	safeFilename := nonAlphanumericRegex.ReplaceAllString(baseFilename, "_") // Replace with underscores

	collapseUnderscoresRegex := regexp.MustCompile(`_+`)                        // Regex to collapse multiple underscores
	safeFilename = collapseUnderscoresRegex.ReplaceAllString(safeFilename, "_") // Replace multiple with single underscore

	if trimmed, found := strings.CutPrefix(safeFilename, "_"); found { // Remove leading underscore if present
		safeFilename = trimmed
	}

	invalidPre := fmt.Sprintf("_%s", ext)                    // Build invalid suffix pattern
	safeFilename = removeSubstring(safeFilename, invalidPre) // Remove it if present

	safeFilename = safeFilename + ext          // Ensure extension is added
	return trimAfterQuestionMark(safeFilename) // Trim everything after ?
}

// trimAfterQuestionMark removes query params from filenames
func trimAfterQuestionMark(input string) string {
	parts := strings.SplitN(input, "?", 2) // Split into 2 parts at first "?"
	return parts[0]                        // Return first part only
}

// removeSubstring removes all occurrences of a substring
func removeSubstring(input string, toRemove string) string {
	result := strings.ReplaceAll(input, toRemove, "") // Replace target substring with ""
	return result
}

// getFileExtension returns the file extension of a path
func getFileExtension(path string) string {
	return filepath.Ext(path) // Extract extension using filepath.Ext
}

// fileExists checks whether a file exists at a given path
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {                // If error occurs, file does not exist
		return false
	}
	return !info.IsDir() // Return true if it's a file (not directory)
}

// downloadFile downloads a file from a URL and saves it locally
func downloadFile(finalURL, outputDir string) bool {
	filename := strings.ToLower(urlToFilename(finalURL)) // Generate safe filename
	filePath := filepath.Join(outputDir, filename)       // Build full path

	if fileExists(filePath) { // Skip if file already exists
		log.Printf("File already exists, skipping: %s", filePath)
		return false
	}

	client := &http.Client{Timeout: 1 * time.Minute} // HTTP client with timeout
	resp, err := client.Get(finalURL)                // Send GET request
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err) // Log error
		return false
	}
	defer resp.Body.Close() // Ensure response body is closed

	if resp.StatusCode != http.StatusOK { // Ensure HTTP status is OK
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return false
	}

	var buf bytes.Buffer                     // Buffer to store response
	written, err := io.Copy(&buf, resp.Body) // Copy response body to buffer
	if err != nil {
		log.Printf("Failed to read PDF data from %s: %v", finalURL, err)
		return false
	}
	if written == 0 { // Check if empty file
		log.Printf("Downloaded 0 bytes for %s; not creating file", finalURL)
		return false
	}

	out, err := os.Create(filePath) // Create file for writing
	if err != nil {
		log.Printf("Failed to create file for %s: %v", finalURL, err)
		return false
	}
	defer out.Close() // Ensure file is closed

	if _, err := buf.WriteTo(out); err != nil { // Write buffer to file
		log.Printf("Failed to write PDF to file for %s: %v", finalURL, err)
		return false
	}

	log.Printf("Successfully downloaded %d bytes: %s → %s", written, finalURL, filePath) // Log success
	return true
}

// directoryExists checks if a given path is a directory
func directoryExists(path string) bool {
	directory, err := os.Stat(path) // Get path info
	if err != nil {
		return false // Path does not exist
	}
	return directory.IsDir() // Return true if directory
}

// createDirectory creates a directory with given permissions
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission) // Attempt to create directory
	if err != nil {
		log.Println(err) // Log error if creation fails
	}
}

// isUrlValid validates if a string is a properly formatted URL
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Try to parse as URL
	return err == nil                  // Valid if no error
}

// removeDuplicatesFromSlice removes duplicate strings
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)  // Map to track seen elements
	var newReturnSlice []string     // Slice to return unique items
	for _, content := range slice { // Loop over input slice
		if !check[content] { // If not seen before
			check[content] = true                            // Mark as seen
			newReturnSlice = append(newReturnSlice, content) // Append to result
		}
	}
	return newReturnSlice
}

// extractFileUrls extracts file URLs with specific extensions from HTML
func extractFileUrls(input string) []string {
	re := regexp.MustCompile(`href="([^"]+\.(?:pdf|png|jpg|webp|zip|rar|stl|7z|json|txt)[^"]*)"`) // Regex to capture file links
	matches := re.FindAllStringSubmatch(input, -1)                                                // Find matches

	var fileLinks []string          // Slice to store links
	for _, match := range matches { // Loop over matches
		if len(match) > 1 {
			fileLinks = append(fileLinks, match[1]) // Extract link
		} else {
			log.Println("Unexpected match format:", match) // Log if invalid format
		}
	}
	return fileLinks
}

// getDataFromURL fetches the body of a URL as a string
func getDataFromURL(uri string) string {
	log.Println("Scraping", uri)   // Log scraping start
	response, err := http.Get(uri) // Send HTTP GET request
	if err != nil {
		log.Println(err) // Log error if GET fails
	}

	body, err := io.ReadAll(response.Body) // Read response body
	if err != nil {
		log.Println(err) // Log read error
	}

	err = response.Body.Close() // Close response body
	if err != nil {
		log.Println(err) // Log close error
	}
	return string(body) // Return response body as string
}
