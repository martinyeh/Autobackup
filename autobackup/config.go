package autobackup

import "flag"


var ConfigPath *string

func init() {
	ConfigPath  = flag.String("configPath", "../dropbox/backup_config.json", "backup tokens and some settings")
}


type Config struct {
	Provider     string
	AccessToken   string
	RefreshToken string
	Iskeep       bool
}
