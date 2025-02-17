package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type Response struct {
	Answer struct {
		Text string `json:"text"`
	} `json:"answer"`
	Sources []struct {
		Content    string `json:"content"`
		Date       any    `json:"date"`
		ID         int    `json:"id"`
		Link       string `json:"link"`
		SourceName any    `json:"source_name"`
		Text       string `json:"text"`
	} `json:"sources"`
	MessageID int `json:"message_id"`
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// comment the line below out if you are not using .env file
	apiKey := os.Getenv("MENDABLE_API_KEY")

	// uncomment the line below and add your api key if you are not using .env file
	// apiKey := "MENDABLE_API_KEY"

	// Parse command-line arguments
	cmd := flag.String("cmd", "", "Command to execute")

	// if the user types -h or --help or help, print the help menu
	// help := flag.String("h", "", "help")
	help := flag.Bool("h", false, "help")
	flag.Parse()

	// add help case as -h or --help or help
	if *help == true {
		fmt.Println("Usage: go run seadog.go [command] [arguments]")
		fmt.Println("Commands:")
		fmt.Println("  -cmd [arguments] - Command followed by argument to execute")
		fmt.Println("  -h - help")
		// Arguemnts commands
		fmt.Println("Arguments:")
		fmt.Println("  ask - Ask a question")
		fmt.Println("  ingest - Ingest data")

		// Exit the program
		os.Exit(0)
	}

	switch *cmd {
	case "ask":
		response, err := newConversation(apiKey)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		// printResponse(response)

		// Get the conversation_id from the response
		conversationID := response["conversation_id"]

		// create a scanner to read user input
		scanner := bufio.NewScanner(os.Stdin)

		// loop until exit condition is met
		for {
			// get the question from the user
			fmt.Print("Ask a question (type 'quit' to exit): ")
			scanner.Scan()
			questionString := scanner.Text()

			// check if user wants to exit
			if questionString == "quit" {
				break
			}

			url := "https://api.mendable.ai/v0/mendableChat"

			data := map[string]interface{}{
				"api_key":         apiKey,
				"question":        questionString,
				"history":         []interface{}{},
				"conversation_id": conversationID,
				"shouldStream":    false,
			}

			payload, err := json.Marshal(data)
			if err != nil {
				panic(err)
			}

			req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
			if err != nil {
				panic(err)
			}

			req.Header.Set("Accept", "text/event-stream")
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var result Response
			if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
				fmt.Println("Can not unmarshal JSON")
			}

			fmt.Println(result.Answer.Text)

			fmt.Println("Sources:" + "\n" + "--------")
			for _, source := range result.Sources {
				fmt.Println(source.Link)
				fmt.Println()
			}
		}

	case "ingest":

		scanner := bufio.NewScanner(os.Stdin)

		// Loop until exit condition is met
		for {
			// get the data source from the user
			fmt.Print("Enter a data source url (type 'quit' to exit): ")
			scanner.Scan()
			dataSource := scanner.Text()
			if dataSource == "quit" {
				break
			}
			if dataSource == "" {
				fmt.Fprintln(os.Stderr, "url and type are required")
				os.Exit(1)
			}
			// get type of ingestion from the user
			fmt.Print("Enter the type of data ingestion (url only for now): ")
			scanner.Scan()
			dataType := scanner.Text()

			if dataType == "" {
				fmt.Fprintln(os.Stderr, "url and type are required")
				os.Exit(1)
			}
			response, err := ingestData(apiKey, dataSource, dataType)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			printResponse(response)

		}
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
		os.Exit(1)
	}
}

func newConversation(apiKey string) (map[string]interface{}, error) {
	url := "https://api.mendable.ai/v0/newConversation"

	// Create the request body
	requestBody := map[string]interface{}{
		"api_key": apiKey,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Send the HTTP request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func ingestData(apiKey string, dataSource string, dataType string) (map[string]interface{}, error) {
	url := "https://api.mendable.ai/v0/ingestData"

	// Create the request body
	requestBody := map[string]interface{}{
		"api_key": apiKey,
		"url":     dataSource,
		"type":    dataType,
	}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// Send the HTTP request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func printResponse(response map[string]interface{}) {
	// Convert the response map into a JSON string and print the response
	responseJSON, err := json.Marshal(response)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(responseJSON))
}

// PrettyPrint to print struct in a readable way
func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
