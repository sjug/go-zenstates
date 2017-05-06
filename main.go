package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
)

var listFlag bool
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

func initFlags() {
	flag.BoolVar(&listFlag, "l", false, "List all pstates")
	flag.IntVar(&pstateFlag, "p", -1, "pstate to set")
	flag.Uint64Var(&fidFlag, "fid", 0, "FID to set (hex)")
	flag.Uint64Var(&didFlag, "did", 0, "DID to set (hex)")
	flag.Uint64Var(&vidFlag, "vid", 0, "VID to set (hex)")
	flag.Parse()
}

func main() {
	initFlags()
	//newFID := uint64(0xA0)
	//newMSR := (msrValue &^ 0xff) + newFID
	//newDID := uint64(0x8)
	//newMSR2 := (newMSR &^ (0x3f << 8)) + (newDID << 8)
	//fmt.Println(pstateToString(newMSR2))
	//newVID := uint64(0x20)
	//newMSR3 := (newMSR2 &^ (0xff << 14)) + (newVID << 14)
	//fmt.Println(pstateToString(newMSR3))
	if listFlag {
		for _, pstate := range pstates {
			fmt.Println(pstateToString(readMSR(pstate)))
		}
	}
	if pstateFlag >= 0 {
		msrValue := readMSR(pstates[0])
		newMSR := msrValue
		fmt.Println("Current pstate" + strconv.Itoa(pstateFlag) + ": " + pstateToString(msrValue))
		if fidFlag > 0 {
			newMSR = setFid(msrValue, fidFlag)
			fmt.Printf("Setting FID to %X\n", fidFlag)
		}
		if didFlag > 0 {
			newMSR = setDid(msrValue, didFlag)
			fmt.Printf("Setting DID to %X\n", didFlag)
		}
		if vidFlag > 0 {
			newMSR = setVid(msrValue, vidFlag)
			fmt.Printf("Setting VID to %X\n", vidFlag)
		}
		if newMSR != msrValue {
			writeMSR(pstates[pstateFlag], newMSR)
			fmt.Println("New pstate" + strconv.Itoa(pstateFlag) + ": " + pstateToString(readMSR(pstates[pstateFlag])))
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
	fmt.Println(value)
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
		fmt.Println(byteValue)
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
