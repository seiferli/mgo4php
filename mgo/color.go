package mgo

import (
	"fmt"
	"runtime"
	"strconv"
)

const (
	TextBlack = iota + 30
	TextRed
	TextGreen
	TextYellow
	TextBlue
	TextMagenta
	TextCyan
	TextWhite
)

func ConvertToString(input interface{}) string {
	var output string
	switch val := input.(type) {
	case int:
		output = strconv.Itoa(val)
	case string:
		output = val
	case byte:
		output = string(val)
	}
	return output
}

func ConsoleBlack(str string) string {
	return textColor(TextBlack, str)
}

func ConsoleRed(str string) string {
	return textColor(TextRed, str)
}

func ConsoleGreen(str string) string {
	return textColor(TextGreen, str)
}

func ConsoleYellow(str string) string {
	return textColor(TextYellow, str)
}

func ConsoleBlue(str string) string {
	return textColor(TextBlue, str)
}

func ConsoleMagenta(str string) string {
	return textColor(TextMagenta, str)
}

func ConsoleCyan(str string) string {
	return textColor(TextCyan, str)
}

func ConsoleWhite(str string) string {
	return textColor(TextWhite, str)
}

func textColor(color int, str string) string {
	if runtime.GOOS == "windows" {
		return str
	}
	switch color {
	case TextBlack:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextBlack, str)
	case TextRed:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextRed, str)
	case TextGreen:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextGreen, str)
	case TextYellow:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextYellow, str)
	case TextBlue:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextBlue, str)
	case TextMagenta:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextMagenta, str)
	case TextCyan:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextCyan, str)
	case TextWhite:
		return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", TextWhite, str)
	default:
		return str
	}
}
