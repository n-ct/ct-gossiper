package CTData

import (
  "time"
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
    TBS       SignedFields
    Signature []byte
}

// all fields in a CTData that must be signed
type SignedFields struct{
  TypeID      string
  Signer      int32
  Timestamp   int64
  Blob        []byte
  Digest      []byte
  Subject     int32
  AlgorithmID string
}

// returns a pointer to a CTData struct with a given timestamp and blob, all other fields are placeholders
func NewCTData(Timestamp int64, Blob []byte) (*CTData){
  ctd := CTData{};
  tbs := SignedFields{};
  tbs.TypeID = "Dummy";
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

// returns a pointer to a CTData struct with a TypeId of "DummyPoM", and a timestamp of the current unix timestamp, all other fields are placeholders
func NewDummyPoM() (*CTData){
  ctd := CTData{};
  tbs := SignedFields{};
  tbs.TypeID = "DummyPoM";
  tbs.Signer = -1;
  tbs.Timestamp = time.Now().Unix();
  tbs.Blob = []byte{255,127,63,31,15,7,3,1};
  tbs.Digest = NewSHA256(tbs.Blob);
  tbs.Subject = -1;
  tbs.AlgorithmID = "SHA256";
  ctd.TBS = tbs;
  ctd.Signature = []byte{0b1010101,0b0101010,0b1010101,0b0101010};
  return &ctd;
}

// returns a pointer to a CTData struct with a timestamp of the current unix timestamp, all other fields are placeholders
func DummyCTData() (*CTData){
  ctd := CTData{};
  tbs := SignedFields{};
  tbs.TypeID = "Dummy";
  tbs.Signer = -1;
  tbs.Timestamp = time.Now().Unix();
  tbs.Blob = []byte{0b1001000, 0b1100101, 0b1101100, 0b1101100, 0b1101111, 0b101100, 0b100000, 0b1110111, 0b1101111, 0b1110010, 0b1101100, 0b1100100, 0b100001};
  tbs.Digest = NewSHA256(tbs.Blob);
  tbs.Subject = -1;
  tbs.AlgorithmID = "SHA256";
  ctd.TBS = tbs;
  ctd.Signature = []byte{0b0000000,0b0000001,0b0000010,0b0000011};
  return &ctd;
}

// returns the identifier of a struct
// TODO calculate this once and save rather than recalculate?
func (data *CTData) Identifier() (string){
  return fmt.Sprintf("%s%v%v",data.TBS.TypeID,data.TBS.Signer,data.TBS.Timestamp);
}
