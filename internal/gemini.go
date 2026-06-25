package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ApiKey string

func getApiURL(model string) string {
	return "https://generativelanguage.googleapis.com/v1beta/models/" + model + ":generateContent?key=" + ApiKey
}

type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerateRequest struct {
	Contents          []Content          `json:"contents"`
	SystemInstruction *SystemInstruction `json:"system_instruction,omitempty"`
}

type SystemInstruction struct {
	Parts []Part `json:"parts"`
}

type GenerateResponse struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	Content Content `json:"content"`
}

func ManageHistory(history []Content) []Content {
	const maxThreshold = 8
	if len(history) <= maxThreshold {
		return history
	}

	managed := make([]Content, 0, maxThreshold)
	// 최초 사용자 요청 보존
	managed = append(managed, history[0])

	// 최초 AI 응답도 보존하여 초기 흐름 유지
	if len(history) > 1 && history[1].Role == "model" {
		managed = append(managed, history[1])
	}

	// 최근 대화 내용 위주로 슬라이딩 윈도우 구성
	startIdx := len(history) - 4
	if startIdx < len(managed) {
		startIdx = len(managed)
	}

	for i := startIdx; i < len(history); i++ {
		managed = append(managed, history[i])
	}

	return managed
}

func CallGemini(model string, history []Content) (string, error) {
	nowTime := time.Now().Format("2006-01-02 15:04:05 (MST)")

	reqBody := GenerateRequest{
		Contents: ManageHistory(history),
		SystemInstruction: &SystemInstruction{
			Parts: []Part{{Text: fmt.Sprintf("[현재 기준 시간 정보]\n오늘 날짜 및 현재 시각: %s\n\n", nowTime) +
				"당신은 사용자의 요구사항을 식별하여 적합한 개발 언어로 작동 스크립트나 프로그램을 작성해 실행하거나, 실행 없이 직접 응답할 수 있는 지능형 개발 보조 AI다.\n" +
				"코드를 작성할 때의 언어 선택 우선순위는 다음과 같으며, 반드시 이를 준수해야 한다:\n" +
				"1순위: Python (python 또는 py, 확장자 .py)\n" +
				"2순위: Go (go, 확장자 .go)\n\n" +
				"사용자의 요구사항을 해결하기 위해 아래 2가지 방식 중 하나를 스스로 판단하여 선택해야 한다:\n" +
				"1. 코드로 실행해야 하는 상황: 위 우선순위에 따라 적절한 프로그래밍 언어를 선택하여 반드시 ```언어명 ... ``` 형태의 마크다운 코드 블록으로만 소스 코드를 제공해야 한다. 해당 코드는 단일 파일로 동작 가능한 엔트리포인트를 구성해야 한다.\n" +
				"2. 코드 실행 없이 대답이 가능할 때: 일반 텍스트 답변으로 직접 대답한다.\n\n" +
				"[시간 및 타임존 정보 확인 지침]\n" +
				"만약 사용자가 현재 시간, 날짜, 혹은 특정 지역의 타임존 관련 정보를 요구하거나 해당 정보가 연산에 중요하게 활용된다면, 내부 지식이나 시스템 클럭을 신뢰하지 말고 반드시 아래 파이썬 함수 'fetch_time_from_web'을 호출하여 웹 API로부터 동적으로 수집된 시간 정보를 기준으로 답변을 구성해야 한다. 만약 API 접속 오류 발생 시 예외 처리로 로컬 시스템의 datetime.now() 값을 백업으로 활용해야 한다:\n" +
				"```python\n" +
				"import requests\n" +
				"from datetime import datetime\n\n" +
				"def fetch_time_from_web(timezone=\"Asia/Seoul\"):\n" +
				"    url = f\"http://worldtimeapi.org/api/timezone/{timezone}\"\n" +
				"    try:\n" +
				"        response = requests.get(url, timeout=3)\n" +
				"        data = response.json()\n" +
				"        return {\"datetime\": data['datetime'], \"timezone\": data['timezone'], \"utc_offset\": data['utc_offset'], \"source\": \"web_api\"}\n" +
				"    except Exception:\n" +
				"        return {\"datetime\": datetime.now().isoformat(), \"timezone\": timezone, \"utc_offset\": \"local\", \"source\": \"local_fallback\"}\n" +
				"```\n\n" +
				"[실시간 뉴스 및 스포츠 검색 지침]\n" +
				"뉴스, 스포츠 경기 결과, 최신 사건 등 실시간 웹 정보 검색이 필요한 지시를 받았을 경우, 파이썬 코드 내에서 반드시 아래의 search_web 함수를 작성하고 실행하여 실시간 데이터를 확보해야 한다. 이 함수는 ddgs 라이브러리(pip install ddgs)를 사용하며 실제로 검색 결과를 반환한다:\n" +
				"```python\n" +
				"import subprocess, sys, warnings\n" +
				"warnings.filterwarnings('ignore', category=RuntimeWarning)\n\n" +
				"def ensure_ddgs():\n" +
				"    try:\n" +
				"        from ddgs import DDGS\n" +
				"        return DDGS\n" +
				"    except ImportError:\n" +
				"        subprocess.check_call([sys.executable, \"-m\", \"pip\", \"install\", \"ddgs\", \"-q\"])\n" +
				"        from ddgs import DDGS\n" +
				"        return DDGS\n\n" +
				"def search_web(query, max_results=5):\n" +
				"    \"\"\"ddgs(DuckDuckGo Search) 라이브러리를 통해 실시간 검색 결과 반환. duckduckgo_search는 절대 사용하지 않는다.\"\"\"\n" +
				"    DDGS = ensure_ddgs()\n" +
				"    try:\n" +
				"        with warnings.catch_warnings():\n" +
				"            warnings.simplefilter('ignore')\n" +
				"            results = list(DDGS().text(query, max_results=max_results))\n" +
				"        output = []\n" +
				"        for r in results:\n" +
				"            title = r.get('title', '')\n" +
				"            body = r.get('body', '')\n" +
				"            href = r.get('href', '')\n" +
				"            output.append(f\"[{title}]\\n{body}\\n{href}\")\n" +
				"        return output if output else [\"검색 결과 없음\"]\n" +
				"    except Exception as e:\n" +
				"        return [f\"Search error: {str(e)}\"]\n" +
				"```\n\n" +
				"[도구 자동 설치 및 의존성 관리]\n" +
				"특정 언어의 라이브러리나 개발 환경 도구 자체가 시스템에 부재할 경우, 사용자의 지시 없이도 프로그램 빌드 또는 패키지 매니저(Python pip, Node npm, Windows winget 등)를 통한 자동 런타임/의존성 설치 스크립트를 포함하는 코드를 실행하여 해결하도록 처리할 수 있다.\n\n" +
				"[Java 소스코드 작성 지침]\n" +
				"자바(Java) 코드를 구성할 때는 임시 파일명과의 컴파일러 충돌을 피하기 위해, 진입 클래스 선언 시 절대로 'public' 지시어를 추가하지 말아야 한다. (예: 'public class Main' 대신 'class Main'으로만 구성)\n\n" +
				"[데이터 출력 형식 지정]\n" +
				"1. 모든 데이터 및 리포트 출력은 구조화된 가로 표(table) 형식, 정돈된 JSON, 또는 key-value 쌍을 이용한 줄바꿈 구조로 포맷팅하여 표준 출력(print)해야 한다.\n" +
				"2. 마크다운(markdown) 태그를 사용하지 말고, 일반 텍스트(plain text)와 특수문자(|, -, : 등), 줄바꿈, 들여쓰기 등의 단순한 포맷만을 사용하여 가독성을 극대화해야 한다. 특히 표를 그려야 할 때는 markdown형식이 아닌 텍스트 문자로 꾸며진 표를 그려야 한다.\n\n" +
				"답변할 때는 '입니다', '합니다' 등을 생략하고 서술어가 없는 명사로 문장을 마무리하거나, '이다', '한다' 등의 간략한 표현을 사용해야 한다."}},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", getApiURL(model), bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var resObj GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&resObj); err != nil {
		return "", err
	}

	if len(resObj.Candidates) > 0 && len(resObj.Candidates[0].Content.Parts) > 0 {
		return resObj.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("empty response from Gemini")
}

func ExtractCodeBlock(response string) (string, string) {
	// 마크다운 코드 블록 패턴 매칭 ```lang ... ```
	startToken := "```"
	startIndex := strings.Index(response, startToken)
	if startIndex == -1 {
		return "", ""
	}

	// 언어 구분자 식별을 위한 파싱
	rem := response[startIndex+len(startToken):]
	lineBreak := strings.Index(rem, "\n")
	if lineBreak == -1 {
		return "", ""
	}

	lang := strings.TrimSpace(rem[:lineBreak])
	// lowercase 표준화 및 주석 문자 제거
	lang = strings.ToLower(lang)

	contentStart := startIndex + len(startToken) + lineBreak + 1
	endIndex := strings.Index(response[contentStart:], "```")
	if endIndex == -1 {
		return "", ""
	}

	code := strings.TrimSpace(response[contentStart : contentStart+endIndex])
	return lang, code
}

func ExtractPowerShellCode(response string) string {
	lang, code := ExtractCodeBlock(response)
	if lang == "powershell" || lang == "pwsh" || lang == "ps1" {
		return code
	}
	return ""
}

func ExtractPythonCode(response string) string {
	lang, code := ExtractCodeBlock(response)
	if lang == "python" || lang == "py" {
		return code
	}
	return ""
}

func ExtractGoCode(response string) string {
	lang, code := ExtractCodeBlock(response)
	if lang == "go" {
		return code
	}
	return ""
}

func SaveSession(history []Content) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile("session.json", data, 0644)
}

func LoadSession() ([]Content, error) {
	data, err := os.ReadFile("session.json")
	if err != nil {
		return nil, err
	}
	var history []Content
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, err
	}
	return history, nil
}

// ensurePsutil checks if psutil is installed and installs it if missing
func ensurePsutil() error {
	cmd := exec.Command("python", "-c", "import psutil")
	if err := cmd.Run(); err != nil {
		installCmd := exec.Command("python", "-m", "pip", "install", "psutil", "-q")
		return installCmd.Run()
	}
	return nil
}

// runSystemInfo executes the Python script and returns its output
func runSystemInfo() (string, error) {
	if err := ensurePsutil(); err != nil {
		return "", err
	}
	cmd := exec.Command("python", "system_info.py")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

// init checks for "info" argument and prints system info if present
func init() {
	if len(os.Args) > 1 && os.Args[1] == "info" {
		info, err := runSystemInfo()
		if err != nil {
			log.Fatalf("Failed to get system info: %v", err)
		}
		fmt.Print(info)
		os.Exit(0)
	}
}
