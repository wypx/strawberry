package main

import (
	"fmt"
	"reflect"
	"runtime"
	"unicode/utf8"
)

func TestFn(a int, b string) (string, int) {
	return b, a
}

func TestCB(fn func(int, string) (string, int), a int, b string) (string, int) {
	fmt.Println(runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name())
	return fn(a, b)
}

// 切片，而不是数组
// arr [5]int)和arr [10]int)不同类型
func printArr(arr [5]int) {
	arr[0] = 100
	for i, v := range arr {
		fmt.Println(i, v)
	}
}

func printArrP(arr *[5]int) {
	arr[0] = 100
	for i, v := range arr {
		fmt.Println(i, v)
	}
}

func updateArr(arr []int) {
	arr[0] = 100
	for i, v := range arr {
		fmt.Println(i, v)
	}
}

type Volume struct {
	a int
	b string
}

func (v Volume) print() {
	fmt.Println(v)
}

func (v *Volume) set(a int, b string) {
	v.a = a
	v.b = b
}

func main() {
	var a = 3
	var p *int = &a
	*p = 4
	fmt.Println(a, *p)

	var arr0 [5]int
	arr1 := [3]int{1, 3, 5}
	arr2 := [...]int{1, 2, 3, 4, 5, 6, 7}
	var grid [2][3]bool
	fmt.Println(arr0, arr1, arr2, grid)

	for i := 0; i < len(arr1); i++ {
		fmt.Println(arr1[i])
	}

	for i := range arr2 {
		fmt.Println(arr2[i])
	}

	for i, v := range arr2 {
		fmt.Println(i, v)
	}

	printArr(arr0)
	fmt.Println(arr0)

	printArrP(&arr0)
	fmt.Println(arr0)

	fmt.Println("===================")

	// slice是对底层结构的view
	s := arr2[:6]
	fmt.Println(s)
	fmt.Println("===================")

	s = arr2[2:]
	fmt.Println(s)
	fmt.Println("===================")

	s = arr2[2:6]
	fmt.Println(s)
	updateArr(s)
	fmt.Println(s)

	s = arr2[:]
	fmt.Println(s)
	updateArr(s)
	fmt.Println(s)
	fmt.Println("===================")

	s = arr2[2:6]
	fmt.Println(s)
	fmt.Printf("len: %d cap: %d\n", len(s), cap(s))

	s = s[2:5]
	fmt.Println(s)

	s1 := append(s, 9)
	s2 := append(s1, 10)

	fmt.Println(s)
	fmt.Println(s1)
	fmt.Println(s2)

	fmt.Println(arr2)

	// len =  6
	s3 := make([]int, 6)
	fmt.Println(s3)
	//  len = 6, cap = 7
	s4 := make([]int, 6, 7)
	fmt.Println(s4)

	copy(s3, s2)
	fmt.Println(s3)

	// delete
	s3 = append(s3[:4], s3[5:]...)
	fmt.Println(s3)

	// m1 = make(map[string]string)
	// m1["test"] = "123"

	m2 := map[string]string{
		"test": "123",
		"jjjj": "456",
	}

	fmt.Println(m2, len(m2))

	for k, v := range m2 {
		fmt.Println(k, v)
	}

	// get value
	fmt.Println(m2["jjjj"])
	fmt.Println(m2["jjj"])

	v, ok := m2["test"]
	fmt.Println(v, ok)

	// delete key
	delete(m2, "jjjj")
	for k, v := range m2 {
		fmt.Println(k, v)
	}

	// 遍历字符串
	str := "梦杰, i love you"
	for _, b := range []byte(str) {
		fmt.Printf("%X\n", b)
	}

	// ch is a rune
	for i, ch := range str {
		fmt.Printf("(%d, %X)\n", i, ch)
	}
	fmt.Printf("rune count: %d\n", utf8.RuneCountInString(str))

	bytes := []byte(str)
	for len(bytes) > 0 {
		ch, size := utf8.DecodeRune(bytes)
		bytes = bytes[size:]
		fmt.Printf("%c", ch)
	}
	fmt.Println("")

	for i, ch := range []rune(str) {
		fmt.Printf("%d %c\n", i, ch)
	}
	fmt.Println("")

	TestCB(TestFn, 3, "hello")

	var vol Volume
	vol.a = 1
	vol.b = "love"
	fmt.Println(vol)

	vol.print()
	vol.set(4, "love you")
	vol.print()

}
