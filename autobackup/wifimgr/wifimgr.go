package wifimgr

import(
    "fmt"
)


type Config struct{
     Ssid string
     Password string
     Security string
}


func CreateConnection() error{
   
    config, err := openWifiConfig()
     
    if err!= nil{
         fmt.Println("open wifi config file error")
    }

    fmt.Println( *config)
   

    return nil
}


func openWifiConfig() (*Config, error){

   
    return &Config{ "Martin_hotspot", 
                   "12345678",
                    "WPA2", }, nil

}

func IsWifiConnected() bool{

    return true
}

