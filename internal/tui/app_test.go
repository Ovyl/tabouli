package tui

import (
	"fmt"
	"testing"
)

func TestGetRune(t *testing.T) {
	fmt.Println("Say hi")
	num := 0
	t.Log("Running test")
	for num < 100 {
		var runeVal = getShortcutRuneForCount(num)
		fmt.Printf("Num: %v Rune: %s\n", num, string(runeVal))
		num += 1

	}
}
