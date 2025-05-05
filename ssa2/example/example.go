package main

import "errors"

func f() error {
	err := errors.New("something")
	if err != nil {
		return nil  // ここを検出する
	}
	return nil
}