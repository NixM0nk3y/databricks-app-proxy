package version

import "net/http"

// Version of the program.
var Version = "SNAPSHOT"

// Commit Hash
var BuildHash = "AAAAAAAA"

// date the program was built
var BuildDate = "19760101"

type VersionResponse struct {
	Version   string `json:"version"`
	BuildHash string `json:"buildhash"`
	BuildDate string `json:"builddate"`
}

func (u *VersionResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
