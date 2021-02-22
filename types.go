package CTObject

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"encoding/json"
	mtrUtils "github.com/n-ct/ct-monitor/utils"
)

// function to compute sha256 of a []byte
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

type MessagesMap map[string]map[string]map[uint64]map[string] *CTObject;

type CTObject struct{
  TypeID string
  Version string
  Timestamp uint64
  Signer string
  Hash []byte
  Subject string
  Blob []byte
}

func NewCTObject(TypeID string, Timestamp uint64, Blob []byte, Subject string) (*CTObject){
	cto := CTObject{};
	cto.TypeID = TypeID;
  cto.Timestamp = Timestamp;
	cto.Signer = "";
  cto.Hash = NewSHA256(Blob);
	cto.Blob = Blob;
	cto.Subject = Subject;
	return &cto;
}

// returns a pointer to a new CTObject struct with all the same elements but with the 'Blob' field set to nil.
func (data *CTObject) CopyWithoutBlob() (*CTObject){
	cto := CTObject{};
	cto.TypeID = data.TypeID;
	cto.Timestamp = data.Timestamp;
	cto.Signer = data.Signer;
	cto.Hash = data.Hash;
	cto.Blob = nil; //blob is nil
	cto.Subject = data.Subject;
	return &cto;
}

//return the information from CTObject for printing
func (data *CTObject) ToDebugString() (string){
	return fmt.Sprintf("CTObject{\n\tTypeID:%v\n\tTimestamp:%v\n\tHash:%v\n\tBlob:%v\n}", data.TypeID, data.Timestamp, data.Hash, data.Blob);
}

type ObjectIdentifier struct{
	First string
	Second string
	Third uint64
	Fourth string
}

//creates identifier for each type of CTObject
//Alerts [Subject][Signer][Timestamp][Version]
//The rest [TypeID][Subject|Signer][Timestamp][Version]
func (data *CTObject) Identifier() (ObjectIdentifier){
	if data.TypeID == "Alert"{
		return ObjectIdentifier{First: data.Subject, Second: data.Signer, Third: data.Timestamp, Fourth: data.Version,};
	}

	var subjectOrSigner string;
	if len(data.Subject) == 0 {
		subjectOrSigner = data.Signer;
	} else {
		subjectOrSigner = data.Subject;
	}
	return ObjectIdentifier{First: data.TypeID, Second: subjectOrSigner, Third: data.Timestamp, Fourth: data.Version,};
}

//structs for JSON files
type GossipConfig struct{
	Monitors_ids []string `json:"monitor_ids"`
	Monitor_id string `json:"monitor_id"`
}

func NewGossipConfig (filename string) (*GossipConfig) {
	var gossipConfig GossipConfig;
	byteData := mtrUtils.JSONFiletoBytes(configFilename);
	if err := json.Unmarshal(byteData, &gossipConfig); err != nil {
		fmt.Errorf("Failed to parse gossip configuration: %v", err);
		return nil;
	}
	return gossipConfig;
}
/*
type MonitorsList struct {
	Operators []*Operator `json:"operators"`
}

type Operator struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Monitors []*Monitor `json:"monitors"`
}

type Monitor struct {
	Id string	`json:"monitor_id"`
	Key string 	`json:"monitor_key"`
	Url string	`json:"monitor_url"`
	Gossip string	`json:"gossiper_url"`
}

func NewMonitorList(monitorsFilename string) *MonitorsList{
	byteData := mtrUtils.JSONFiletoBytes(monitorsFilename)
	monitorList, err := NewFromJSON(byteData)
	if err != nil {
		return nil
	}
	return monitorList
}

// NewFromJSON creates a MonitorList from JSON data.
func NewFromJSON(data []byte) (*MonitorsList, error) {
	var ml MonitorsList
	if err := json.Unmarshal(data, &ml); err != nil {
		return nil, fmt.Errorf("Failed to parse monitor list: %v", err)
	}
	return &ml, nil
}

//Find and return specific monitor based on its ID
func (ml *MonitorsList) FindMonitorByID(monitorID string) (*Monitor) {
	for _, op := range ml.Operators {
		for _, monitor := range op.Monitors {
			if (strings.Contains(monitor.Id, monitorID)){
				return monitor
			}
		}
	}
	return nil
}
*/
