package a

func panicTest() {
	panic("ошибка") // want "использование panic недопустимо"
}
