package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color" // Import the color package
	"golang.org/x/net/html"
)

// Define a struct to store the form information
type FormInfo struct {
	Method string   // The form method (GET or POST)
	Inputs []string // The input names
}

func getForm(url string) []FormInfo {
	// Get the response from the URL
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error getting", url, err)
		return nil
	}
	defer resp.Body.Close()

	// Parse the response body as HTML
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Error parsing", url, err)
		return nil
	}

	// Find all the form elements using a recursive function
	var forms []*html.Node // Use a slice of nodes instead of a single node
	var findForm func(*html.Node)
	findForm = func(n *html.Node) {
		// Check if the node is a form element and append it to the slice
		if n.Type == html.ElementNode && n.Data == "form" {
			forms = append(forms, n)
			// Do not return here, keep looking for more forms
		}
		// Recurse through the node's children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findForm(c)
		}
	}
	findForm(doc)

	// If no form is found, return nil
	if len(forms) == 0 {
		return nil
	}

	// Create a slice of FormInfo structs to store the form information
	var infos []FormInfo

	// Loop through the form elements and fill the structs
	for _, form := range forms {
		// Create an empty FormInfo struct
		var info FormInfo

		// Loop through the form element's attributes and find the method
		for _, attr := range form.Attr {
			if attr.Key == "method" {
				info.Method = attr.Val // Store the method in the struct
				break                  // No need to loop further
			}
		}

		// Loop through the form element's children and find the input elements
		for c := form.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "input" {
				// Loop through the input element's attributes and find the name
				for _, attr := range c.Attr {
					if attr.Key == "name" {
						info.Inputs = append(info.Inputs, attr.Val) // Append the name to the struct's slice
						break                                       // No need to loop further
					}
				}
			}
		}

		// Append the FormInfo struct to the slice
		infos = append(infos, info)
	}

	// Return the slice of FormInfo structs
	return infos
}

func main() {
	// Define some colors as functions
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()

	// Parse the -f flag to get the file name
	fileName := flag.String("f", "", "file name with URLs")
	flag.Parse()

	// Check if the file name is valid
	if *fileName == "" {
		fmt.Println("Please provide a file name with -f flag")
		os.Exit(1)
	}

	// Open the file and defer closing it
	file, err := os.Open(*fileName)
	if err != nil {
		fmt.Println("Error opening", *fileName, err)
		os.Exit(1)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Loop through the lines and print all the form information if it exists with colors
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		// Check if the url is empty or invalid
		if url == "" || !strings.HasPrefix(url, "http") {
			fmt.Println("This is not a valid URL:", url)
			continue // Skip this line and go to the next one
		}
		// Get all the form information from the url
		infos := getForm(url)
		if len(infos) > 0 {
			// Print the url with green color
			fmt.Println(green(url))
			// Print the number of forms found
			fmt.Println("Found", len(infos), "forms in this page.")
			// Loop through the infos slice and print the form information
			for i, info := range infos {
				// Check if there is only one form and print accordingly
				if len(infos) == 1 {
					// Print form instead of form1 with yellow color
					fmt.Println(yellow("Form:"))
				} else {
					// Print the form index with yellow color
					fmt.Println(yellow("Form", i+1, ":"))
				}
				// Print the form method with yellow color
				fmt.Println(yellow("Method:"), info.Method)
				// Print the input names with cyan color
				for _, input := range info.Inputs {
					fmt.Println(cyan("Input:"), input)
				}
				fmt.Printf("\n\n____________________________________________________\n\n")

			}
		}
	}

	// Check for any scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning", *fileName, err)
	}
}
