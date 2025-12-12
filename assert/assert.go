package assert

func IsTrue(given bool) {
	if given == false {
		panic("value is not true")
	}
}

func IsFalse(given bool) {
	if given != false {
		panic("value is not false")
	}
}

func Is(given bool, expected bool) {
	if given != expected {
		panic("values do not match")
	}
}
