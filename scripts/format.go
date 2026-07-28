package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	// Wails roda hooks na pasta build/bin/
	// Navegando iterativamente até achar o go.mod na raiz:
	for {
		if _, err := os.Stat("go.mod"); err == nil {
			break
		}
		if err := os.Chdir(".."); err != nil {
			break
		}
	}

	fmt.Println("[Format] Formatando o backend (Go)...")
	cmd1 := exec.Command("go", "fmt", "./...")
	cmd1.Stdout = os.Stdout
	cmd1.Stderr = os.Stderr
	if err := cmd1.Run(); err != nil {
		fmt.Println("Erro ao formatar Go:", err)
	}

	fmt.Println("[Format] Formatando o frontend (React)...")
	var cmd2 *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd2 = exec.Command("cmd.exe", "/c", "npm run format")
	} else {
		cmd2 = exec.Command("npm", "run", "format")
	}

	cmd2.Dir = "frontend"
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr
	if err := cmd2.Run(); err != nil {
		fmt.Println("Erro ao formatar Frontend:", err)
	}

	fmt.Println("[Format] Tudo formatado com sucesso!")
}
