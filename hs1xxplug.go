package hs1xxplug

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"time"
)

type Hs1xxPlug struct {
	IPAddress string
}

type dailyStatsMessage struct {
	Emeter struct {
		GetDaystat struct {
			Month int `json:"month"`
			Year  int `json:"year"`
		} `json:"get_daystat"`
	} `json:"emeter"`
}

type sysInfoMessage struct {
	System struct {
		GetSysinfo struct {
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type meterInfoMessage struct {
	Emeter struct {
		GetRealtime   struct{} `json:"get_realtime"`
		GetVgainIgain struct{} `json:"get_vgain_igain"`
	} `json:"emeter"`
	System struct {
		GetSysinfo struct {
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type setStateMessage struct {
	System struct {
		SetRelayState struct {
			State int `json:"state"`
		} `json:"set_relay_state"`
	} `json:"system"`
}

// SetState sets the state of the plug where true = on and false = off
func (p *Hs1xxPlug) SetState(state bool) (err error) {
	message := setStateMessage{}
	if state {
		message.System.SetRelayState.State = 1
	}
	json, _ := json.Marshal(message)
	_, err = send(p.IPAddress, encrypt(json))
	return
}

func (p *Hs1xxPlug) TurnOn() (err error) {
	err = p.SetState(true)
	return
}

func (p *Hs1xxPlug) TurnOff() (err error) {
	err = p.SetState(false)
	return
}

func (p *Hs1xxPlug) SystemInfo() (results string, err error) {
	message := sysInfoMessage{}
	json, _ := json.Marshal(message)
	reading, err := send(p.IPAddress, encrypt(json))
	if err == nil {
		results = decrypt(reading[4:])
	}
	return
}

func (p *Hs1xxPlug) MeterInfo() (results string, err error) {
	message := meterInfoMessage{}
	json, _ := json.Marshal(message)
	reading, err := send(p.IPAddress, encrypt(json))
	if err == nil {
		results = decrypt(reading[4:])
	}
	return
}

func (p *Hs1xxPlug) DailyStats(month int, year int) (results string, err error) {
	message := dailyStatsMessage{}
	message.Emeter.GetDaystat.Month = month
	message.Emeter.GetDaystat.Year = year
	json, _ := json.Marshal(message)
	reading, err := send(p.IPAddress, encrypt(json))
	if err == nil {
		results = decrypt(reading[4:])
	}
	return
}

func encrypt(plaintext []byte) []byte {
	n := len(plaintext)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(n))
	ciphertext := []byte(buf.Bytes())

	key := byte(0xAB)
	payload := make([]byte, n)
	for i := 0; i < n; i++ {
		payload[i] = plaintext[i] ^ key
		key = payload[i]
	}

	for i := 0; i < len(payload); i++ {
		ciphertext = append(ciphertext, payload[i])
	}

	return ciphertext
}

func decrypt(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 0; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

func send(ip string, payload []byte) (data []byte, err error) {
	// 10 second timeout
	conn, err := net.DialTimeout("tcp", ip+":9999", time.Duration(10)*time.Second)
	if err != nil {
		fmt.Println("Cannot connnect to plug:", err)
		data = nil
		return
	}
	_, err = conn.Write(payload)
	data, err = ioutil.ReadAll(conn)
	if err != nil {
		fmt.Println("Cannot read data from plug:", err)
	}
	return
}
