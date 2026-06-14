package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	server string
	token  string
	client = &http.Client{}
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 解析全局参数
	globalFlags := flag.NewFlagSet("moon-cli", flag.ExitOnError)
	globalFlags.StringVar(&server, "server", "http://localhost:8080", "server address")
	globalFlags.StringVar(&token, "token", "", "auth token (required, or set MOON_AUTH_TOKEN env var)")
	
	// 先解析全局参数（在子命令之前）
	args := os.Args[1:]
	cmdIndex := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		// 检查是否是子命令
		if arg == "hl" || arg == "highlights" || arg == "books" || arg == "export" || arg == "help" {
			cmdIndex = i
			break
		}
		// 如果是-flag形式，跳过这个参数和它的值
		if strings.HasPrefix(arg, "-") {
			// 如果是-flag value形式（如 -token xxx），需要跳过两个参数
			// 但如果是-flag=value形式（如 -token=xxx），只需要跳过一个参数
			if strings.Contains(arg, "=") {
				continue
			}
			// -flag value形式，跳过下一个参数
			if i+1 < len(args) {
				i++ // 跳过参数值
			}
			continue
		}
		// 不是以-开头的参数，可能是命令
		cmdIndex = i
		break
	}
	
	if cmdIndex > 0 {
		// 解析全局参数
		if err := globalFlags.Parse(args[:cmdIndex]); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing global flags: %v\n", err)
			os.Exit(1)
		}
		args = args[cmdIndex:]
	} else {
		if err := globalFlags.Parse(args); err != nil {
			fmt.Fprintf(os.Stderr, "error parsing global flags: %v\n", err)
			os.Exit(1)
		}
	}

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	cmd := args[0]
	subArgs := args[1:]

	var err error
	switch cmd {
	case "hl", "highlights":
		err = handleHighlights(subArgs)
	case "books":
		err = handleBooks(subArgs)
	case "export":
		err = handleExport(subArgs)
	case "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: moon-cli [global-options] <command> [command-options]

Commands:
  hl|highlights <action>  Manage highlights
  books <action>          Manage books
  export <format>         Export data

Global Options:
  -server string       Server address (default "http://localhost:8080")
  -token string        Auth token (required, or set MOON_AUTH_TOKEN env var)

Examples:
  # List all highlights
  moon-cli hl list

  # Create a highlight
  moon-cli hl create --title "书名" --text "高亮内容" --author "作者"

  # Get a highlight
  moon-cli hl get 1

  # Update a highlight
  moon-cli hl update 1 --text "新内容" --note "新备注"

  # Delete a highlight
  moon-cli hl delete 1

  # Delete all highlights
  moon-cli hl delete --all

  # List books
  moon-cli books list

  # Get a book
  moon-cli books get 1

  # Export as JSON
  moon-cli export json

  # Export as CSV
  moon-cli export csv

Environment Variables:
  MOON_AUTH_TOKEN      Auth token (alternative to -token flag)
`)
}

func handleHighlights(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("highlights command requires a subcommand: list, create, get, update, delete")
	}

	action := args[0]
	actionArgs := args[1:]

	switch action {
	case "list":
		return listHighlights(actionArgs)
	case "create":
		return createHighlight(actionArgs)
	case "get":
		return getHighlight(actionArgs)
	case "update":
		return updateHighlight(actionArgs)
	case "delete":
		return deleteHighlight(actionArgs)
	default:
		return fmt.Errorf("unknown highlights action: %s", action)
	}
}

func handleBooks(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("books command requires a subcommand: list, get")
	}

	action := args[0]
	actionArgs := args[1:]

	switch action {
	case "list":
		return listBooks(actionArgs)
	case "get":
		return getBook(actionArgs)
	default:
		return fmt.Errorf("unknown books action: %s", action)
	}
}

func handleExport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("export command requires a format: json, csv")
	}

	format := args[0]
	formatArgs := args[1:]

	switch format {
	case "json":
		return exportJSON(formatArgs)
	case "csv":
		return exportCSV(formatArgs)
	default:
		return fmt.Errorf("unknown export format: %s", format)
	}
}

// parseToken 解析token，优先使用命令行参数，其次使用环境变量
func parseToken() error {
	if token == "" {
		token = os.Getenv("MOON_AUTH_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("token is required (use -token flag or MOON_AUTH_TOKEN env var)")
	}
	return nil
}

func doRequest(method, path string, body interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, server+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/csv") {
		fmt.Print(string(respBody))
		return nil
	}

	var out bytes.Buffer
	if err := json.Indent(&out, respBody, "", "  "); err != nil {
		fmt.Println(string(respBody))
		return nil
	}
	fmt.Println(out.String())
	return nil
}

func listHighlights(args []string) error {
	fs := flag.NewFlagSet("highlights list", flag.ExitOnError)
	fs.Parse(args)
	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("GET", "/api/moon/v2/highlights/", nil)
}

func createHighlight(args []string) error {
	fs := flag.NewFlagSet("highlights create", flag.ExitOnError)
	title := fs.String("title", "", "book title (required)")
	text := fs.String("text", "", "highlight text (required)")
	author := fs.String("author", "", "author")
	chapter := fs.String("chapter", "", "chapter")
	note := fs.String("note", "", "note")
	color := fs.String("color", "", "color")
	fs.Parse(args)

	if *title == "" || *text == "" {
		return fmt.Errorf("--title and --text are required")
	}

	highlight := map[string]string{"title": *title, "text": *text}
	if *author != "" {
		highlight["author"] = *author
	}
	if *chapter != "" {
		highlight["chapter"] = *chapter
	}
	if *note != "" {
		highlight["note"] = *note
	}
	if *color != "" {
		highlight["color"] = *color
	}

	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("POST", "/api/moon/v2/highlights/", map[string]interface{}{
		"highlights": []interface{}{highlight},
	})
}

func getHighlight(args []string) error {
	fs := flag.NewFlagSet("highlights get", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("highlight ID is required")
	}

	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid highlight ID: %s", fs.Arg(0))
	}

	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("GET", fmt.Sprintf("/api/moon/v2/highlights/%d", id), nil)
}

func updateHighlight(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("highlight ID is required")
	}

	// 第一个参数是ID
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid highlight ID: %s", args[0])
	}

	// 解析剩余参数
	fs := flag.NewFlagSet("highlights update", flag.ExitOnError)
	text := fs.String("text", "", "highlight text")
	note := fs.String("note", "", "note")
	title := fs.String("title", "", "book title")
	author := fs.String("author", "", "author")
	source := fs.String("source", "", "source")
	color := fs.String("color", "", "color")
	fs.Parse(args[1:])

	updates := make(map[string]interface{})
	if *text != "" {
		updates["text"] = *text
	}
	if *note != "" {
		updates["note"] = *note
	}
	if *title != "" {
		updates["title"] = *title
	}
	if *author != "" {
		updates["author"] = *author
	}
	if *source != "" {
		updates["source"] = *source
	}
	if *color != "" {
		updates["color"] = *color
	}

	if len(updates) == 0 {
		return fmt.Errorf("at least one field to update is required")
	}

	if err := parseToken(); err != nil {
		return err
	}

	// 发送PATCH请求
	data, err := json.Marshal(updates)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequest("PATCH", server+fmt.Sprintf("/api/moon/v2/highlights/%d", id), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Token "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	var out bytes.Buffer
	if err := json.Indent(&out, respBody, "", "  "); err != nil {
		fmt.Println(string(respBody))
		return nil
	}
	fmt.Println(out.String())
	return nil
}

func deleteHighlight(args []string) error {
	fs := flag.NewFlagSet("highlights delete", flag.ExitOnError)
	all := fs.Bool("all", false, "delete all highlights")
	fs.Parse(args)

	// 删除全部
	if *all {
		if err := parseToken(); err != nil {
			return err
		}
		return doRequest("DELETE", "/api/moon/v2/highlights/", nil)
	}

	// 删除单个高亮
	if fs.NArg() == 0 {
		return fmt.Errorf("highlight ID is required")
	}

	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid highlight ID: %s", fs.Arg(0))
	}

	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("DELETE", fmt.Sprintf("/api/moon/v2/highlights/%d", id), nil)
}

func listBooks(args []string) error {
	fs := flag.NewFlagSet("books list", flag.ExitOnError)
	fs.Parse(args)
	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("GET", "/api/moon/v2/books/", nil)
}

func getBook(args []string) error {
	fs := flag.NewFlagSet("books get", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() == 0 {
		return fmt.Errorf("book ID is required")
	}

	id, err := strconv.ParseInt(fs.Arg(0), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid book ID: %s", fs.Arg(0))
	}

	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("GET", fmt.Sprintf("/api/moon/v2/books/%d", id), nil)
}

func exportJSON(args []string) error {
	fs := flag.NewFlagSet("export json", flag.ExitOnError)
	fs.Parse(args)
	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("GET", "/api/moon/export/json", nil)
}

func exportCSV(args []string) error {
	fs := flag.NewFlagSet("export csv", flag.ExitOnError)
	fs.Parse(args)
	if err := parseToken(); err != nil {
		return err
	}
	return doRequest("GET", "/api/moon/export/csv", nil)
}