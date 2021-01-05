package main

import (
	"time"
	"fmt"
	"net/http"
	"encoding/json"
	ctd "GossipServer/CTData"
	"os"
	"os/signal"
	"syscall"
	"bytes"
	"io/ioutil"
	"strings"
	"github.com/golang/glog"
	"flag"
)

var peers []string;

var messages map[string]ctd.CTData

func main() {

	done := make(chan os.Signal, 1); //create a channel to signify when server is shut down with ctrl+c
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM);//notify the channel when program terminated

	go func() {
		<-done
		glog.Infoln("kill recived");
		glog.Flush();
		os.Exit(1);
	}(); //when channel is notified print debug info, make sure all logs get written and exit

	//Setting flags
	var peersNumber = flag.Int("n", 0, "Number of gossipers in the experiment. Must be more than 0");
  var ip = flag.Int("ip", 0, "Local ip of gossiper server. Must be more than 1.");
  var port = flag.String("port", "2000", "Server port.");

	flag.Parse(); //read the cmd line flags
	defer glog.Flush(); //if the program ends unexpectedly make sure all debug info is printed

	if(*peersNumber < 1 || *ip < 2){
    fmt.Println("ip and n flags are require.");
    return;
  }

	for i:=2; i <= *peersNumber+1; i++ {
		if(i == *ip){
			continue;
		}
		peers = append(peers, fmt.Sprintf("http://10.1.1.%d:%v/ct/v1", i, *port)); //fill array of peers with adresses
		glog.Infoln(peers[len(peers)-1]); //for debug
	}

	messages = make(map[string]ctd.CTData);

	http.HandleFunc("/ct/v1/gossip", GossipHandler); // call GossipHandler on post to /gossip

	glog.Infof("Starting server on %v\n", *port); // for debug

	err := http.ListenAndServe(fmt.Sprintf(":%v", *port), nil); // start server
	if err != nil {
		fmt.Errorf("err: %v", err);
	}
}

// GossipHandler is called on a post request to /ct/v1/gossip.
// It handles the logic of gossip within a network system
func GossipHandler(w http.ResponseWriter, req *http.Request){
	data := ctd.CTData{}; // create an empty CTData struct
	err := json.NewDecoder(req.Body).Decode(&data); // fill that struct using the JSON encoded struct send via the post
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // if there is an eror report and abort
		return;
	}
	// TODO figure out logger for go? i think Jeremy used glog?
	glog.Infof("CTData recived from source: %s\n\n", data.ToDebugString()); //print contents of message (for debugging)
	if message, ok := messages[data.Identifier()]; ok { // if I have the message already check for conflict
		if bytes.Compare(data.TBS.Digest, message.TBS.Digest)==0 {
			http.Error(w, "duplicate item\n", http.StatusBadRequest); // if no conflic send back "duplicate item", and bad request status code to sender
		} else {
			glog.Infof("misbehavior detected\n\n"); // if conflict send a PoM to all peers
			PoM := ctd.NewCTData("PoM", time.Now().Unix(), []byte{0,1,2,3}); //new dummy Proof of misbehavior
			messages[PoM.Identifier()] = *PoM; // store PoM
			for i := 0; i < len(peers); i++ {
				glog.Infof("gossiping PoM to peer: %v\n", peers[i]);
				post(fmt.Sprintf("%v/gossip", peers[i]), PoM, false); //gossip new PoM to all peers (false indicates send without blob)
			}
		}
	}else{
		if data.TBS.Blob == nil{ //If the message does not contain the blob
			fmt.Fprintf(w, "blob-request"); //respond with "blob-request"
		} else {
			fmt.Fprintf(w, "new data"); //respond with "new data"
			messages[data.Identifier()] = data; // if message is new add it to messages map

			for i := 0; i < len(peers); i++ {
				glog.Infof("gossiping new info to peer: %v\n", peers[i]);
				post(fmt.Sprintf("%v/gossip", peers[i]), &data, false); //gossip new message to all peers (false indicates send without blob)
			}
		}
	}
}

// post takes in an address as a string and a pointer to a CTData struct
// and makes a post request to that address with the JSON encoded version of that struct
func post(address string, data *ctd.CTData, withBlob bool){
	var toSend *ctd.CTData;
	if withBlob {
		toSend = data;
	} else {
		toSend = data.CopyWithoutBlob();
	}
	var jsonStr, _ = json.Marshal(toSend); //create JSON string from struct w/o the blob

	req, err := http.NewRequest("POST", address, bytes.NewBuffer(jsonStr)); //create a post request
	req.Header.Set("X-Custom-Header", "myvalue"); //not sure if this is needed but the tutorial I copied from had it
	req.Header.Set("Content-Type", "application/json"); //set message type to JSON

	client := &http.Client{};
	resp, err := client.Do(req); //make the request
	if err != nil {
		panic(err);
	}

	defer resp.Body.Close();

	//print info for debug
	glog.Infoln("response Status:", resp.Status);
	glog.Infoln("response Headers:", resp.Header);
	body, _ := ioutil.ReadAll(resp.Body);
	sbody := string(body);
	glog.Infoln("response Body:", sbody);

	if strings.ToLower(sbody) == "blob-request" {
		glog.Infof("sending blob to peer: %v\n\n", address);
		post(address, data, true); // if the recipient sends back a blob request resend the message with the blob
	}
}
