package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var ErrorLogger *log.Logger
var InfoLogger *log.Logger

func CreateLogFile(name string) *os.File {
	file, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error creating log file : ", err)
		return nil
	}
	return file
}

func ioMultiWriter(writers ...io.Writer) io.Writer {
	return io.MultiWriter(writers...)
}

// Method to initialize the multiwrite logger
func InitializeLogger() {
	ErrorLogger = log.New(ioMultiWriter(os.Stdout, CreateLogFile("einv_sol.log")), "Error:", log.Ldate|log.Ltime)
	InfoLogger = log.New(ioMultiWriter(os.Stdout, CreateLogFile("einv_sol.log")), "Info:", log.Ldate|log.Ltime)
}

func CheckError(cust_info string, err error) {
	if err != nil {
		ErrorLogger.Fatalln(cust_info+":", err)
		//ErrorLogger.Fatal()
	}
}
