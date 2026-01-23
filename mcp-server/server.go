// Run with: go run server.go

package main

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// 1. MCP 서버 인스턴스 생성
	s := server.NewMCPServer(
		"text-utilities",
		"1.0.0",
	)

	// --- Tool 1: Encode/decode text ---
	encodeTool := mcp.NewTool("text_encode",
		mcp.WithDescription("Encode or decode text using various methods"),
		mcp.WithString("text", mcp.Description("Text to encode/decode")),
		mcp.WithString("method", mcp.Description("Method: base64_encode, base64_decode, url_encode")),
	)

	s.AddTool(encodeTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 입력값 파싱
		// 입력값 파싱
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			// 인자가 없거나 올바르지 않은 형식이면 에러 처리 혹은 기본값
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		text, _ := args["text"].(string)
		method, _ := args["method"].(string)

		fmt.Printf("🔧 Executing text_encode: %s (%s)\n", method, text)

		var result map[string]interface{}

		switch method {
		case "base64_encode":
			encoded := base64.StdEncoding.EncodeToString([]byte(text))
			result = map[string]interface{}{"original": text, "method": method, "result": encoded}
		case "base64_decode":
			decoded, err := base64.StdEncoding.DecodeString(text)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("invalid base64: %v", err)), nil
			}
			result = map[string]interface{}{"original": text, "method": method, "result": string(decoded)}
		case "url_encode":
			encoded := strings.ReplaceAll(text, " ", "%20")
			encoded = strings.ReplaceAll(encoded, "&", "%26")
			result = map[string]interface{}{"original": text, "method": method, "result": encoded}
		default:
			return mcp.NewToolResultError(fmt.Sprintf("unsupported method: %s", method)), nil
		}

		return jsonResult(result)
	})

	// --- Tool 2: Generate hashes ---
	hashTool := mcp.NewTool("hash_generate",
		mcp.WithDescription("Generate hash values for text"),
		mcp.WithString("text", mcp.Description("Text to hash")),
		mcp.WithString("type", mcp.Description("Hash type: md5, sha256")),
	)

	s.AddTool(hashTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		text, _ := args["text"].(string)
		hashType, _ := args["type"].(string)

		fmt.Printf("🔧 Executing hash_generate: %s\n", hashType)

		var result map[string]interface{}

		switch hashType {
		case "md5":
			hash := md5.Sum([]byte(text))
			result = map[string]interface{}{"original": text, "type": hashType, "hash": hex.EncodeToString(hash[:])}
		case "sha256":
			hash := sha256.Sum256([]byte(text))
			result = map[string]interface{}{"original": text, "type": hashType, "hash": hex.EncodeToString(hash[:])}
		default:
			return mcp.NewToolResultError(fmt.Sprintf("unsupported hash type: %s", hashType)), nil
		}

		return jsonResult(result)
	})

	// --- Tool 3: Fetch URL content ---
	fetchTool := mcp.NewTool("fetch_url",
		mcp.WithDescription("Fetch content from a URL"),
		mcp.WithString("url", mcp.Description("URL to fetch content from")),
	)

	s.AddTool(fetchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}
		url, _ := args["url"].(string)

		fmt.Printf("🔧 Executing fetch_url: %s\n", url)

		resp, err := http.Get(url)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to fetch URL: %v", err)), nil
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to read response body: %v", err)), nil
		}

		result := map[string]interface{}{
			"url":     url,
			"status":  resp.StatusCode,
			"content": string(body), // 너무 길면 자르는 로직 추가 권장
			"headers": resp.Header,
			"length":  len(body),
		}

		return jsonResult(result)
	})

	// 4. SSE 서버 설정 및 구동
	sseServer := server.NewSSEServer(s, server.WithMessageEndpoint("/messages"))

	// SSE 엔드포인트 (Client가 듣는 곳)
	http.Handle("/sse", sseServer)
	// 메시지 엔드포인트 (Client가 요청 보내는 곳)
	http.Handle("/messages", sseServer)

	fmt.Println("🚀 SSE MCP Server running on http://localhost:8080")
	fmt.Println("   - Endpoint: http://localhost:8080/sse")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// 헬퍼 함수: Map을 JSON 문자열 결과로 변환
func jsonResult(data map[string]interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return mcp.NewToolResultError("failed to marshal result"), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}
