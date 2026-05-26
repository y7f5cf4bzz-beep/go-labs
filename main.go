package main

import "fmt"

func showSimpleTypes() {
    var age int = 25
    var name string = "Alice"
    var isStudent bool = true
    fmt.Println("=== Простые типы данных ===")
    fmt.Println("int:", age)
    fmt.Println("string:", name)
    fmt.Println("bool:", isStudent)
    fmt.Println()
}

func arrayAndSlices() {
    arr := [5]int{10, 20, 30, 40, 50}
    fmt.Println("=== Массив и слайсы ===")
    fmt.Println("Исходный массив:", arr)
    slice1 := arr[1:4]
    fmt.Println("Слайс slice1 (arr[1:4]):", slice1)
    slice2 := make([]int, 3, 5)
    slice2[0] = 100
    slice2[1] = 200
    slice2[2] = 300
    fmt.Println("Слайс slice2 (make):", slice2)
    slice2 = append(slice2, 400, 500)
    fmt.Println("После append:", slice2)
    slice1[0] = 999
    fmt.Println("После изменения slice1[0]=999:")
    fmt.Println("  slice1:", slice1)
    fmt.Println("  исходный массив arr:", arr)
    fmt.Println()
}

type Person struct {
    FirstName string
    LastName  string
    Age       int
    City      string
}

func showStruct() {
    p := Person{
        FirstName: "Иван",
        LastName:  "Петров",
        Age:       30,
        City:      "Москва",
    }
    fmt.Println("=== Структура Person ===")
    fmt.Printf("Структура: %+v\n", p)
    fmt.Println("Имя:   ", p.FirstName)
    fmt.Println("Фамилия:", p.LastName)
    fmt.Println("Возраст:", p.Age)
    fmt.Println("Город: ", p.City)
    fmt.Println()
}

type Describer interface {
    describe() string
}

func (p Person) describe() string {
    return fmt.Sprintf("%s %s, %d лет, живёт в %s", p.FirstName, p.LastName, p.Age, p.City)
}

type Product struct {
    Name  string
    Price float64
}

func (pr Product) describe() string {
    return fmt.Sprintf("Товар: %s, цена: %.2f руб.", pr.Name, pr.Price)
}

func printDescription(d Describer) {
    fmt.Println(d.describe())
}

func showInterface() {
    fmt.Println("=== Интерфейс Describer ===")
    person := Person{FirstName: "Мария", LastName: "Сидорова", Age: 25, City: "Санкт-Петербург"}
    product := Product{Name: "Ноутбук", Price: 59999.99}
    fmt.Println("Вызов printDescription для Person:")
    printDescription(person)
    fmt.Println("Вызов printDescription для Product:")
    printDescription(product)
}

func main() {
    showSimpleTypes()
    arrayAndSlices()
    showStruct()
    showInterface()
}
