
```bash
curl -sL cli.genkit.dev | bash
curl -sL cli.genkit.dev | upgrade=true bash

genkit start -- go run .
```



```ts
import { genkit } from 'genkit';
import { googleAI } from '@genkit-ai/google-genai';

const ai = genkit({ plugins: [googleAI()] });

const { text } = await ai.generate({
    model: googleAI.model('gemini-2.5-flash'),
    prompt: 'Why is Firebase awesome?'
});
```


## .env 파일 생성
```bash
OLLAMA_SERVER_ADDRESS="http://localhost:11434"
MODEL_NAME="qwen2.5-coder:latest"
```

## main.go 실행
```bash
go run main.go
```

## curl로 테스트
```bash
curl -X POST http://localhost:8080/codingFlow \
  -H "Content-Type: application/json" \
  -d '{"data": "피보나치 수열을 구하는 파이썬 함수를 만들어줘"}'
