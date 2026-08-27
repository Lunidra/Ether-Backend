package chat

// Message is the authoritative EtherChat message sent to clients.
//
// This intentionally mirrors the current Ether-Client ChatMessage model.
type Message struct {
	UUID        string `json:"uuid"`
	Username    string `json:"username"`
	Message     string `json:"message"`
	Server      string `json:"server"`
	ServerIP    string `json:"serverIP"`
	GameVersion string `json:"gameVersion"`

	IsStaff      bool `json:"isStaff"`
	IsDeveloper  bool `json:"isDeveloper"`
	IsBetaTester bool `json:"isBetaTester"`
}

// IncomingPacket is the current client → backend ChatPacket.
//
// The client wraps ChatMessage inside:
//
//	{
//	  "type": "chat",
//	  "data": {
//	    "message": "..."
//	  }
//	}
//
// We intentionally only care about Message here.
// Identity and presence are authoritative on the backend.
type IncomingPacket struct {
	Message string `json:"message"`
}
