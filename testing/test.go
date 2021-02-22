package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	cto "ct-gossiper"
	"os"
	"strconv"
	//"time"
)

func main(){


	if len(os.Args) != 4 {
		fmt.Println("use: test <PORT> <TIMESTAMP> <BLOB>"); //in case I forget how to run my program
		return;
	}

	var jsonStr []byte;

	timestamp, _ := strconv.ParseUint(os.Args[2], 10, 64);
	blob := []byte(os.Args[3]);

	jsonStr, _ = json.Marshal(*cto.NewCTObject("Fake", timestamp, blob, "testSub")); //create a JSON string from CTData struct

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%v/ct/v1/gossip", os.Args[1]), bytes.NewBuffer(jsonStr)); //create a post request
	req.Header.Set("X-Custom-Header", "myvalue"); //not sure if this is needed but the tutorial I copied from had it
	req.Header.Set("Content-Type", "application/json"); //set message type to JSON

	client := &http.Client{};
	resp, err := client.Do(req); //make the request
	if err != nil {
		panic(err);
	}

	defer resp.Body.Close();

	//print info for debug
	fmt.Println("response Status:", resp.Status);
	fmt.Println("response Headers:", resp.Header);
	body, _ := ioutil.ReadAll(resp.Body);
	fmt.Println("response Body:", string(body));
}
