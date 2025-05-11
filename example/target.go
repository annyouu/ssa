package example

func doubleClose() {
	ch := make(chan int)
	close(ch)

	close(ch)
}

func safeClose() {
	ch := make(chan int)
	close(ch)
}