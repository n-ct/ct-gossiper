package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    ctd "GossipServer/CTData"
    "os"
    "bytes"
    //"strconv"
)

var Port string;
var Peers []string;

var messages map[string]ctd.CTData

func main() {
  if len(os.Args) < 2 {
    fmt.Println("use: gossipServer.exe <MY-PORT> <PEER-1-PORT> ... <PEER-n-PORT>"); //in case I forget how to run my program
    return;
  }
  Port := os.Args[1];
  for i:=2; i < len(os.Args); i++ {
    Peers = append(Peers, fmt.Sprintf("http://localhost:%v/ct/v1", os.Args[i])); //fill array of peers with adresses
    fmt.Println(Peers[i-2]); //for debug
  }

  messages = make(map[string]ctd.CTData);

  http.HandleFunc("/ct/v1/gossip", GossipHandler); // call GossipHandler on post to /gossip
  http.HandleFunc("/ct/v1/request", RequestHandler); // call RequestHandler on post to /request

  fmt.Printf("Starting server on %v\n", Port); // for debug

  http.ListenAndServe(fmt.Sprintf(":%v", Port), nil); // start server
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
  fmt.Printf("CTData: %+v %+v\n", data, messages); //print contents of message (for debugging)
  if message, ok := messages[data.Identifier()]; ok { // if I have the message already check for conflict
    if bytes.Compare(data.TBS.Digest, message.TBS.Digest)==0 {
      http.Error(w, "duplicate item\n", http.StatusBadRequest); // if no conflic send back "duplicate item", and bad request status code to sender
    } else {
      fmt.Printf("misbehavior detected"); // if conflict send a PoM to all peers
      PoM := ctd.NewDummyPoM();
      for i := 0; i < len(Peers); i++ {
        post(fmt.Sprintf("%v/gossip", Peers[i]), PoM);
      }
    }
  }else{
    fmt.Fprintf(w, "New data thank you! ->[%+v]\n", data); // if message is new send back message for debugging
    messages[data.Identifier()] = data; // if message is new add it to messages map
    //TODO add optimization where you set blob to nil unless requested?
    for i := 0; i < len(Peers); i++ {
      post(fmt.Sprintf("%v/gossip", Peers[i]), &data); //gossip new message to all peers
    }
  }
}

func RequestHandler(w http.ResponseWriter, req *http.Request){
  fmt.Fprintf(w, "test\n");
}

// post takes in an adress as a string and a pointer to a CTData struct
// and makes a post request to that address with the JSON encoded version of that struct
func post(address string, data *ctd.CTData){
  var jsonStr, _ = json.Marshal(data); //create JSON string from struct

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
  fmt.Println("response Status:", resp.Status);
  fmt.Println("response Headers:", resp.Header);
  body, _ := ioutil.ReadAll(resp.Body);
  fmt.Println("response Body:", string(body));
}
