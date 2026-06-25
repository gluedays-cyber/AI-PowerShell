# Pi - AI-Powered System Command Execution via Python

Pi is an advanced Command Line Interface (CLI) assistant built in Go. When a user inputs natural language, Pi automatically generates and executes the corresponding Python code. Through this generated Python code, Pi can seamlessly execute PowerShell commands and interact with the host system. By leveraging the Gemini API, Pi creates an autonomous feedback loop that automatically fixes runtime exceptions, manages virtual environments, and installs required packages dynamically.

---

## Key Features

- **Natural Language Execution:** Simply input natural language prompts, and Pi automatically generates and executes the necessary Python code.
- **PowerShell Command Execution:** The generated Python scripts can directly execute Windows PowerShell commands, enabling deep system integration and automation.
- **Auto-Execution & Feedback Loop:** Pi writes code into temporary files (`py-cli-YYYYMMDDhhmmss.py`), runs them, and captures standard outputs, errors, and system exceptions. If the execution fails, the AI automatically analyzes the failure and attempts self-recovery.
- **Dynamic Dependency Management:** Checks the project's dependency state and automatically installs required packages using `pip` if a `requirements.txt` file is present.
- **Interactive REPL Commands:**
  - `stats`: View detailed metrics on script executions (success/failure rates).
  - `coverage`: Diagnose test automation rate and coverage details.
  - `depgraph`: Visualize imports and module dependency structures.
  - `save` / `load`: Save and load chat sessions locally (`session.json`).

---

## Technical Specifications

- **Language:** Go 1.20+
- **Primary AI Model:** Google Gemini API (`gemini-2.5-flash-lite`)
- **Config Storage:** Decrypts system parameters dynamically from APPDATA (`.tea/.apikeys.json`) with AES/DPAPI fallback.

---

## Dependency on `tea` Repository

> [!IMPORTANT]
> **Pi does not feature native API key or model configuration management.** Instead, it is designed to share settings with the [tea](https://github.com/sezzi/tea) (or equivalent parent GUI) repository.
>
> To use Pi, you must first install and configure the **`tea`** application, which saves and encrypts the Gemini API keys and preferred target models inside the local profile path (`%APPDATA%/.tea/.apikeys.json`). Pi will automatically decrypt and load these configurations at runtime.

---

## Getting Started

### 1. Requirements

- **Go:** Version 1.20 or higher installed.
- **Python:** Version 3.8 or higher installed.
- **Dependency Setup:** The [tea](https://github.com/sezzi/tea) repository must be configured on the host system to supply decrypted credentials. Alternatively, a local environment variable fallback (`GEMINI_API_KEY`) is supported.

### 2. Configuration Setup

Provide your API key either via setting the environment variable or entering it on prompt:

```bash
# Set environment variable
export GEMINI_API_KEY="your-gemini-api-key-here"  # macOS/Linux
$env:GEMINI_API_KEY="your-gemini-api-key-here"   # Windows PowerShell
```

### 3. Building from Source

Build the Go executable directly in the root folder:

```bash
go build -o pi.exe
```

---

## Usage

You can run Pi with a one-off request directly from the terminal or enter interactive REPL mode by invoking the executable without parameters.

### Command Line Invocation (One-off prompt)

```bash
./pi.exe "Write a script that pulls the current weather of Seoul using public APIs"
```

### Interactive REPL Mode

```bash
./pi.exe
```

This enters the interactive shell prompt, which dynamically displays your current working directory:

```text
C:\Users\username\project> 
```

---

## Deployment & Distribution

### Building Cross-Platform Executables

You can build the binary for different target platforms using Go's build variables:

**Windows:**

```bash
GOOS=windows GOARCH=amd64 go build -o dist/pi-windows-amd64.exe
```

**Linux:**

```bash
GOOS=linux GOARCH=amd64 go build -o dist/pi-linux-amd64
```

**macOS:**

```bash
GOOS=darwin GOARCH=amd64 go build -o dist/pi-darwin-amd64
```

### CI/CD Deployment Flow

1. **GitHub Actions:** Create a release workflow triggered by version tags (e.g., `v*`).
2. **Matrix Build:** Compile for Windows, Linux, and macOS targets.
3. **Artifact Upload:** Package targets as zip/tar.gz and attach them to a GitHub Release.

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
