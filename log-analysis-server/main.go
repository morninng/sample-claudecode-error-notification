package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type PubSubMessage struct {
	Message struct {
		Data       string            `json:"data"`
		Attributes map[string]string `json:"attributes"`
	} `json:"message"`
}

type LogEntry struct {
	Severity    string                 `json:"severity"`
	TextPayload string                 `json:"textPayload"`
	JsonPayload map[string]interface{} `json:"jsonPayload"`
	Timestamp   string                 `json:"timestamp"`
	Resource    struct {
		Type   string            `json:"type"`
		Labels map[string]string `json:"labels"`
	} `json:"resource"`
	Labels map[string]string `json:"labels"`
}

type SlackMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	ThreadTS string `json:"thread_ts,omitempty"`
}

type SlackResponse struct {
	Ok      bool   `json:"ok"`
	Channel string `json:"channel"`
	TS      string `json:"ts"`
	Error   string `json:"error,omitempty"`
}

type ClaudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []ClaudeMessage `json:"messages"`
}

type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ClaudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

func main() {
	http.HandleFunc("/", handlePubSubPush)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting log-analysis-server on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handlePubSubPush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var pubsubMsg PubSubMessage
	if err := json.NewDecoder(r.Body).Decode(&pubsubMsg); err != nil {
		log.Printf("Error decoding Pub/Sub message: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(pubsubMsg.Message.Data)
	if err != nil {
		log.Printf("Error decoding base64 data: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Parse log entry
	var logEntry LogEntry
	if err := json.Unmarshal(data, &logEntry); err != nil {
		log.Printf("Error parsing log entry: %v", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("Received log entry: severity=%s", logEntry.Severity)

	// Process the error log asynchronously
	go processErrorLog(logEntry)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func processErrorLog(logEntry LogEntry) {
	// 1. Send Slack notification and get thread_ts
	threadTS, err := sendSlackNotification(logEntry)
	if err != nil {
		log.Printf("Error sending Slack notification: %v", err)
		return
	}

	log.Printf("Sent Slack notification, thread_ts: %s", threadTS)

	// 2. Get GitHub repository code
	repoCode, err := getGitHubCode()
	if err != nil {
		log.Printf("Error getting GitHub code: %v", err)
		return
	}

	log.Printf("Retrieved GitHub code, total size: %d bytes", len(repoCode))

	// 3. Analyze with Claude API
	analysis, err := analyzeWithClaude(logEntry, repoCode)
	if err != nil {
		log.Printf("Error analyzing with Claude: %v", err)
		return
	}

	log.Printf("Received Claude analysis")

	// 4. Send analysis result to Slack thread
	if err := sendSlackThreadReply(threadTS, analysis); err != nil {
		log.Printf("Error sending Slack thread reply: %v", err)
		return
	}

	log.Printf("Sent analysis to Slack thread")
}

func sendSlackNotification(logEntry LogEntry) (string, error) {
	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	slackChannel := os.Getenv("SLACK_CHANNEL")

	if slackToken == "" || slackChannel == "" {
		return "", fmt.Errorf("SLACK_BOT_TOKEN or SLACK_CHANNEL not set")
	}

	// Format log message
	logText := fmt.Sprintf("*Error Log Detected*\n```\nSeverity: %s\nTimestamp: %s\nPayload: %s\n```",
		logEntry.Severity,
		logEntry.Timestamp,
		logEntry.TextPayload)

	slackMsg := SlackMessage{
		Channel: slackChannel,
		Text:    logText,
	}

	msgJSON, err := json.Marshal(slackMsg)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(msgJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+slackToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var slackResp SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return "", err
	}

	if !slackResp.Ok {
		return "", fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return slackResp.TS, nil
}

func getGitHubCode() (string, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")
	githubRepo := os.Getenv("GITHUB_REPOSITORY")

	if githubToken == "" || githubRepo == "" {
		return "", fmt.Errorf("GITHUB_TOKEN or GITHUB_REPOSITORY not set")
	}

	// Parse owner/repo
	parts := strings.Split(githubRepo, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid GITHUB_REPOSITORY format, expected owner/repo")
	}
	owner, repo := parts[0], parts[1]

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Get repository tree
	tree, _, err := client.Git.GetTree(ctx, owner, repo, "main", true)
	if err != nil {
		return "", err
	}

	var allCode strings.Builder
	allCode.WriteString("# Repository Code\n\n")

	// Fetch content of all files
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			path := entry.GetPath()

			// Skip binary files and large files
			if shouldSkipFile(path) {
				continue
			}

			fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)
			if err != nil {
				log.Printf("Error getting file %s: %v", path, err)
				continue
			}

			if fileContent != nil {
				content, err := fileContent.GetContent()
				if err != nil {
					log.Printf("Error decoding file %s: %v", path, err)
					continue
				}

				allCode.WriteString(fmt.Sprintf("\n## File: %s\n```\n%s\n```\n", path, content))
			}
		}
	}

	return allCode.String(), nil
}

func shouldSkipFile(path string) bool {
	// Skip common binary and unnecessary files
	skipExtensions := []string{
		".png", ".jpg", ".jpeg", ".gif", ".ico", ".pdf",
		".zip", ".tar", ".gz", ".exe", ".dll", ".so",
		".lock", ".sum",
	}
	skipPaths := []string{
		".git/", "node_modules/", "vendor/", ".terraform/",
	}

	for _, ext := range skipExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	for _, skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			return true
		}
	}

	return false
}

func analyzeWithClaude(logEntry LogEntry, repoCode string) (string, error) {
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		return "", fmt.Errorf("ANTHROPIC_API_KEY not set")
	}

	// Create prompt
	prompt := fmt.Sprintf(`You are a software engineer analyzing an error log from a production system.

Error Log:
- Severity: %s
- Timestamp: %s
- Message: %s

Repository Code:
%s

Please analyze this error and provide:
1. Root cause of the error
2. Which part of the code is causing this issue
3. Suggested fix

Keep your response concise and actionable.`,
		logEntry.Severity,
		logEntry.Timestamp,
		logEntry.TextPayload,
		repoCode)

	claudeReq := ClaudeRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 2048,
		Messages: []ClaudeMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	reqJSON, err := json.Marshal(claudeReq)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(reqJSON))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", anthropicKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("claude API error: %s", string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return "", err
	}

	if len(claudeResp.Content) == 0 {
		return "", fmt.Errorf("no content in Claude response")
	}

	return claudeResp.Content[0].Text, nil
}

func sendSlackThreadReply(threadTS, message string) error {
	slackToken := os.Getenv("SLACK_BOT_TOKEN")
	slackChannel := os.Getenv("SLACK_CHANNEL")

	if slackToken == "" || slackChannel == "" {
		return fmt.Errorf("SLACK_BOT_TOKEN or SLACK_CHANNEL not set")
	}

	slackMsg := SlackMessage{
		Channel:  slackChannel,
		Text:     fmt.Sprintf("*Claude Analysis*\n%s", message),
		ThreadTS: threadTS,
	}

	msgJSON, err := json.Marshal(slackMsg)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://slack.com/api/chat.postMessage", bytes.NewBuffer(msgJSON))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+slackToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var slackResp SlackResponse
	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return err
	}

	if !slackResp.Ok {
		return fmt.Errorf("slack API error: %s", slackResp.Error)
	}

	return nil
}
