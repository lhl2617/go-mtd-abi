package mtdabi

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

const kernelVersion = "5.12.8-arch1-1"

/*
This simulated MTD (32MiB, 512 bytes page NAND flash) is created by the command
```
modprobe nandsim first_id_byte=0x20 second_id_byte=0x35
```

This MTD cannot be locked, and has no different erase regions.
*/
const mtdPath = "/dev/mtd0"
const procMtdContents = `dev:    size   erasesize  name
mtd0: 02000000 00004000 "NAND simulator partition 0"
`
const regionCount = 0

var nandOobinfo = unix.NandOobinfo{
	Useecc:   0x2,
	Eccbytes: 0x6,
	Oobfree: [8][2]uint32{
		{0x8, 0x8},
		{0x0, 0x0},
		{0x0, 0x0},
		{0x0, 0x0},
		{0x0, 0x0},
		{0x0, 0x0},
		{0x0, 0x0},
		{0x0, 0x0},
	},
	Eccpos: [32]uint32{
		0x0, 0x1, 0x2, 0x3, 0x6, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	},
}
var nandEcclayout = unix.NandEcclayout{
	Eccbytes: 0x6,
	Eccpos:   [64]uint32{0x0, 0x1, 0x2, 0x3, 0x6, 0x7, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
	Oobavail: 0x8,
	Oobfree: [8]unix.NandOobfree{
		{Offset: 0x8, Length: 0x8},
		{Offset: 0x0, Length: 0x0},
		{Offset: 0x0, Length: 0x0},
		{Offset: 0x0, Length: 0x0},
		{Offset: 0x0, Length: 0x0},
		{Offset: 0x0, Length: 0x0},
		{Offset: 0x0, Length: 0x0},
		{Offset: 0x0, Length: 0x0},
	},
}

// There should not be any ecc happening through the simulated NAND
var mtdEccStats = unix.MtdEccStats{
	Corrected: 0x0,
	Failed:    0x0,
	Badblocks: 0x0,
	Bbtblocks: 0x0,
}

// By right, these linux errors from <errno.h> should not be seen by user programs.
var (
	errENOTSUPP error = unix.Errno(524)
)

var mtdInfo = unix.MtdInfo{
	Type:      0x4,
	Flags:     0x400,
	Size:      0x2000000,
	Erasesize: 0x4000,
	Writesize: 0x200,
	Oobsize:   0x10,
}

func TestMain(m *testing.M) {
	err := checkKernelVersionMatch()
	if err != nil {
		log.Fatal(err)
	}
	err = checkPwdIsVagrant()
	if err != nil {
		log.Fatal(err)
	}
	err = setupNandsim()
	if err != nil {
		log.Fatal(err)
	}
	defer teardownNandsim()
	os.Exit(m.Run())
}

func checkKernelVersionMatch() error {
	gotKernelVersionBuf, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("Can't run 'uname -r': %v", err))
	}
	gotKernelVersion := strings.TrimSpace(string(gotKernelVersionBuf))
	if kernelVersion != gotKernelVersion {
		return errors.New(fmt.Sprintf("Kernel version: want '%v' got '%v'\n", kernelVersion, gotKernelVersion) +
			"Please use the Vagrant VM to run tests!\n" +
			"See the README for more info.")
	}
	return nil
}

func checkPwdIsVagrant() error {
	dir, err := os.Getwd()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get working directory: %v", err))
	}
	if dir != "/vagrant" {
		return errors.New(fmt.Sprintf("Working directory is not '/vagrant\n'") +
			"Please use the Vagrant VM to run tests!\n" +
			"See the README for more info.")
	}
	return nil
}

func setupNandsim() error {
	_, err := exec.Command("modprobe", "nandsim", "first_id_byte=0x20", "second_id_byte=0x35").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("modprobe command failed: %v", err))
	}
	buf, err := ioutil.ReadFile("/proc/mtd")
	if err != nil {
		return errors.New(fmt.Sprintf("Can't read from '/proc/mtd': %v", err))
	}
	gotProcMtdContents := string(buf)
	if procMtdContents != gotProcMtdContents {
		return errors.New("nandsim not set up properly!\n" +
			fmt.Sprintf("/proc/mtd: want '%v'\ngot '%v'", procMtdContents, gotProcMtdContents))
	}
	return nil
}

func teardownNandsim() error {
	_, err := exec.Command("modprobe", "-r", "nandsim").Output()
	if err != nil {
		return errors.New(fmt.Sprintf("modprobe command failed: %v", err))
	}
	return nil
}

func getMTDFile() (*os.File, error) {
	file, err := os.OpenFile(mtdPath, os.O_SYNC|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func allErased(s []byte) bool {
	for _, v := range s {
		if v != 255 {
			return false
		}
	}
	return true
}

func genRandomBytes(size int) (blk []byte, err error) {
	blk = make([]byte, size)
	_, err = rand.Read(blk)
	return
}

// Tests MemErase
func eraseAndCheckMtd(fd uintptr) error {
	// First erase the whole MTD
	eraseInfo := unix.EraseInfo{
		Start:  0,
		Length: mtdInfo.Size,
	}
	err := MemErase(fd, &eraseInfo)
	if err != nil {
		return errors.New(fmt.Sprintf("MemErase failed: %v", err))
	}
	// Now read the whole MTD and check that everything is erased
	mtdBuf, err := ioutil.ReadFile(mtdPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to read '%v': %v", mtdPath, err))
	}
	if !allErased(mtdBuf) {
		return errors.New("MemErase did not erase all bytes on MTD!")
	}
	return nil
}

// Tests MemGetInfo
func TestMemGetInfo(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	want := mtdInfo
	var got unix.MtdInfo
	err = MemGetInfo(fd, &got)
	if err != nil {
		t.Fatalf("MemGetInfo failed: %v", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want '%#v' got '%#v", want, got)
	}

}

// Test MemErase, MemErase64
func TestPwriteMTD(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}

	// Test write using Pwrite
	writeData, err := genRandomBytes(int(mtdInfo.Erasesize))
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	_, err = unix.Pwrite(int(fd), writeData, int64(mtdInfo.Erasesize))
	if err != nil {
		t.Fatalf("Failed to write to mtd: %v", err)
	}
	// Now read the written part of the MTD and make sure it is correct
	mtdBuf, err := ioutil.ReadFile(mtdPath)
	if err != nil {
		t.Fatalf("Failed to read '%v': %v", mtdPath, err)
	}
	regionOfInterest := mtdBuf[mtdInfo.Erasesize : mtdInfo.Erasesize*2]
	if !bytes.Equal(regionOfInterest, writeData) {
		t.Fatalf("Write failed: want '%v' got '%v'", writeData, regionOfInterest)
	}

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}
}

// Tests MemWrite
func TestMemWrite(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}

	// Test write using MemWrite
	writeData, err := genRandomBytes(int(mtdInfo.Erasesize) * 2)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	writeReq := unix.MtdWriteReq{
		Start: uint64(mtdInfo.Erasesize) * 2,
		Len:   uint64(mtdInfo.Erasesize) * 2,
		Data:  uint64(uintptr(unsafe.Pointer(&writeData[0]))),
	}
	err = MemWrite(fd, &writeReq)
	if err != nil {
		t.Fatalf("MemWrite failed")
	}
	// Now read the written part of the MTD and make sure it is correct
	mtdBuf, err := ioutil.ReadFile(mtdPath)
	if err != nil {
		t.Fatalf("Failed to read '%v': %v", mtdPath, err)
	}
	regionOfInterest := mtdBuf[mtdInfo.Erasesize*2 : mtdInfo.Erasesize*4]
	if !bytes.Equal(regionOfInterest, writeData) {
		t.Fatalf("Write failed: want '%v' got '%v'", writeData, regionOfInterest)
	}

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}
}

// Tests MemReadOob, MemWriteOob
func TestReadWriteOob(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}

	// Check that OOB region is all erased
	bufSize := uint32(16)
	buf := make([]byte, bufSize)
	mtdOobBuf := unix.MtdOobBuf{
		Start:  0,
		Length: bufSize,
		Ptr:    &buf[0],
	}
	err = MemReadOob(fd, &mtdOobBuf)
	if err != nil {
		t.Fatalf("MemReadOob failed: %v", err)
	}
	if !allErased(buf) {
		t.Fatalf("Oob: want all erased, got '%v'", err)
	}

	// Write some junk to it
	buf, err = genRandomBytes(int(bufSize))
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	mtdOobBuf = unix.MtdOobBuf{
		Start:  0,
		Length: bufSize,
		Ptr:    &buf[0],
	}
	err = MemWriteOob(fd, &mtdOobBuf)
	if err != nil {
		t.Fatalf("MemWriteOob failed: %v", err)
	}

	// Now we read the region back to see if it's correct
	readBuf := make([]byte, bufSize)
	mtdOobBuf = unix.MtdOobBuf{
		Start:  0,
		Length: bufSize,
		Ptr:    &readBuf[0],
	}
	err = MemReadOob(fd, &mtdOobBuf)
	if err != nil {
		t.Fatalf("MemReadOob failed: %v", err)
	}
	if !bytes.Equal(readBuf, buf) {
		t.Fatalf("Oob: want '%v', got '%v'", buf, readBuf)
	}

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}
}

// Tests MemReadOob64, MemWriteOob64
func TestReadWriteOob64(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}

	// Check that OOB region is all erased
	bufSize := uint32(16)
	buf := make([]byte, bufSize)
	mtdOobBuf := unix.MtdOobBuf64{
		Start:  0,
		Length: bufSize,
		Ptr:    uint64(uintptr(unsafe.Pointer(&buf))),
	}
	err = MemReadOob64(fd, &mtdOobBuf)
	if err != nil {
		t.Fatalf("MemReadOob64 failed: %v", err)
	}
	if !allErased(buf) {
		t.Fatalf("Oob: want all erased, got '%v'", err)
	}

	// Write some junk to it
	buf, err = genRandomBytes(int(bufSize))
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	mtdOobBuf = unix.MtdOobBuf64{
		Start:  0,
		Length: bufSize,
		Ptr:    uint64(uintptr(unsafe.Pointer(&buf))),
	}
	err = MemWriteOob64(fd, &mtdOobBuf)
	if err != nil {
		t.Fatalf("MemWriteOob64 failed: %v", err)
	}

	// Now we read the region back to see if it's correct
	readBuf := make([]byte, bufSize)
	mtdOobBuf = unix.MtdOobBuf64{
		Start:  0,
		Length: bufSize,
		Ptr:    uint64(uintptr(unsafe.Pointer(&readBuf))),
	}
	err = MemReadOob64(fd, &mtdOobBuf)
	if err != nil {
		t.Fatalf("MemReadOob64 failed: %v", err)
	}
	if !bytes.Equal(readBuf, buf) {
		t.Fatalf("Oob: want '%v', got '%v'", buf, readBuf)
	}

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}
}

// Tests MemIsLock, MemLock, MemUnlock
func TestLock(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	eraseInfo := unix.EraseInfo{
		Start:  0,
		Length: mtdInfo.Size,
	}
	err = MemIsLocked(fd, &eraseInfo)
	if err != unix.EOPNOTSUPP {
		t.Errorf("MemIsLocked err: want '%v' got '%v'", unix.EOPNOTSUPP, err)
	}
	err = MemLock(fd, &eraseInfo)
	if err != errENOTSUPP {
		t.Errorf("MemLock err: want '%v' got '%v'", errENOTSUPP, err)
	}
	err = MemUnlock(fd, &eraseInfo)
	if err != errENOTSUPP {
		t.Errorf("MemUnlock err: want '%v' got '%v'", errENOTSUPP, err)
	}
}

// Tests MemGetRegionCount, MemGetRegionInfo
func TestMemGetRegion(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	var gotRegionCount int32
	err = MemGetRegionCount(fd, &gotRegionCount)
	if err != nil {
		t.Fatalf("MemGetRegionCount failed: %v", err)
	}
	if regionCount != gotRegionCount {
		t.Fatalf("region count: want '%v' got '%v'", regionCount, gotRegionCount)
	}

	// Because this MTD has regionCount == 0, this will fail with invalid argument
	gotRegionInfo := unix.RegionInfo{
		Regionindex: 0,
	}
	err = MemGetRegionInfo(fd, &gotRegionInfo)
	if err != unix.EINVAL {
		t.Fatalf("MemGetRegionInfo err: want '%v' got '%v'", unix.EINVAL, err)
	}
}

func TestMemErase(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}

	// Write over 3 erase blocks from the front
	writeData, err := genRandomBytes(int(mtdInfo.Erasesize) * 3)
	if err != nil {
		t.Fatalf("Failed to generate random bytes: %v", err)
	}
	_, err = unix.Pwrite(int(fd), writeData, 0)
	if err != nil {
		t.Fatalf("Failed to write to mtd: %v", err)
	}
	// Now read the written part of the MTD and make sure it is correct
	mtdBuf, err := ioutil.ReadFile(mtdPath)
	if err != nil {
		t.Fatalf("Failed to read '%v': %v", mtdPath, err)
	}
	regionOfInterest := mtdBuf[:mtdInfo.Erasesize*3]
	if !bytes.Equal(regionOfInterest, writeData) {
		t.Fatalf("Write failed")
	}

	// Now use erase(32) to erase the first block and ensure it is erased
	eraseInfo := unix.EraseInfo{
		Start:  0,
		Length: mtdInfo.Erasesize,
	}
	err = MemErase(fd, &eraseInfo)
	if err != nil {
		t.Fatalf("MemErase failed: %v", err)
	}
	// Now read the written part of the MTD and make sure it is correct
	mtdBuf, err = ioutil.ReadFile(mtdPath)
	if err != nil {
		t.Fatalf("Failed to read '%v': %v", mtdPath, err)
	}
	if !allErased(mtdBuf[:mtdInfo.Erasesize]) {
		t.Fatalf("Region to erase not erased properly by MemErase")
	}
	if !bytes.Equal(mtdBuf[mtdInfo.Erasesize:mtdInfo.Erasesize*3], writeData[mtdInfo.Erasesize:]) {
		t.Fatalf("MemErase over-erased into other regions")
	}

	// Now use erase64 to erase the second block and ensure it is erased
	eraseInfo64 := unix.EraseInfo64{
		Start:  uint64(mtdInfo.Erasesize),
		Length: uint64(mtdInfo.Erasesize),
	}
	err = MemErase64(fd, &eraseInfo64)
	if err != nil {
		t.Fatalf("MemErase64 failed: %v", err)
	}
	// Now read the written part of the MTD and make sure it is correct
	mtdBuf, err = ioutil.ReadFile(mtdPath)
	if err != nil {
		t.Fatalf("Failed to read '%v': %v", mtdPath, err)
	}
	if !allErased(mtdBuf[:mtdInfo.Erasesize*2]) {
		t.Fatalf("Region to erase not erased properly by MemErase64")
	}
	if !bytes.Equal(mtdBuf[mtdInfo.Erasesize*2:mtdInfo.Erasesize*3], writeData[mtdInfo.Erasesize*2:]) {
		t.Fatalf("MemErase64 over-erased into other regions")
	}

	err = eraseAndCheckMtd(fd)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemGetOobSel(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	got := unix.NandOobinfo{}
	err = MemGetOobSel(fd, &got)
	if err != nil {
		t.Fatalf("MemGetOobSel failed: %v", err)
	}
	if !reflect.DeepEqual(nandOobinfo, got) {
		t.Errorf("OobSel: want '%v' got '%v'", nandOobinfo, got)
	}
}

// Tests MemGetBadBlock and MemSetBadBlock
func TestMemBadBlock(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	// DO NOT defer closing it here, because we're tearing down and setting up after
	// setting a bad block
	// defer file.Close()
	fd := file.Fd()

	blockNum := int64(mtdInfo.Erasesize)
	err = MemGetBadBlock(fd, &blockNum)
	if err != nil {
		// This should be fine
		t.Fatalf("MemGetBadBlock failed: %v", err)
	}
	err = MemSetBadBlock(fd, &blockNum)
	if err != nil {
		t.Fatalf("MemSetBadBlock failed: %v", err)
	}
	// Now this block is corrupted

	// Get again
	err = MemGetBadBlock(fd, &blockNum)
	if err != nil {
		// For some reason, this is fine
		t.Fatalf("MemGetBadBlock failed: %v", err)
	}

	eraseInfo := unix.EraseInfo{
		Start:  0,
		Length: mtdInfo.Size,
	}
	err = MemErase(fd, &eraseInfo)
	if err != unix.EIO {
		t.Fatalf("Erase error: want '%v' got '%v'", unix.EIO, err)
	}
	// close the mtd file
	file.Close()
	// tear it down and set it up again
	err = teardownNandsim()
	if err != nil {
		t.Fatalf("Failed to teardown nandsim: %v", err)
	}
	err = setupNandsim()
	if err != nil {
		t.Fatalf("Failed to setup nandsim: %v", err)
	}
}

// Tests EccGetLayout, EccGetStats
func TestEcc(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	gotNandEcclayout := unix.NandEcclayout{}
	err = EccGetLayout(fd, &gotNandEcclayout)
	if err != nil {
		t.Fatalf("EccGetLayout failed: %v", err)
	}
	if !reflect.DeepEqual(nandEcclayout, gotNandEcclayout) {
		t.Fatalf("NandEcclayout: want '%v' got '%v'", nandEcclayout, gotNandEcclayout)
	}

	gotEccStats := unix.MtdEccStats{}
	err = EccGetStats(fd, &gotEccStats)
	if err != nil {
		t.Fatalf("EccGetStats failed: %v", err)
	}
	if !reflect.DeepEqual(gotEccStats, mtdEccStats) {
		t.Errorf("mtdEccStats: want '%v' got '%v'", mtdEccStats, gotEccStats)
	}
}

// Tests OtpSelect, OtpGetRegionCount, OtpGetRegionInfo, OtpLock
// Because this is not an OTP flash, everything should fail with EINVAL or EOPNOTSUPP
func TestOtp(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	otpMode := int32(unix.MTD_OTP_USER)
	err = OtpSelect(fd, &otpMode)
	if err != unix.EOPNOTSUPP {
		t.Fatalf("OtpSelect err: want '%v' got '%v'", unix.EOPNOTSUPP, err)
	}

	gotOtpRegionCount := int32(0)
	err = OtpGetRegionCount(fd, &gotOtpRegionCount)
	if err != unix.EINVAL {
		t.Fatalf("OtpGetRegionCount err: want '%v' got '%v'", unix.EINVAL, err)
	}

	gotOtpInfo := unix.OtpInfo{}
	err = OtpGetRegionInfo(fd, &gotOtpInfo)
	if err != unix.EINVAL {
		t.Fatalf("OtpGetRegionInfo err: want '%v' got '%v'", unix.EINVAL, err)
	}

	gotLockOtpInfo := unix.OtpInfo{}
	err = OtpLock(fd, &gotLockOtpInfo)
	if err != unix.EINVAL {
		t.Fatalf("OtpLock err: want '%v' got '%v'", unix.EINVAL, err)
	}
}

func TestMtdFileMode(t *testing.T) {
	file, err := getMTDFile()
	if err != nil {
		t.Fatalf("Failed to get MTD fd: %v", err)
	}
	defer file.Close()
	fd := file.Fd()

	err = MtdFileMode(fd, uintptr(unix.MTD_FILE_MODE_NORMAL))
	if err != nil {
		t.Fatalf("MtdFileMode failed: %v", err)
	}
}
