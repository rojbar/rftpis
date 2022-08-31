package structs

type Channel struct {
	Suscribers int
	Files      []string
}

type ChannelStateComm struct {
	Read  chan ReadChannelState
	Write chan WriteChannelState
}

type ChannelState struct {
	Suscribers int
	LastFile   string
}

type ReadChannelState struct {
	Alias    string
	Response chan ChannelState
}

type WriteChannelState struct {
	Alias    string
	Data     ChannelState
	Response chan bool
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
}

type ReadChannelMemory struct {
	Response chan ChannelMemory
}

type WriteChannelMemory struct {
	Data     ChannelMemory
	Response chan bool
}
