package p2p

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"log"
	"testing"

	"fmt"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequest_Unpack(t *testing.T) {
	req := NewRequest(Message, "127.0.0.1", []byte("------------------"))
	Convey("test unpack packet splicing", t, func() {
		testData, err := req.Pack()
		So(err, ShouldEqual, nil)
		buf := new(bytes.Buffer)
		buf.Write(testData)
		buf.Write(testData)
		buf.Write(testData)

		readerCh := make(chan string, 3)
		// scanner
		scanner := bufio.NewScanner(buf)
		scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			if !atEOF && data[0] == 'i' {
				if len(data) > 8 {
					length := int32(0)
					binary.Read(bytes.NewReader(data[4:8]), binary.BigEndian, &length)
					if int(length)+8 <= len(data) {
						return int(length) + 8, data[:int(length)+8], nil
					}
				}
			}
			return
		})
		for scanner.Scan() {
			scannedPack := new(Request)
			scannedPack.Unpack(bytes.NewReader(scanner.Bytes()))
			readerCh <- fmt.Sprintf("%s", scannedPack)
		}
		if err := scanner.Err(); err != nil {
			log.Fatal("无效数据包")
		}
		i := 0
		for {
			select {
			case str := <-readerCh:
				if len(str) > 0 {
					fmt.Println(str)
					i++
				}
				if i == 3 {
					return
				}
			case <-time.After(1 * time.Second):
				So("timeout", ShouldEqual, "")
				break

			}
		}
	})
}
