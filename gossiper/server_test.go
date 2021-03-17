package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  "net/http"
  "net/http/httptest"
  "testing"

  mtr "github.com/n-ct/ct-monitor"
  cto "github.com/n-ct/ct-gossiper"
)

const(
	GossipPath = "/ct/v1/gossip"
)


var version = mtr.VersionData{1,0,0}
var timestamp uint64 =  1615091086773
var loggerSigner = "sh4FzIuizYogTodm+Su5iiUgZ2va+nDnsklTLe+LkF4="

var sthSubject = ""
var sthDigest = []byte{107, 0, 118, 234, 17, 21, 127, 36, 39, 131, 239, 139, 10, 214, 51, 95, 42, 2, 50, 194, 236, 189, 111, 26, 76, 253, 212, 170, 83, 155, 91, 23}
var sthBlob = []byte{123, 34, 76, 111, 103, 73, 68, 34, 58, 34, 115, 104, 52, 70, 122, 73, 117, 105, 122, 89, 111, 103, 84, 111, 100, 109, 43, 83, 117, 53, 105, 105, 85, 103, 90, 50, 118, 97, 43, 110, 68, 110, 115, 107, 108, 84, 76, 101, 43, 76, 107, 70, 52, 61, 34, 44, 34, 84, 114, 101, 101, 72, 101, 97, 100, 68, 97, 116, 97, 34, 58, 123, 34, 86, 101, 114, 115, 105, 111, 110, 34, 58, 48, 44, 34, 83, 105, 103, 110, 97, 116, 117, 114, 101, 84, 121, 112, 101, 34, 58, 49, 44, 34, 84, 105, 109, 101, 115, 116, 97, 109, 112, 34, 58, 49, 54, 49, 53, 48, 57, 49, 48, 56, 54, 55, 55, 51, 44, 34, 84, 114, 101, 101, 83, 105, 122, 101, 34, 58, 57, 54, 49, 57, 50, 50, 55, 54, 52, 44, 34, 83, 72, 65, 50, 53, 54, 82, 111, 111, 116, 72, 97, 115, 104, 34, 58, 34, 90, 57, 110, 77, 73, 76, 88, 116, 98, 101, 88, 84, 67, 114, 75, 89, 106, 100, 48, 77, 77, 74, 71, 84, 48, 97, 121, 76, 106, 119, 89, 54, 106, 118, 88, 110, 71, 109, 52, 107, 79, 43, 103, 61, 34, 125, 44, 34, 83, 105, 103, 110, 97, 116, 117, 114, 101, 34, 58, 34, 66, 65, 77, 65, 83, 68, 66, 71, 65, 105, 69, 65, 50, 90, 74, 53, 65, 121, 51, 56, 83, 86, 73, 105, 78, 71, 120, 97, 87, 48, 116, 48, 109, 112, 116, 80, 112, 99, 81, 49, 71, 110, 101, 51, 70, 73, 43, 51, 72, 75, 70, 108, 88, 50, 85, 67, 73, 81, 68, 73, 67, 115, 120, 101, 55, 69, 57, 53, 48, 98, 121, 86, 115, 115, 79, 51, 88, 108, 51, 90, 81, 76, 101, 113, 86, 65, 97, 71, 48, 102, 68, 65, 112, 84, 51, 52, 68, 55, 67, 111, 109, 119, 61, 61, 34, 125}
//var sthDigestConflict = []byte{107, 1, 118, 234, 17, 21, 127, 36, 39, 131, 239, 139, 10, 214, 51, 95, 42, 2, 50, 194, 236, 189, 111, 26, 76, 253, 212, 170, 83, 155, 91, 23}

var sthInvalidSignatureBlob = []byte{123, 34, 76, 111, 103, 73, 68, 34, 58, 34, 115, 104, 52, 70, 122, 73, 117, 105, 122, 89, 111, 103, 84, 111, 100, 109, 43, 83, 117, 53, 105, 105, 85, 103, 90, 50, 118, 97, 43, 110, 68, 110, 115, 107, 108, 84, 76, 101, 43, 76, 107, 70, 52, 61, 34, 44, 34, 84, 114, 101, 101, 72, 101, 97, 100, 68, 97, 116, 97, 34, 58, 123, 34, 86, 101, 114, 115, 105, 111, 110, 34, 58, 48, 44, 34, 83, 105, 103, 110, 97, 116, 117, 114, 101, 84, 121, 112, 101, 34, 58, 49, 44, 34, 84, 105, 109, 101, 115, 116, 97, 109, 112, 34, 58, 49, 54, 49, 53, 48, 57, 49, 48, 56, 54, 55, 55, 51, 44, 34, 84, 114, 101, 101, 83, 105, 122, 101, 34, 58, 57, 54, 49, 57, 50, 50, 55, 54, 52, 44, 34, 83, 72, 65, 50, 53, 54, 82, 111, 111, 116, 72, 97, 115, 104, 34, 58, 34, 90, 57, 110, 77, 73, 76, 88, 116, 98, 101, 88, 84, 67, 114, 75, 89, 106, 100, 48, 77, 77, 74, 71, 84, 48, 97, 121, 76, 106, 119, 89, 54, 106, 118, 88, 110, 71, 109, 52, 107, 79, 43, 103, 61, 34, 125, 44, 34, 83, 105, 103, 110, 97, 116, 117, 114, 101, 34, 58, 34, 66, 65, 77, 65, 83, 68, 66, 71, 65, 105, 69, 65, 50, 90, 74, 53, 65, 121, 51, 56, 83, 86, 73, 105, 78, 71, 120, 97, 87, 48, 116, 48, 109, 112, 116, 80, 112, 99, 81, 49, 71, 110, 101, 51, 70, 73, 43, 51, 72, 75, 70, 108, 88, 50, 85, 67, 73, 81, 68, 73, 67, 115, 120, 101, 55, 69, 57, 53, 48, 98, 121, 86, 115, 115, 79, 51, 88, 108, 51, 90, 81, 76, 101, 113, 86, 65, 97, 71, 48, 102, 68, 65, 112, 84, 51, 52, 68, 55, 67, 111, 109, 119, 61, 61, 34, 126}

var sthCTObject = mtr.CTObject{mtr.STHTypeID, version, timestamp, loggerSigner, sthSubject, sthDigest, sthBlob}
var withoutBlobCTObject = mtr.CTObject{mtr.STHTypeID, version, timestamp, loggerSigner, sthSubject, sthDigest, nil}
var sthInvalidCTObject = mtr.CTObject{mtr.STHTypeID, version, timestamp, loggerSigner, sthSubject, sthDigest, sthInvalidSignatureBlob}

var logFilename = "../testing/config/log_list.json"
var monitorFilename =  "../testing/config/monitors_list.json"
var configFilename = "../testing/config/config1.json"

func mustGossiperSetup(t *testing.T){
  messages = make(cto.MessagesMap);
  alertsMap = make(cto.MessagesMap);
  GossiperSetup(configFilename, monitorFilename, logFilename)
}

func TestGossipHandler(t *testing.T)  {
  mustGossiperSetup(t)

  testTables := []struct {
    object *mtr.CTObject
    expected string

  }{
    // {&sthInvalidCTObject, "invalid data"},
    {&withoutBlobCTObject, "blob-request"},
    {&sthCTObject, "new data"},
    //{&sthCTObject, " Duplicate item"}, issues with http.StatusBadRequest
  }
  var req *http.Request
  var err error
  var jsonStr []byte
  var recorder *httptest.ResponseRecorder

  handler := http.HandlerFunc(GossipHandler)


  for _, testTable := range testTables{
    recorder = httptest.NewRecorder()
    jsonStr, _ = json.Marshal(testTable.object);
    req, err = http.NewRequest("POST", GossipPath, bytes.NewBuffer(jsonStr))
    handler.ServeHTTP(recorder, req)
    if err != nil {
          t.Fatal(err)
    }

    if recorder.Body.String() != testTable.expected {
          t.Error(fmt.Sprintf("Handler returned %v want %v for %v:\n", recorder.Body.String(), testTable.expected, testTable.object.TypeID))
    }
  }
}

//In progress. Test table
func TestValidateSignature(t *testing.T){
  testTables := []struct {
    object *mtr.CTObject
    expected bool

  }{
    {&sthCTObject, true},
    {&withoutBlobCTObject, false},
    {&sthInvalidCTObject, false},

  }

  for _, testTable := range testTables{
    result := ValidateSignature(testTable.object)
    if testTable.expected != result {
      t.Errorf("Error validating %s, got %v when expecting %v", testTable.object.TypeID, result, testTable.expected)
    }
  }
}//TestValidateSignature
