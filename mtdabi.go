// Package mtdabi is a Golang implementation of helper functions for
// the `ioctl` calls in the Linux MTD ABI found at
// https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/uapi/mtd/mtd-abi.h.
//
// This package is currently based on version `v5.12` of the Linux kernel.
package mtdabi

import (
	"unsafe"

	"golang.org/x/sys/unix"
)

// MemGetInfo gets MTD characteristics info (better to use sysfs)
//
// #define MEMGETINFO _IOR('M', 1, struct mtd_info_user)
func MemGetInfo(fd uintptr, value *unix.MtdInfo) error {
	return ioctl(fd, unix.MEMGETINFO, uintptr(unsafe.Pointer(value)))
}

// MemErase erases segment of MTD
//
// #define MEMERASE	_IOW('M', 2, struct erase_info_user)
func MemErase(fd uintptr, value *unix.EraseInfo) error {
	return ioctl(fd, unix.MEMERASE, uintptr(unsafe.Pointer(value)))
}

// MemWriteOob writes out-of-band data from MTD
//
// #define MEMWRITEOOB _IOWR('M', 3, struct mtd_oob_buf)
func MemWriteOob(fd uintptr, value *unix.MtdOobBuf) error {
	return ioctl(fd, unix.MEMWRITEOOB, uintptr(unsafe.Pointer(value)))
}

// MemReadOob reads out-of-band data from MTD
//
// #define MEMREADOOB _IOWR('M', 4, struct mtd_oob_buf)
func MemReadOob(fd uintptr, value *unix.MtdOobBuf) error {
	return ioctl(fd, unix.MEMREADOOB, uintptr(unsafe.Pointer(value)))
}

// MemLock locks a chip (for MTD that supports it)
//
// #define MEMLOCK _IOW('M', 5, struct erase_info_user)
func MemLock(fd uintptr, value *unix.EraseInfo) error {
	return ioctl(fd, unix.MEMLOCK, uintptr(unsafe.Pointer(value)))
}

// MemUnlock unlocks a chip (for MTD that supports it)
//
// #define MEMUNLOCK _IOW('M', 6, struct erase_info_user)
func MemUnlock(fd uintptr, value *unix.EraseInfo) error {
	return ioctl(fd, unix.MEMUNLOCK, uintptr(unsafe.Pointer(value)))
}

// MemGetRegionCount gets the number of different erase regions
//
// #define MEMGETREGIONCOUNT _IOR('M', 7, int)
func MemGetRegionCount(fd uintptr, value *int32) error {
	return ioctl(fd, unix.MEMGETREGIONCOUNT, uintptr(unsafe.Pointer(value)))
}

// MemGetRegionInfo gets information about the erase region for a specific index
//
// #define MEMGETREGIONINFO	_IOWR('M', 8, struct region_info_user)
func MemGetRegionInfo(fd uintptr, value *unix.RegionInfo) error {
	return ioctl(fd, unix.MEMGETREGIONINFO, uintptr(unsafe.Pointer(value)))
}

// MemGetOobSel gets info about OOB modes (e.g., RAW, PLACE, AUTO) - legacy interface
//
// #define MEMGETOOBSEL	_IOR('M', 10, struct nand_oobinfo)
func MemGetOobSel(fd uintptr, value *unix.NandOobinfo) error {
	return ioctl(fd, unix.MEMGETOOBSEL, uintptr(unsafe.Pointer(value)))
}

// MemGetBadBlock checks if an eraseblock is bad
//
// #define MEMGETBADBLOCK _IOW('M', 11, __kernel_loff_t)
func MemGetBadBlock(fd uintptr, value *int64) error {
	return ioctl(fd, unix.MEMGETBADBLOCK, uintptr(unsafe.Pointer(value)))
}

// MemSetBadBlock marks an eraseblock as bad
//
// #define MEMSETBADBLOCK _IOW('M', 12, __kernel_loff_t)
func MemSetBadBlock(fd uintptr, value *int64) error {
	return ioctl(fd, unix.MEMSETBADBLOCK, uintptr(unsafe.Pointer(value)))
}

// OtpSelect sets OTP (One-Time Programmable) mode (factory vs. user)
//
// #define OTPSELECT _IOR('M', 13, int)
func OtpSelect(fd uintptr, value *int32) error {
	return ioctl(fd, unix.OTPSELECT, uintptr(unsafe.Pointer(value)))
}

// OtpGetRegionCount gets number of OTP (One-Time Programmable) regions
//
// #define OTPGETREGIONCOUNT	_IOW('M', 14, int)
func OtpGetRegionCount(fd uintptr, value *int32) error {
	return ioctl(fd, unix.OTPGETREGIONCOUNT, uintptr(unsafe.Pointer(value)))
}

// OtpGetRegionInfo gets all OTP (One-Time Programmable) info about MTD
//
// #define OTPGETREGIONINFO	_IOW('M', 15, struct otp_info)
func OtpGetRegionInfo(fd uintptr, value *unix.OtpInfo) error {
	return ioctl(fd, unix.OTPGETREGIONINFO, uintptr(unsafe.Pointer(value)))
}

// OtpLock locks a given range of user data (must be in mode %MTD_FILE_MODE_OTP_USER)
//
// #define OTPLOCK _IOR('M', 16, struct otp_info)
func OtpLock(fd uintptr, value *unix.OtpInfo) error {
	return ioctl(fd, unix.OTPLOCK, uintptr(unsafe.Pointer(value)))
}

// EccGetLayout gets ECC layout (deprecated)
//
// #define ECCGETLAYOUT _IOR('M', 17, struct nand_ecclayout_user)
func EccGetLayout(fd uintptr, value *unix.NandEcclayout) error {
	return ioctl(fd, unix.ECCGETLAYOUT, uintptr(unsafe.Pointer(value)))
}

// EccGetStats gets statistics about corrected/uncorrected errors
//
// #define ECCGETSTATS		_IOR('M', 18, struct mtd_ecc_stats)
func EccGetStats(fd uintptr, value *unix.MtdEccStats) error {
	return ioctl(fd, unix.ECCGETSTATS, uintptr(unsafe.Pointer(value)))
}

// MtdFileMode sets MTD mode on a per-file-descriptor basis (see "MTD file modes")
//
// #define MTDFILEMODE _IO('M', 19)
func MtdFileMode(fd, value uintptr) error {
	return ioctl(fd, unix.MTDFILEMODE, value)
}

// MemErase64 erases segment of MTD (supports 64-bit address)
//
// #define MEMERASE64 _IOW('M', 20, struct erase_info_user64)
func MemErase64(fd uintptr, value *unix.EraseInfo64) error {
	return ioctl(fd, unix.MEMERASE64, uintptr(unsafe.Pointer(value)))
}

// MemWriteOob64 writes data to OOB (64-bit version)
//
// #define MEMWRITEOOB64 _IOWR('M', 21, struct mtd_oob_buf64)
func MemWriteOob64(fd uintptr, value *unix.MtdOobBuf64) error {
	return ioctl(fd, unix.MEMWRITEOOB64, uintptr(unsafe.Pointer(value)))
}

// MemReadOob64 reads data from OOB (64-bit version)
//
// #define MEMREADOOB64 _IOWR('M', 22, struct mtd_oob_buf64)
func MemReadOob64(fd uintptr, value *unix.MtdOobBuf64) error {
	return ioctl(fd, unix.MEMREADOOB64, uintptr(unsafe.Pointer(value)))
}

// MemIsLocked checks if chip is locked (for MTD that supports it)
//
// #define MEMISLOCKED _IOR('M', 23, struct erase_info_user)
func MemIsLocked(fd uintptr, value *unix.EraseInfo) error {
	return ioctl(fd, unix.MEMISLOCKED, uintptr(unsafe.Pointer(value)))
}

// MemWrite is the most generic write interface; can write in-band and/or out-of-band in various
// modes (see "struct mtd_write_req"). This ioctl is not supported for flashes
// without OOB, e.g., NOR flash.
//
// #define MEMWRITE _IOWR('M', 24, struct mtd_write_req)
func MemWrite(fd uintptr, value *unix.MtdWriteReq) error {
	return ioctl(fd, unix.MEMWRITE, uintptr(unsafe.Pointer(value)))
}
