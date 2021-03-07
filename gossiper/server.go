package main

import (
	"fmt"
	"net"
	"net/http"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"bytes"
	"io/ioutil"
	"strings"
	"flag"
	"time"

	//"github.com/golang/glog"
	cto "github.com/n-ct/ct-gossiper"
	mtr "github.com/n-ct/ct-monitor"
	mtrList "github.com/n-ct/ct-monitor/entitylist"
	signature "github.com/n-ct/ct-monitor/signature"
	mtrUtils "github.com/n-ct/ct-monitor/utils"
)


var peers []*mtrList.MonitorInfo;
var messages cto.MessagesMap; //[TypeID][subjectOrSigner][Timestamp][Version]
var alertsMap cto.MessagesMap;//[Subject][Signer][Timestamp][Version]
var port string;
var myAddress string;
var allMonitors *mtrList.MonitorList;
var gossipConfig *cto.GossipConfig;
var allLogs *mtrList.LogList;


func main() {

	done := make(chan os.Signal, 1); //create a channel to signify when server is shut down with ctrl+c
  signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM);//notify the channel when program terminated

	go func() {
		<-done
		fmt.Println("kill recived");
		os.Exit(1);
	}(); //when channel is notified print debug info, make sure all logs get written and exit

	//Setting flags
  var configFilename = flag.String("config", "", "File containing gossiper configuration");
	var monitorsFilename = flag.String("monitor_list", "", "File containing monitor-gossiper pairs");
	var logsFilename = flag.String("log_list", "", "File containing the list of logs");

	flag.Parse();

	//if filenames are not defined, terminate
	if len(*configFilename) == 0 || len(*monitorsFilename) == 0 || len(*logsFilename) == 0 {
    fmt.Println("configuration files are required.")
    return
  }

	GossiperSetup(*configFilename, *monitorsFilename, *logsFilename);

	http.HandleFunc(cto.GossipPath, GossipHandler); // call GossipHandler on Post to /gossip

	fmt.Printf("Starting server on %v\n", port);

	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); // start server
	if err != nil {
		fmt.Errorf("err: %v", err);
	}
}

// GossipHandler is called on a Post request to /ct/v1/gossip.
// It handles the logic of gossip within a network system
func GossipHandler(w http.ResponseWriter, req *http.Request){
	data := mtr.CTObject{};
	err := json.NewDecoder(req.Body).Decode(&data); // fill that struct using the JSON encoded struct send via the Post
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // if there is an eror report and abort
		return;
	}
	requesterAddress := req.Header.Get("requesterAddress")

	//Get data identifier and select map to use
	identifier := data.Identifier();
	identifierStr := cto.IdentifierToString(identifier)

	var workingMap cto.MessagesMap;
	if data.TypeID == "Alert" {
		workingMap = alertsMap;
	} else {
		workingMap = messages;
	}

	fmt.Printf("%s Received request\n", identifierStr)
	if message, ok := workingMap[identifier.First][identifier.Second][identifier.Third][identifier.Fourth]; ok { // if I have the message already check for conflict
		if bytes.Compare(data.Digest, message.Digest)==0 {
			fmt.Printf("%s Duplicate Item\n\n", identifierStr)
			http.Error(w, "Duplicate item\n", http.StatusBadRequest); // if no conflic send back "duplicate item", and bad request status code to sender

		} else {
			if data.Blob == nil{ //If the message does not contain the blob
				fmt.Printf("%s blob-request sent\n", identifierStr)
				fmt.Fprintf(w, "blob-request"); //respond with "blob-request"
			} else {
				if ValidateSignature(&data) {
					fmt.Printf("%s Misbehavior detected\n", identifierStr); // if conflict send a PoM to all peers.
					PoM, err := mtr.CreateConflictingSTHPOM(&data, message)
					if err == nil {
						addEntry(messages, *PoM, PoM.Identifier()); // store PoM
						fmt.Printf("%s Stored PoM\n", identifierStr)
						gossipPeers(PoM, requesterAddress)
						gossipMonitor(PoM, requesterAddress)
						fmt.Printf("%s Finished gossiping PoM\n\n", identifierStr)

					} else {
						fmt.Errorf("Error creating ConflictingSTHPOM: %s\n", err) //error creating "ConflictingSTHPOM"
					}
				}
			}
		}
		} else { //message not in MessagesMap
			if data.Blob == nil{ //If the message does not contain the blob
				fmt.Printf("%s blob-request sent\n", identifierStr)
				fmt.Fprintf(w, "blob-request"); //respond with "blob-request"

		} else {
			if ValidateSignature(&data) {
				fmt.Fprintf(w, "new data"); //respond with "new data"
				addEntry(workingMap, data, identifier);// if message is new add it to messages map
				fmt.Printf("%s Stored new data\n", identifierStr)
				gossipPeers(&data, requesterAddress)
				gossipMonitor(&data, requesterAddress)
				fmt.Printf("%s Finished gossiping new data\n\n", identifierStr)

			} else {
				//invalid Signature
				fmt.Printf("%s Invalid Signature\n\n", identifierStr)
				http.Error(w, "Invalid Signature\n", http.StatusBadRequest);
			}
		}
	}
}

// Post takes in an address as a string and a pointer to a CTObject struct
// and makes a Post request to that address with the JSON encoded version of that struct
func Post(address string, data *mtr.CTObject, withoutBlob bool){
	var toSend *mtr.CTObject;
	if withoutBlob && len(data.Blob) > cto.Threshold { //threshold  to be change after testing
		toSend = cto.CopyWithoutBlob(data);
	} else {
		toSend = data;
	}
	var jsonStr, _ = json.Marshal(toSend);

	req, err := http.NewRequest("POST", address, bytes.NewBuffer(jsonStr)); //create a Post request
	req.Header.Set("X-Custom-Header", "myvalue");
	req.Header.Set("Content-Type", "application/json"); //set message type to JSON
	req.Header.Add("requesterAddress", myAddress);

	client := &http.Client{};
	resp, err := client.Do(req); //make the request
	if err != nil {
		fmt.Errorf("Unable to make request: %s\n", err)
		return
	}

	defer resp.Body.Close();

	//print info for debug
	fmt.Printf("response Status: %s\n", resp.Status);
	fmt.Printf("response Headers: %s\n", resp.Header);
	body, _ := ioutil.ReadAll(resp.Body);
	sbody := string(body);
	fmt.Printf("response Body: %s\n", sbody);

	if strings.ToLower(sbody) == "blob-request" {
		fmt.Printf("Sending blob to peer: %v\n", address);
		Post(address, data, false); // if the recipient sends back a blob request resend the message with the blob
	}
}

//addEntry adds a new entry to the selected map using the data identifier as keys
func addEntry(dataMap cto.MessagesMap, data mtr.CTObject, identifier mtr.ObjectIdentifier){
	if _, ok := dataMap[identifier.First]; !ok {
		dataMap[identifier.First] = make(map[string]map[uint64]map[string] *mtr.CTObject);
	}
	if _, ok := dataMap[identifier.First][identifier.Second]; !ok {
		dataMap[identifier.First][identifier.Second] = make(map[uint64]map[string] *mtr.CTObject);
	}
	if _, ok := dataMap[identifier.First][identifier.Second][identifier.Third]; !ok {
		dataMap[identifier.First][identifier.Second][identifier.Third] = make(map[string] *mtr.CTObject);
	}
	dataMap[identifier.First][identifier.Second][identifier.Third][identifier.Fourth] = &data;
}

//GossiperSetup configures gossiper varialbes from json files
func GossiperSetup(configFilename string, monitorsFilename string, logsFilename string){
	//create message maps
	messages = make(cto.MessagesMap);
	alertsMap = make(cto.MessagesMap);
	//get gossiper configuration
	gossipConfig = cto.NewGossipConfig(configFilename);

	GetPeers(gossipConfig, monitorsFilename);
	myAddress = allMonitors.FindMonitorByMonitorID(gossipConfig.Monitor_id).GossiperURL;
	port = strings.Split(myAddress, ":")[2];

	var err error
	allLogs, err = mtrList.NewLogList(logsFilename); //get all logs
	if err != nil{
		panic(err)
	}
	fmt.Println("Setup completed")
}

//GetPeers get all monitors from file and populates the peers with the respective monitors
func GetPeers(gossipConfig *cto.GossipConfig, monitorsFilename string)  {
	var err error
	allMonitors, err = mtrList.NewMonitorList(monitorsFilename)
	if err != nil{
		panic(err)
	}

	for _, monitorId := range gossipConfig.Monitors_ids {
		peers = append(peers, allMonitors.FindMonitorByMonitorID(monitorId));
	}
}

//gossipPeers sends new data to other gossip servers
func gossipPeers(data *mtr.CTObject, requesterAddress string){
	for _, peer := range peers{

		if requesterAddress != peer.GossiperURL{
			fmt.Printf("Gossiping info to peer: %v\n", peer.MonitorID);
			Post(mtrUtils.CreateRequestURL(peer.GossiperURL, cto.GossipPath), data, true);
		}
	}
}

//gossipMonitor sends new data to the monitor
func gossipMonitor(data *mtr.CTObject, requesterAddress string){
	monitorUrl := allMonitors.FindMonitorByMonitorID(gossipConfig.Monitor_id).MonitorURL;
	if requesterAddress == monitorUrl {
		return
	}
	//Check if monitor is reachable
	timeout := 1 * time.Second
	_, err := net.DialTimeout("tcp", monitorUrl, timeout)
	if err != nil {
		fmt.Println("Monitor unreachable.")
	} else {
		Post(mtrUtils.CreateRequestURL(monitorUrl, mtr.NewInfoPath), data, false)
	}
}

//ValidateSignature check if the received message has a valid signature
func ValidateSignature(data *mtr.CTObject) bool {
	var signature_err error

	switch data.TypeID{
	case mtr.STHTypeID:
		logger := allLogs.FindLogByLogID(data.Signer)
		blob, err := data.DeconstructSTH()
		if err != nil{
			fmt.Errorf("Error deconstructing STH: %s\n", err)
			return false
		}
		signature_err = signature.VerifySignature(logger.Key, blob.TreeHeadData, blob.Signature)

	case mtr.AlertTypeID:
		monitor := allMonitors.FindMonitorByMonitorID(data.Signer)
		blob, err := data.DeconstructAlert();
		if err != nil{
			fmt.Errorf("Error deconstructing Alert: %s\n", err)
			return false
		}
		signature_err = signature.VerifySignature(monitor.MonitorKey, blob.TBS, blob.Signature)

	case mtr.STHPOCTypeID:
		logger := allLogs.FindLogByLogID(data.Signer);
		blob, err := data.DeconstructSTH()
		if err != nil{
			fmt.Errorf("Error deconstructing STH_POC: %s\n", err)
			return false
		}
		signature_err = signature.VerifySignature(logger.Key, blob.TreeHeadData, blob.Signature)

	case mtr.ConflictingSTHPOMTypeID:
		blob, err := data.DeconstructConflictingSTHPOM()//ConflictingSTHPOM
		if err != nil{
			fmt.Errorf("Error deconstructing Conflicting STH: %s\n", err)
			return false
		}
		logger := allLogs.FindLogByLogID(blob.STH1.LogID);
		signature_err = signature.VerifySignature(logger.Key, blob.STH1, blob.STH1.Signature)
		if signature_err != nil{
			break
		}

		logger = allLogs.FindLogByLogID(blob.STH2.LogID);
		signature_err = signature.VerifySignature(logger.Key, blob.STH2, blob.STH2.Signature)

	default:
		signature_err = fmt.Errorf("Unknown type %v\n", data.TypeID)
		break
	}

	if signature_err != nil{
		fmt.Errorf("%v\n", signature_err)
		return false
	}

	return true
}
