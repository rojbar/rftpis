package handlers

import (
	"bufio"
	"net"

	"github.com/google/uuid"
	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
	"go.uber.org/zap"
)

//OK
func HandleSubscription(conn net.Conn, message string, chStComm structs.ChannelStateComm) {
	defer conn.Close()

	suscription := structs.Suscriber{
		Id: uuid.NewString(),
		Comm: structs.SuscriberComm{
			Read:  make(chan structs.ReadChannelMemory),
			Write: make(chan structs.WriteChannelMemory),
		},
	}

	utils.Logger.Info("initiated handle subscription recieved message", zap.String("message", message))
	channelName, _ := utils.GetKey(message, "CHANNEL")
	updateState := structs.WriteChannelState{
		Alias:            channelName,
		AddSuscriber:     true,
		AddSuscriberData: suscription,
		RemoveSuscriber:  "",
		AddFile:          "",
		RemoveFile:       "",
		Response:         make(chan bool),
	}
	chStComm.Write <- updateState
	<-updateState.Response
	utils.Logger.Info("handle subscription updated state")

	// here we inform he is subscribed
	errI := utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	if errI != nil {
		utils.Logger.Error("user didnt respond to maintain connection", zap.String("error", errI.Error()))
		return
	}

	//we read the channel
	for {
		readMessage := <-suscription.Comm.Write
		readMessage.Response <- true

		memoryMessage := readMessage.Data

		if memoryMessage.IsMessage {
			utils.Logger.Info("handle subscription sending message from memory", zap.String("message", string(memoryMessage.Data)))
			utils.SendMessage(conn, string(memoryMessage.Data))

			messageOk, errMess := utils.ReadMessage(conn)
			if errMess != nil {
				utils.Logger.Error(errMess.Error())
				return
			}

			isOk, errMesOk := utils.GetKey(messageOk, "STATUS")
			if errMesOk != nil {
				utils.Logger.Error(errMesOk.Error())
				return
			}

			if isOk != "OK" {
				return
			}

			utils.Logger.Info("hanlde subscription client ok initiating transfer of raw data")
			writer := bufio.NewWriter(conn)
			passed := memoryMessage
			for {
				readChunk := <-suscription.Comm.Write
				readChunk.Response <- true
				current := readChunk.Data

				if current.Id == "EMPTY MEMORY" {
					continue
				}

				if current.Id == passed.Id && current.Count == passed.Count {
					//we are reading a message that has already be send
					//fmt.Println("HANDLE SUBSCRIPTION: WE ARE READING A MESSAGE THAT HAS ALREADY BE SEND", len(current.Data), string(current.Data))
					continue
				}

				if current.Id != memoryMessage.Id || current.Count != passed.Count+1 {
					//we missed a memory chunk abort!
					utils.Logger.Info("handle subscription we missed a memory chunk abort!", zap.Int("chunkSize", len(current.Data)))
					break
				}

				if current.IsMessage {
					//a message is only meant to be send when a new file is gonna be transfer
					utils.Logger.Info("handle subscription we recieve a message an this FOR is only for sending chunks of data", zap.Int("chunkSize", len(current.Data)))
					break
				}
				//utils.Logger.Info("hanlde subscription sending chunk to client", zap.Int("chunk", current.Count), zap.Int("chunk size", len(current.Data)), zap.Any("a", string(current.Data)))

				_, errW := writer.Write(current.Data)
				if errW != nil {
					utils.Logger.Info("AH", zap.String("ah", errW.Error()))
					utils.Logger.Error(errW.Error())
				}
				errF := writer.Flush()
				if errF != nil {
					utils.Logger.Error(errF.Error())
				}
				passed = current

				if current.IsEOF {
					break
				}
			}
			utils.ReadMessage(conn)
			utils.Logger.Info("handle subscription recived file transfer ok by client")
		}
	}
}
