package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"os"
	"os/exec"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/ollama"
	"github.com/firebase/genkit/go/plugins/server"
	"github.com/joho/godotenv"
)

/*
*

	curl -X POST http://localhost:8080/codingFlow \
	  -H "Content-Type: application/json" \
	  -d '{"data": "피보나치 수열을 구하는 파이썬 함수를 만들어줘"}'
*/
func main() {
	_ = godotenv.Load()
	ctx := context.Background()

	ollamaServerUrl := os.Getenv("OLLAMA_SERVER_ADDRESS")
	modelName := os.Getenv("MODEL_NAME")
	// 1. Genkit 초기화 (Ollama 플러그인 사용)
	// Ollama 플러그인은 WithPlugins에 구조체 포인터를 전달하여 초기화합니다.
	ollamaPlugin := &ollama.Ollama{
		ServerAddress: ollamaServerUrl,
	}
	g := genkit.Init(ctx, genkit.WithPlugins(ollamaPlugin))

	// 사용할 모델 정의 (Ollama Qwen 2.5 7B)
	// DefineModel을 사용하여 모델을 명시적으로 등록해야 합니다.
	model := ollamaPlugin.DefineModel(g, ollama.ModelDefinition{
		Name: modelName, //"qwen2.5-coder:latest",
		Type: "chat",
	}, nil)

	// 2. Reflection 에이전트 Flow 정의
	genkit.DefineFlow(g, "codingFlow", func(ctx context.Context, input string) (string, error) {
		if input == "" {
			return "", fmt.Errorf("요청할 코딩 내용을 입력해주세요.")
		}

		fmt.Printf("\n🚀 요청 시작: %s\n", input)

		currentCode := ""
		feedback := ""

		// --- 루프 시작 (최대 3회 반복) ---
		for i := 0; i < 3; i++ {
			fmt.Printf("\n--- [Cycle %d] ---\n", i+1)

			// [단계 1] Generator: 코드 작성
			// 피드백이 있으면 반영하고, 없으면 새로 작성
			genPrompt := fmt.Sprintf(`
				Role: Python Expert
				Request: %s
				
				[Previous Code]:
				%s
				
				[Feedback]:
				%s
				
				Task:
				1. Write/Fix Python code based on request and feedback.
				2. Output ONLY the code block.
			`, input, currentCode, feedback)

			genResp, err := genkit.GenerateText(ctx, g,
				ai.WithModel(model),
				ai.WithPrompt(genPrompt),
			)
			if err != nil {
				return "", err
			}
			currentCode = genResp

			// Markdown Code Block 제거 logic 추가
			currentCode = strings.ReplaceAll(currentCode, "```python", "")
			currentCode = strings.ReplaceAll(currentCode, "```", "")
			currentCode = strings.TrimSpace(currentCode)

			fmt.Printf("📝 작성된 코드 길이: %d bytes\n", len(currentCode))

			if len(currentCode) < 10 {
				fmt.Println("⚠️코드가 너무 짧습니다. 다시 작성해주세요.")
				continue
			}

			// [단계 2] Verifier: 코드 검증
			verPrompt := fmt.Sprintf(`
				Role: Python Code Verifier
				
				Code to verify:
				%s
				
				Task:
				1. Check for syntax error and logic error.
				2. If the code is safe and valid, reply with "VALID".
				3. If there are errors, describe them briefly.
			`, currentCode)

			verResp, err := genkit.GenerateText(ctx, g,
				ai.WithModel(model),
				ai.WithPrompt(verPrompt),
			)
			if err != nil {
				return "", err
			}
			verificationResult := strings.TrimSpace(verResp)
			fmt.Printf("🔍 검증 결과: %s\n", verificationResult)

			if !strings.Contains(strings.ToUpper(verificationResult), "VALID") {
				fmt.Println("⚠️코드가 유효하지 않습니다. 다시 작성해주세요.")
				continue
			}

			// [단계 3] Reviewer: 코드 평가
			revPrompt := fmt.Sprintf(`
				Role: Strict Code Reviewer
				
				Code to review:
				%s
				
				Verification Result:
				%s
				
				Task:
				1. If the code is perfect/good AND Verification Result is VALID, reply with exactly "APPROVE".
				2. If not, provide short, constructive feedback. Consider the Verification Result.
			`, currentCode, verificationResult)

			// 리뷰는 창의성이 필요 없으므로 온도를 낮춤
			revResp, err := genkit.GenerateText(ctx, g,
				ai.WithModel(model),
				ai.WithPrompt(revPrompt),
			)
			if err != nil {
				return "", err
			}
			feedback = strings.TrimSpace(revResp)
			fmt.Printf("🧐 리뷰 결과: %s\n", feedback)

			// [단계 4] 판단
			if strings.Contains(strings.ToUpper(feedback), "APPROVE") {
				fmt.Println("🎉 승인 완료! 실행 테스트를 진행합니다...")

				// [단계 5] Execution Tester: 실제 실행 확인
				// 1. 코드를 임시 파일로 저장
				fileName := "temp.py"
				if err := os.WriteFile(fileName, []byte(currentCode), 0644); err != nil {
					fmt.Printf("⚠️ 파일 저장 실패: %v\n", err)
					feedback = fmt.Sprintf("System Error: Failed to save code to file: %v", err)
					continue
				}

				// 2. 파이썬 코드 실행
				cmd := exec.Command("python3", fileName)
				output, err := cmd.CombinedOutput()
				if err != nil {
					fmt.Printf("❌ 실행 실패:\n%s\n", string(output))
					feedback = fmt.Sprintf("Reviewer Approved, but Execution Failed.\n\nError Output:\n%s", string(output))
					continue
				}

				fmt.Printf("✅ 실행 성공!\nOutput:\n%s\n", string(output))
				break
			}
		}

		return currentCode, nil
	})

	// 3. 서버 실행
	// Genkit 개발자 UI 또는 curl 명령어로 호출 가능하도록 서버를 띄웁니다.
	mux := http.NewServeMux()
	for _, a := range genkit.ListFlows(g) {
		mux.HandleFunc("POST /"+a.Name(), genkit.Handler(a))
	}

	fmt.Println("Running server on localhost:8080...")
	log.Fatal(server.Start(ctx, "127.0.0.1:8080", mux))
}
