# Excel Agent

Genkit을 사용하여 Excel 및 Google Sheets 데이터를 JSON으로 변환하고, AI를 통해 Go 구조체를 자동으로 생성하는 도구입니다.

## 프로젝트 구조 (Tree)

```text
excel/
├── main.go             # 진입점 및 CLI/Genkit 초기화
├── internal/
│   ├── cmd/            # CLI 플래그 파싱 및 핸들링
│   ├── config/         # 설정 관리 (env load)
│   ├── flows/          # Genkit Flow 정의 및 도구 등록
│   └── processor/      # 비즈니스 로직 (Excel, Sheets, Generator, Redis, Tool Logic)
├── xlsx/               # 원본 .xlsx 파일 저장 폴더
├── json/               # 변환된 .json 파일 저장 폴더
├── data/               # AI로 생성된 .go 구조체 파일 저장 폴더
├── go.mod/go.sum       # 의존성 관리
└── .env                # 환경 변수 설정
```

## 빌드 방법

프로젝트 루트 폴더(`excel/`)에서 다음 명령어를 실행합니다:

```bash
go build -o excel-agent main.go
```

## 사용 방법

### 1. 커맨드 라인 (CLI) 모드
특정 작업을 직접 실행할 때 사용합니다.

- **로컬 엑셀 파일 처리**:
  ```bash
  ./excel-agent -cmd xlsx
  ```
- **구글 스프레드시트 처리**:
  ```bash
  ./excel-agent -cmd sheets -id <spreadsheet_id>
  ```
- **Go 구조체 생성 (AI)**:
  ```bash
  ./excel-agent -cmd gen -file <filename.json>
  ```
- **Redis 데이터 캐싱**:
  ```bash
  ./excel-agent -cmd redis
  ```
- **Redis 데이터 조회 (AI Agent)**:
  사용자의 자연어 질문을 분석하여 적절한 Redis 데이터를 찾아 답변을 생성합니다.
  ```bash
  ./excel-agent -cmd query -key "Character:UnitData에서 10개만 보여줘"
  ```

### 2. Genkit 에이전트 모드
UI를 통해 Flow를 확인하거나 대기 모드로 실행할 때 사용합니다. 옵션 없이 실행하면 대기 모드로 들어갑니다.

```bash
# 에이전트 실행
./excel-agent

# Genkit UI 실행 (개발 모드)
GENKIT_ENV=dev genkit start
```

## 환경 변수 설정 (.env)

- `GEMINI_API_KEY`: Google AI API 키
- `GOOGLE_SHEET_ID`: (선택) 기본 구글 시트 ID
- `REDIS_ADDR`: Redis 서버 주소 (기본값: `localhost:6379`)
- `REDIS_DB`: Redis DB 인덱스 (기본값: `0`)
- `XLSX_DIR`: (선택) 엑셀 파일 기본 경로 (기본값: `xlsx`)
- `JSON_DIR`: (선택) JSON 출력 기본 경로 (기본값: `json`)
- `DATA_DIR`: (선택) Go 구조체 출력 기본 경로 (기본값: `data`)
- `DEFAULT_MODEL`: (선택) AI 모델 (기본값: `googleai/gemini-2.5-flash`)
