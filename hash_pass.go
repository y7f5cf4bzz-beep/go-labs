package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := []byte("parol") // замените на свой пароль

	hashed, err := bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Хеш пароля '%s': %s\n", password, string(hashed))

	// Проверка с правильным паролем
	err = bcrypt.CompareHashAndPassword(hashed, []byte("parol"))
	if err == nil {
		fmt.Println("✅ Пароль совпадает с хешем")
	} else {
		fmt.Println("❌ Пароль НЕ совпадает")
	}

	// Проверка с неправильным паролем
	err = bcrypt.CompareHashAndPassword(hashed, []byte("wrong"))
	if err != nil {
		fmt.Println("❌ Ожидаемо: неправильный пароль не подходит")
	}
}
