package CTData

import (
	"crypto/sha256"
	"fmt"
)

// function to compute sha256 of a []byte
func NewSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// struct to hold all fields for a CTData message
type CTData struct{
	TBS			SignedFields
	Signature	[]byte
}

// all fields in a CTData that must be signed
type SignedFields struct{
	TypeID			string
	Signer			int32
	Timestamp		int64
	Blob			[]byte
	Digest			[]byte
	Subject			int32
	AlgorithmID		string
}

// returns a pointer to a CTData struct with a given typreId, timestamp and blob, all other fields are placeholders
func NewCTData(TypeID string, Timestamp int64, Blob []byte) (*CTData){
	ctd := CTData{};
	tbs := SignedFields{};
	tbs.TypeID = TypeID;
	tbs.Signer = -1;
	tbs.Timestamp = Timestamp;
	tbs.Blob = Blob;
	tbs.Digest = NewSHA256(tbs.Blob);
	tbs.Subject = -1;
	tbs.AlgorithmID = "SHA256";
	ctd.TBS = tbs;
	ctd.Signature = []byte{0b0000000,0b0000001,0b0000010,0b0000011};
	return &ctd;
}

// returns a pointer to a new CTData struct with all the same elements but with the 'Blob' field set to nil.
func (data *CTData) CopyWithoutBlob() (*CTData){
	ctd := CTData{};
	tbs := SignedFields{};
	tbs.TypeID = data.TBS.TypeID;
	tbs.Signer = data.TBS.Signer;
	tbs.Timestamp = data.TBS.Timestamp;

	tbs.Blob = nil; //blob is nil

	tbs.Digest = data.TBS.Digest; //however is is important to note that the digest remains in tact
	tbs.Subject = data.TBS.Subject;
	tbs.AlgorithmID = data.TBS.AlgorithmID;
	ctd.TBS = tbs;
	ctd.Signature = data.Signature;
	return &ctd;
}

// returns the identifier of a struct
// TODO calculate this once and save rather than recalculate?
func (data *CTData) Identifier() (string){
	return fmt.Sprintf("%s%v%v",data.TBS.TypeID,data.TBS.Signer,data.TBS.Timestamp);
}

func (data *CTData) ToDebugString() (string){
	return fmt.Sprintf("CTData{\n\tTypeID:%v\n\tTimestamp:%v\n\tDigest:%v\n\tBlob:%v\n}",data.TBS.TypeID,data.TBS.Timestamp,data.TBS.Digest,data.TBS.Blob);
}
