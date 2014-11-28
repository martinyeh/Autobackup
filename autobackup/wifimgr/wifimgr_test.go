package wifimgr

import(
    "testing"
)



func TestCreateConnection(t *testing.T) {
   
    if CreateConnection() != nil {
       t.Error("Create connection errors")        
    }

}

