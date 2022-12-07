package testdata

func TestRoutine() {
	func1()
}

func func1() {
	go func2()
	go func() {
		func2()
	}()
}

func func2() {
	
}
