package rftpis

import (
	"net"
	"os"
	"strconv"

	"github.com/rojbar/rftpis/handlers"
	"github.com/rojbar/rftpis/mgmt"
	"github.com/rojbar/rftpis/structs"
	utils "github.com/rojbar/rftpiu"
)

const BUFFERSIZE = 4096

// OK
func Server() {
	utils.InitializeLogger()
	channelComm, channelMemoryComm := mgmt.Init()

	for i := 0; i < 50; i++ {
		err := os.MkdirAll("recieve/channels/"+strconv.Itoa(i), 0750)
		if err != nil && !os.IsExist(err) {
			utils.Logger.Panic(err.Error())
		}
	}

	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		utils.Logger.Panic(err.Error())
	}
	utils.Logger.Info("initiated server in port 5000")
	defer ln.Close()
	for {
		conn, errA := ln.Accept()
		if errA != nil {
			utils.Logger.Error(errA.Error())
			continue
		}
		utils.Logger.Info("recieved new client connection")
		go recieve(conn, channelComm, channelMemoryComm)
	}
}

//OK
func recieve(conn net.Conn, chComm structs.ChannelStateComm, chMeComm map[string]structs.ChannelMemoryComm) {
	// here we recieve the request
	message, errR := utils.ReadMessage(conn)
	if errR != nil {
		utils.Logger.Error(errR.Error())
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		conn.Close()
		return
	}

	// here we parse
	action, errA := utils.GetKey(message, "ACTION")
	channel, errCh := utils.GetKey(message, "CHANNEL")
	if errA != nil || errCh != nil {
		utils.Logger.Error(errA.Error())
		utils.Logger.Error(errCh.Error())
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	//redirect according to client action
	switch action {
	case "SEND":
		go handlers.HandleRecieveFile(conn, message, chComm)
	case "SUBSCRIBE":
		go handlers.HandleSubscription(conn, message, chComm, chMeComm[channel])
	default:
		conn.Close()
	}
}
