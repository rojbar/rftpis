package mgmt

import (
	"fmt"

	"github.com/rojbar/rftpis/structs"
)

func channelMemory(alias string, chMeComm structs.ChannelMemoryComm) {
	messages := make([][]byte, 3)

	fmt.Println("channelMemory", alias, "current memory", messages)

}
