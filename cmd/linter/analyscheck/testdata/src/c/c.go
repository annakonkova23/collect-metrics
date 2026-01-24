package c

import "os"

func osExitTest() {
	if true {
		os.Exit(1) // want "использование os.Exit вне main недопустимо"
	}
}
