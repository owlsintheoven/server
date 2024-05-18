package common

import (
	"bufio"
	"log"
	"os"
)

func ReadByte(reader *bufio.Reader) (byte, error) {
	buffer := make([]byte, 1)
	n, err := reader.Read(buffer)
	if err != nil {
		log.Println("readByte error", err.Error())
		return 0, err
	}
	log.Printf("read %d bytes\n", n)
	return buffer[0], nil
}

func ReadBytes(reader *bufio.Reader, num int) ([]byte, error) {
	buffer := make([]byte, num)
	_, err := reader.Read(buffer)
	if err != nil {
		log.Println("readBytes error", err.Error())
		return nil, err
	}
	return buffer, nil
}

func ReadBytesTilNil(reader *bufio.Reader) ([]byte, error) {
	var buffer []byte
	buffer, err := reader.ReadBytes(0)
	if err != nil {
		log.Println("readBytesTilNil error", err.Error())
		return nil, err
	}
	return buffer, nil
}

// ScanFromStdin asks for a string value using the label
func ScanFromStdin() string {
	//var s string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	return scanner.Text()
}
