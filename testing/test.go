package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"encoding/json"
	"os"
	"context"

	mtr "github.com/n-ct/ct-monitor"
	entitylist "github.com/n-ct/ct-monitor/entitylist"
)

const(
	logListName = "config/log_list.json"
	logID = "sh4FzIuizYogTodm+Su5iiUgZ2va+nDnsklTLe+LkF4="
)

func main(){


	if len(os.Args) != 2 {
		fmt.Println("use: test <PORT>"); //in case I forget how to run my program
		return;
	}

	var jsonStr []byte;
	logList := entitylist.NewLogList(logListName)
	log := logList.FindLogByLogID(logID)
	logClient, err := mtr.NewLogClient(log)
	ctx := context.Background()
	sth, err := logClient.GetSTH(ctx)
	if err != nil{
		panic(err)
	}


	jsonStr, _ = json.Marshal(sth); //create a JSON string from CTObject struct

	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%v/ct/v1/gossip", os.Args[1]), bytes.NewBuffer(jsonStr)); //create a post request
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
