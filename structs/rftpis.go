package structs

type Channel struct {
	Id         string
	Suscribers map[string]*Suscriber
	Files      []string
}

type ChannelState struct {
	Id         string
	Suscribers map[string]Suscriber
	Files      []string
}

type ChannelStateComm struct {
	Read  chan ReadChannelState
	Write chan WriteChannelState
}

type ReadChannelState struct {
	Alias    string
	Response chan ChannelState
}

type WriteChannelState struct {
	Alias            string
	AddSuscriber     bool
	AddSuscriberData Suscriber
	RemoveSuscriber  string
	AddFile          string
	RemoveFile       string
	Response         chan bool
}

type Suscriber struct {
	Id   string
	Comm SuscriberComm
}

type SuscriberComm struct {
	Read  chan ReadChannelMemory
	Write chan WriteChannelMemory
}

type TransmissionStatusComm struct {
	Read  chan ReadTransmissionStatus
	Write chan WriteTranmissionStatus
}

type TransmissionStatus struct {
	File          string
	IsTransfering bool
	IsError       bool
}

type ReadTransmissionStatus struct {
	Alias    string
	Response chan TransmissionStatus
}

type WriteTranmissionStatus struct {
	Alias    string
	Data     TransmissionStatus
	Response chan bool
}

type ChannelMemoryComm struct {
	Read  chan ReadChannelMemory
	Write chan WriteChannelMemory
}

type ChannelMemory struct {
	Data      []byte
	IsMessage bool
	Count     int
	Id        string
	IsEOF     bool
}

type ReadChannelMemory struct {
	Response chan ChannelMemory
}

type WriteChannelMemory struct {
	Data     ChannelMemory
	Response chan bool
}
