package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
)

var listFlag, pstateEnable, pstateDisable bool
var pstateFlag int
var fidFlag, didFlag, vidFlag uint64
var pstates = [...]int64{
	0xC0010064,
	0xC0010065,
	0xC0010066,
	0xC0010067,
	0xC0010068,
	0xC0010069,
	0xC001006A,
	0xC001006B,
}
var tscLock int64 = 0xC0010015

func initFlags() {
	flag.BoolVar(&listFlag, "l", false, "List all pstates")
	flag.BoolVar(&pstateEnable, "enable", false, "Enable pstate")
	flag.BoolVar(&pstateDisable, "disable", false, "Disable pstate")
	flag.IntVar(&pstateFlag, "p", -1, "pstate to set")
	flag.Uint64Var(&fidFlag, "fid", 0, "FID to set (hex)")
	flag.Uint64Var(&didFlag, "did", 0, "DID to set (hex)")
	flag.Uint64Var(&vidFlag, "vid", 0, "VID to set (hex)")
	flag.Parse()
}

func main() {
	initFlags()
	if listFlag {
		for _, pstate := range pstates {
			fmt.Println(pstateToString(readMSR(pstate)))
		}
	}
	if pstateFlag >= 0 && pstateFlag < 8 {
		msrValue := readMSR(pstates[0])
		newMSR := msrValue
		fmt.Printf("Current pstate%d: %s\n", pstateFlag, pstateToString(msrValue))
		if pstateEnable {
			newMSR = setBit(newMSR, 63)
			fmt.Printf("Enabled pstate%d", pstateFlag)
		}
		if pstateDisable {
			newMSR = clearBit(newMSR, 63)
			fmt.Printf("Disabled pstate%d", pstateFlag)
		}
		if fidFlag > 0 {
			newMSR = setFid(newMSR, fidFlag)
			fmt.Printf("Setting FID to %X\n", fidFlag)
		}
		if didFlag > 0 {
			newMSR = setDid(newMSR, didFlag)
			fmt.Printf("Setting DID to %X\n", didFlag)
		}
		if vidFlag > 0 {
			newMSR = setVid(newMSR, vidFlag)
			fmt.Printf("Setting VID to %X\n", vidFlag)
		}
		if newMSR != msrValue {
			tscValue := readMSR(tscLock)
			if !hasBit(tscValue, 21) {
				writeMSR(tscLock, setBit(tscValue, 21))
				fmt.Println("Locking TSC frequency")
			}
			writeMSR(pstates[pstateFlag], newMSR)
			fmt.Printf("New pstate%d: %v\n", pstateFlag, pstateToString(readMSR(pstates[pstateFlag])))
		}
	}
}

func setFid(msr uint64, fid uint64) uint64 {
	return (msr &^ 0xff) + fid
}

func setDid(msr uint64, did uint64) uint64 {
	return (msr &^ (0x3f << 8)) + (did << 8)
}

func setVid(msr uint64, vid uint64) uint64 {
	return (msr &^ (0xff << 14)) + (vid << 14)
}

func readMSR(msr int64) uint64 {
	f, err := os.Open("/dev/cpu/0/msr")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	value := make([]byte, 8)
	_, err = f.ReadAt(value, msr)
	if err != nil {
		panic(err)
	}
	//fmt.Println(value)
	return binary.LittleEndian.Uint64(value)
}

func writeMSR(msr int64, value uint64) {
	for i := 0; i < 16; i++ {
		f, err := os.OpenFile("/dev/cpu/"+strconv.Itoa(i)+"/msr", os.O_WRONLY, os.ModeCharDevice)
		if err != nil {
			panic(err)
		}
		byteValue := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteValue, value)
		//fmt.Println(byteValue)
		_, err = f.WriteAt(byteValue, msr)
		if err != nil {
			panic(err)
		}
	}
}

func pstateToString(value uint64) string {
	// Check if pstate is enabled
	if hasBit(value, 63) {
		// First 8 bits are core frequency ID
		fid := value & 0xff
		// Shift and get next 6 bits for core divisor ID
		did := (value >> 8) & 0x3f
		// Shift and get last 8 bits for core voltage ID
		vid := (value >> 14) & 0xff
		// Calculate human readable ratio and vcore
		ratio := 25 * float64(fid) / (12.5 * float64(did))
		vcore := 1.55 - (0.00625 * float64(vid))
		return fmt.Sprintf("Enabled - FID = %X - DID = %X - VID = %X - Ratio = %.2f - vCore = %.5f", fid, did, vid, ratio, vcore)
	}
	return "Disabled"
}

func hasBit(value uint64, pos uint) bool {
	return ((value & (1 << pos)) > 0)
}

func setBit(value uint64, pos uint) uint64 {
	return value | (1 << pos)
}

func clearBit(value uint64, pos uint) uint64 {
	return value &^ (1 << pos)
}
