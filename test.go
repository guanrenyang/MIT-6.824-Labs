package main

import "os"
import "fmt"
import "encoding/json"
type KV struct{
	k string 
	v string
}
func main(){
	kv := KV{"1", "2"}
	file, _ := os.Create("temp.json")
	defer file.Close()
	enc := json.NewEncoder(file)
	fmt.Println(kv)
	enc.Encode(&kv)
} 