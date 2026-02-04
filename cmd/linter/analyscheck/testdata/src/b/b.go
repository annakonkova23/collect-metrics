package b

import "log"

func logFatalTest() {
	if true {
		log.Fatal("test") // want "использование log.Fatal вне main недопустимо"
	}
}
