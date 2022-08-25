package structs

type Channel struct {
	Suscribers int
	Files      []string
}

type ReadFromChannel struct {
	Alias    string
	Response chan ChannelState
}

type WriteToChannel struct {
	Alias     string
	Suscriber int
	File      string
	Response  chan bool
}

type ChannelState struct {
	Suscribers int
	LastFile   string
}
