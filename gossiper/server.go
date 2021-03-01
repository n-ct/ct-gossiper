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

	"github.com/golang/glog"
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
		glog.Infoln("kill recived");
		glog.Flush();
		os.Exit(1);
	}(); //when channel is notified print debug info, make sure all logs get written and exit

	//Setting flags
  var configFilename = flag.String("config", "", "File containing gossiper configuration");
	var monitorsFilename = flag.String("monitor_list", "", "File containing monitor-gossiper pairs");
	var logsFilename = flag.String("log_list", "", "File containing the list of logs");

	flag.Parse();
	defer glog.Flush(); //if the program ends unexpectedly make sure all debug info is printed

	//if filenames are not defined, terminate
	if len(*configFilename) == 0 || len(*monitorsFilename) == 0 || len(*logsFilename) == 0 {
    fmt.Println("configuration files are required.")
    return
  }

	messages = make(cto.MessagesMap);
	alertsMap = make(cto.MessagesMap);

	gossiperSetup(*configFilename, *monitorsFilename, *logsFilename);

	http.HandleFunc(cto.GossipPath, GossipHandler); // call GossipHandler on post to /gossip

	glog.Infof("Starting server on %v\n", port);

	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil); // start server
	if err != nil {
		fmt.Errorf("err: %v", err);
	}
}

// GossipHandler is called on a post request to /ct/v1/gossip.
// It handles the logic of gossip within a network system
func GossipHandler(w http.ResponseWriter, req *http.Request){
	data := mtr.CTObject{};
	err := json.NewDecoder(req.Body).Decode(&data); // fill that struct using the JSON encoded struct send via the post
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

	glog.Infof("%s Received request\n", identifierStr)
	//glog.Infof("CTObject received from source: %s\n\n", cto.ToDebugString(data)); //print contents of message (for debugging)
	if message, ok := workingMap[identifier.First][identifier.Second][identifier.Third][identifier.Fourth]; ok { // if I have the message already check for conflict
		if bytes.Compare(data.Digest, message.Digest)==0 {
			glog.Infof("%s Duplicate Item\n\n", identifierStr)
			http.Error(w, "Duplicate item\n", http.StatusBadRequest); // if no conflic send back "duplicate item", and bad request status code to sender

		} else {
			glog.Infof("%s Misbehavior detected\n", identifierStr); // if conflict send a PoM to all peers
			PoM, err := mtr.CreateConflictingSTHPOM(&data, message)
			if err == nil {
				addEntry(messages, *PoM, PoM.Identifier()); // store PoM
				glog.Infof("%s Stored PoM")
				gossipPeers(PoM, requesterAddress)
				gossipMonitor(PoM, requesterAddress)
				glog.Infof("%s Finished gossiping PoM\n\n", identifierStr)

			} else {
				glog.Error(err)
			}
		}
		} else {
			if data.Blob == nil{ //If the message does not contain the blob
				glog.Infof("%s blob-request sent\n", identifierStr)
				fmt.Fprintf(w, "blob-request"); //respond with "blob-request"

		} else {
			if validateSignature(&data) {
				fmt.Fprintf(w, "new data"); //respond with "new data"
				addEntry(workingMap, data, identifier);// if message is new add it to messages map
				glog.Infof("%s Stored new data\n", identifierStr)
				gossipPeers(&data, requesterAddress)
				gossipMonitor(&data, requesterAddress)
				glog.Infof("%s Finished gossiping new data\n\n", identifierStr)

			} else {
				//invalid Signature
				glog.Infof("%s Invalid Signature\n\n", identifierStr)
				http.Error(w, "Invalid Signature\n", http.StatusBadRequest);
			}
		}
	}
}

// post takes in an address as a string and a pointer to a CTObject struct
// and makes a post request to that address with the JSON encoded version of that struct
func post(address string, data *mtr.CTObject, withBlob bool){
	var toSend *mtr.CTObject;
	if withBlob {
		toSend = data;
	} else {
		toSend = cto.CopyWithoutBlob(data);
	}
	var jsonStr, _ = json.Marshal(toSend);

	req, err := http.NewRequest("POST", address, bytes.NewBuffer(jsonStr)); //create a post request
	req.Header.Set("X-Custom-Header", "myvalue");
	req.Header.Set("Content-Type", "application/json"); //set message type to JSON
	req.Header.Add("requesterAddress", myAddress);

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
		glog.Infof("Sending blob to peer: %v\n", address);
		post(address, data, true); // if the recipient sends back a blob request resend the message with the blob
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

//gossiperSetup configures gossiper varialbes from json files
func gossiperSetup(configFilename string, monitorsFilename string, logsFilename string){
	gossipConfig = cto.NewGossipConfig(configFilename);

	getPeers(gossipConfig, monitorsFilename);
	myAddress = allMonitors.FindMonitorByMonitorID(gossipConfig.Monitor_id).GossiperURL;
	port = strings.Split(myAddress, ":")[2];

	allLogs = mtrList.NewLogList(logsFilename); //get all logs
	glog.Infoln("Setup completed")
}

//getPeers get all monitors from file and populates the peers with the respective monitors
func getPeers(gossipConfig *cto.GossipConfig, monitorsFilename string)  {
	allMonitors = mtrList.NewMonitorList(monitorsFilename)

	for _, monitorId := range gossipConfig.Monitors_ids {
		peers = append(peers, allMonitors.FindMonitorByMonitorID(monitorId));
	}
}

//gossipPeers sends new data to other gossip servers
func gossipPeers(data *mtr.CTObject, requesterAddress string){
	for _, peer := range peers{

		if requesterAddress != peer.GossiperURL{
			glog.Infof("Gossiping info to peer: %v\n", peer.MonitorID);
			post(mtrUtils.CreateRequestURL(peer.GossiperURL, cto.GossipPath), data, false);
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
		glog.Infoln("Monitor unreachable.")
	} else {
		post(mtrUtils.CreateRequestURL(monitorUrl, mtr.NewInfoPath), data, true)
	}
}

func validateSignature(data *mtr.CTObject) bool {

	blob := mtr.DeconstructCTObject(data)
	var err error

	switch v := blob.(type) {
	case mtr.SignedTreeHeadData:
		logger := allLogs.FindLogByLogID(data.Signer);
		err = signature.VerifySignature(logger.Key, blob.(mtr.SignedTreeHeadData).TreeHeadData, blob.(mtr.SignedTreeHeadData).Signature)

	case mtr.Alert:
		monitor := allMonitors.FindMonitorByMonitorID(data.Signer)
		err = signature.VerifySignature(monitor.MonitorKey, blob.(mtr.Alert).TBS, blob.(mtr.Alert).Signature)

	case mtr.SignedTreeHeadWithConsistencyProof:
		logger := allLogs.FindLogByLogID(data.Signer);
		err = signature.VerifySignature(logger.Key, blob.(mtr.SignedTreeHeadWithConsistencyProof).SignedTreeHead.TreeHeadData, blob.(mtr.SignedTreeHeadWithConsistencyProof).SignedTreeHead.Signature)

	default:
		err = fmt.Errorf("Unknown type %v\n", v)
	}

	if err != nil{
		glog.Infof("%v\n", err)
		return false
	}

	return true
}
