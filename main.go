package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"SR-Player/internal/ui"
)

func printLogo() {
	content, err := os.ReadFile("assets/am-ascii.txt")
	if err != nil {
		// Si no existe el archivo, ignoramos la intro gráfica
		return
	}
	lines := strings.Split(string(content), "\n")

	// Limpiar pantalla al inicio antes de imprimir el arte
	fmt.Print("\033[H\033[2J")

	for _, line := range lines {
		// Imprime cada línea coloreada en rojo intenso (\033[31m)
		fmt.Printf("\033[31m%s\033[0m\n", line)
		// Simula el retraso exacto que usabas con perl (sleep 0.008)
		time.Sleep(8 * time.Millisecond)
	}
	fmt.Println()
}

func checkAndPrompt() bool {
	cmd := exec.Command("pgrep", "-x", "Music")
	if err := cmd.Run(); err == nil {
		// Apple Music ya está corriendo. Una pausa de cortesía para ver el logo completo y arrancamos.
		time.Sleep(500 * time.Millisecond)
		return true
	}

	// Apple Music está cerrado
	fmt.Print("\033[1;37mApple Music no está en ejecución.\033[0m\n")
	fmt.Print("\033[1;32m[1]\033[0m Prender Apple Music e iniciar\n")
	fmt.Print("\033[1;31m[2]\033[0m Salir del TUI\n")
	fmt.Print("\n\033[1;36m>\033[0m Elige una opción: ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "1" {
			fmt.Print("\n\033[33mIniciando Apple Music en segundo plano...\033[0m\n")
			exec.Command("open", "-a", "Music").Run()

			// Esperar a que levante completamente y registre AppleEvents
			for i := 0; i < 5; i++ {
				time.Sleep(1 * time.Second)
				if exec.Command("pgrep", "-x", "Music").Run() == nil {
					// Buffer de seguridad para que Music cargue su librería antes de las peticiones
					time.Sleep(1 * time.Second) 
					break
				}
			}
			return true
		} else if text == "2" || strings.ToLower(text) == "q" {
			fmt.Print("\n\033[31mSaliendo del TUI...\033[0m\n")
			return false
		}
		fmt.Print("\033[31mOpción inválida.\033[0m Elige [1] o [2]: ")
	}
	return false
}

func main() {
	// 1. Mostrar la intro de arte ASCII
	printLogo()

	// 2. Verificar estado de Apple Music y preguntar si es necesario
	if !checkAndPrompt() {
		os.Exit(0)
	}

	// 3. Iniciar el TUI
	// Use alternate screen and mouse support
	p := tea.NewProgram(
		ui.NewApp(0, 0),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running SR-Player: %v\n", err)
		os.Exit(1)
	}
}
