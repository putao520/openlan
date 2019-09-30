package main 

import "fmt"

func main() {
    var a = []int{1,2,3} 

    fmt.Println(a)
    a0 := append(a, []int{4,5,6}...)
    a0[0] = 9
    a1 := append(a, []int{7,8}...)
    fmt.Println(a0)
    fmt.Println(a1)
 
    fmt.Println(a)
    a0 = append(a[:3], []int{4,5,6}...)
    a0[0] = 9
    a1 = append(a[:3], []int{7,8}...)
    fmt.Println(a0)
    fmt.Println(a1)

    a = make([]int, 0, 1024)
 
    b := append(a, []int{4,5,6}...)
    fmt.Println(b, a)
    fmt.Println(cap(b), len(b))
    fmt.Println(cap(a), len(a))
    
    // a = make([]int, 1024)
    b = append(a, []int{8,9}...)
    fmt.Println(b, a)
    fmt.Println(cap(b), len(b))
    fmt.Println(cap(a), len(a))
}
