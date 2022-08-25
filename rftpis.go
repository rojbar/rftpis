package rftpis

import (
	"bufio"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/google/uuid"
	utils "github.com/rojbar/rftpiu"
)

const BUFFERSIZE = 4096

// OK
func Server() {
	ln, err := net.Listen("tcp", ":5000")
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, errA := ln.Accept()
		if errA != nil {
			print(errA)
			continue
		}
		go recieve(conn)
	}
}

//OK
func recieve(conn net.Conn) {
	// here we recieve the request
	message, errR := utils.ReadMessage(conn)
	if errR != nil {
		print(errR)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		conn.Close()
		return
	}

	// here we parse
	action, errA := utils.GetKey(message, "ACTION")
	if errA != nil {
		print(errA)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: MALFORMED_REQUEST;")
		conn.Close()
		return
	}

	//redirect according to client action
	switch action {
	case "SEND":
		go handleRecieveFile(conn, message)

	case "SUBSCRIBE":
		go handleSubscription(conn, message)

	default:
		conn.Close()
	}
}

//NOT OK
func handleSubscription(conn net.Conn, message string) {
	defer conn.Close()
	utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
}

//OK
func handleRecieveFile(conn net.Conn, message string) {
	defer conn.Close()
	errI := utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
	if errI != nil {
		print(errI)
		return
	}

	value, errSz := utils.GetKey(message, "SIZE")
	if errSz != nil {
		print(errSz)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}
	fileSize, errAtoi := strconv.Atoi(value)
	if errAtoi != nil || fileSize <= 0 {
		print(errAtoi)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}

	extension, errExt := utils.GetKey(message, "EXTENSION")
	if errExt != nil {
		extension = " "
		print(errExt)
	}

	buffer := make([]byte, BUFFERSIZE)
	loops := fileSize / BUFFERSIZE
	sizeLastRead := fileSize % BUFFERSIZE

	lessBuffer := make([]byte, sizeLastRead)

	file, errC := os.Create(uuid.NewString() + "." + extension)
	if errC != nil {
		print(errC)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for i := 0; i < int(loops); i++ {
		errRnWf := readNetWriteFile(conn, buffer, *writer)
		if errRnWf != nil {
			print(errRnWf)
			utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
			return
		}
	}
	errRnWf := readNetWriteFile(conn, lessBuffer, *writer)
	if errRnWf != nil {
		print(errRnWf)
		utils.SendMessage(conn, "SFTP > 1.0 STATUS: NOT OK;")
		return
	}

	utils.SendMessage(conn, "SFTP > 1.0 STATUS: OK;")
}

func readNetWriteFile(conn net.Conn, buffer []byte, writer bufio.Writer) error {
	_, errR := io.ReadFull(conn, buffer)
	if errR != nil {
		if errR == io.EOF {
			print(errR)
		}
		return errR
	}

	_, errW := writer.Write(buffer)
	if errW != nil {
		return errW
	}
	errF := writer.Flush()
	if errF != nil {
		return errF
	}

	return nil
}
