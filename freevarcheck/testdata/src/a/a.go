package a

func f() {
	var n int  // want "自由関数への書き込みです"
	func() {
		n = 10
	}()
	println(n)
}