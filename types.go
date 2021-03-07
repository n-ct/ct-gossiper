package CTObject

import (
	"fmt"
	"encoding/json"
	mtrUtils "github.com/n-ct/ct-monitor/utils"
	mtr "github.com/n-ct/ct-monitor"
)

const(
	GossipPath = "/ct/v1/gossip"
	Threshold = 1000000
)

type MessagesMap map[string]map[string]map[uint64]map[string] *mtr.CTObject;


// returns a pointer to a new CTObject struct with all the same elements but with the 'Blob' field set to nil.
func CopyWithoutBlob(data *mtr.CTObject) (*mtr.CTObject){
	cto := mtr.CTObject{};
	cto.TypeID = data.TypeID;
	cto.Version = data.Version;
	cto.Timestamp = data.Timestamp;
	cto.Signer = data.Signer;
	cto.Subject = data.Subject;
	cto.Digest = data.Digest;
	cto.Blob = nil; //blob is nil
	return &cto;
}

//return the information from CTObject for printing
func ToDebugString(data mtr.CTObject) (string){
	return fmt.Sprintf("CTObject{\n\tTypeID:%v\n\tTimestamp:%v\n\tDigest:%v\n\tBlob:%v\n}", data.TypeID, data.Timestamp, data.Digest, data.Blob);
}

func IdentifierToString(id mtr.ObjectIdentifier) string {
	return fmt.Sprintf("[%s:%s:%s:%s]", id.First, id.Second, id.Third, id.Fourth)
}

//structs for JSON files
type GossipConfig struct{
	Monitors_ids []string `json:"monitor_ids"`
	Monitor_id string `json:"monitor_id"`
}

func NewGossipConfig (filename string) (*GossipConfig) {
	var gossipConfig GossipConfig;
	byteData, err := mtrUtils.FiletoBytes(filename)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(byteData, &gossipConfig); err != nil {
		fmt.Errorf("Failed to parse gossip configuration: %v", err);
		return nil;
	}
	return &gossipConfig;
}
