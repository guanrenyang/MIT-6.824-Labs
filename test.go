package main

import "os"
import "fmt"
import "encoding/json"
type KV struct{
	k string
	v string
}
func main(){
	temp := make(map[string][]int)
	temp["a"] = []int{1,2,3}
	temp["b"] = []int{4,5,6}

	file, _ := os.Create("temp.json")
	defer file.Close()
	enc := json.NewEncoder(file)
	fmt.Println(temp)
	enc.Encode(&temp)
}
