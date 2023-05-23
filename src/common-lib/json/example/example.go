package main

import (
	"bufio"
	"bytes"
	"fmt"

	"gitlab.kksharmadevdev.com/platform/platform-common-lib/src/v6/json"
)

type Object struct {
	Value1 string `json:"value-1"`
	Value2 int    `json:"value-2"`
}

func main() {
	SerializeObjectToFile()
	SerializeObjectToString()
	//typically usable in case of http.ResponseWriter
	SerializeObjectToWriter()

	DeserializeObjectFromFile()
	DeserializeObjectFromString()
}

func SerializeObjectToFile() {
	jsonSerializer := json.FactoryJSONImpl{}.GetSerializerJSON()
	jsonSerializer.WriteFile("./dest.json", Object{
		Value1: "test",
		Value2: 2,
	})
}

func SerializeObjectToString() {
	jsonSerializer := json.FactoryJSONImpl{}.GetSerializerJSON()
	bytes, _ := jsonSerializer.WriteByteStream(Object{
		Value1: "test",
		Value2: 2,
	})
	fmt.Printf("%v", string(bytes))
}

func SerializeObjectToWriter() {
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	jsonSerializer := json.FactoryJSONImpl{}.GetSerializerJSON()
	jsonSerializer.Write(w, Object{
		Value1: "test",
		Value2: 3,
	})
	w.Flush()
	fmt.Printf("%v", b.String())
}

func DeserializeObjectFromFile() {
	o := Object{}
	jsonDeserializer := json.FactoryJSONImpl{}.GetDeserializerJSON()
	jsonDeserializer.ReadFile(&o, "./src.json")
	fmt.Printf("%v ", o.Value1)
	fmt.Printf("%v ", o.Value2)
}

func DeserializeObjectFromString() {
	o := Object{}
	jsonDeserializer := json.FactoryJSONImpl{}.GetDeserializerJSON()
	jsonDeserializer.ReadString(&o, "{\"value-1\": \"value\",\"value-2\": 20}")
	fmt.Printf("\n%v ", o.Value1)
	fmt.Printf("%v ", o.Value2)
}
