package main

/*
 * go build -v -o mslib.a -buildmode=c-archive mslib.go 
 */

/*
#include <stdint.h> // for uintptr_t
*/
import "C"
import "runtime/cgo"


import (
	"fmt"
	"os"
	
	"errors"
	"unsafe"

	"github.com/BertoldVdb/ms-tools/gohid"
	"github.com/BertoldVdb/ms-tools/mshal"
	"github.com/sstallion/go-hid"
)

func SearchDevice(foundHandler func(info *hid.DeviceInfo) error) error {
	for _, vid := range []uint16{uint16(0x534d), uint16(0x345f)} {
		if err := hid.Enumerate(vid, uint16(0), func(info *hid.DeviceInfo) error {
			return foundHandler(info)
		}); err != nil {
			return err
		}
	}
	return nil
}

func OpenDevice(name string) (gohid.HIDDevice, error) {
	var dev *hid.Device
	err := SearchDevice(func(info *hid.DeviceInfo) error {
		if info.ProductStr == name {
			d, err := hid.Open(info.VendorID, info.ProductID, info.SerialNbr)
			if err == nil {
				dev = d
				return errors.New("Done")
			}
			return err
		}
		return nil
	})
	if dev != nil {
		return dev, nil
	}
	if err == nil {
		err = os.ErrNotExist
	}

	return nil, err
}

type Context struct {
	dev gohid.HIDDevice
	hal *mshal.HAL
	x string
}

//export MsHalOpen
func MsHalOpen(name *C.char) C.uintptr_t {
	c := &Context{}
	dev, err := OpenDevice(C.GoString(name))
	if err != nil {
		fmt.Println("Failed to open device", err)
		return 0
	}
	c.dev = dev
	config := mshal.HALConfig{
		PatchTryInstall: true,

		PatchProbeEEPROM: true,
		EEPromSize:       0,

		PatchIgnoreUserFirmware: false,

		LogFunc: func(level int, format string, param ...interface{}) {
			if level > 0 {
				return
			}
			str := fmt.Sprintf(format, param...)
			fmt.Printf("HAL(%d): %s\n", level, str)
		},
	}
	c.hal, err = mshal.New(dev, config)
	if err != nil {
		fmt.Println("Failed to create HAL", err)
		dev.Close()
		return 0
	}
	h := cgo.NewHandle(c)
	return C.uintptr_t(h)
}

//export MsHalClose
func MsHalClose(handle C.uintptr_t) {
	h := cgo.Handle(handle)
	ctx := h.Value().(*Context)
	ctx.dev.Close()
	h.Delete()
}

//export MsHalI2CTransfer
func MsHalI2CTransfer(handle C.uintptr_t, addr C.int, wrData unsafe.Pointer, wrLen C.int, rdLen C.int, rdData *unsafe.Pointer) C.int {
	h := cgo.Handle(handle)
	ctx := h.Value().(*Context)
	wrBuf := C.GoBytes(wrData, wrLen)
	rdBuf := make([]byte, rdLen)
	ok, err := ctx.hal.I2CTransfer(byte(addr), wrBuf, rdBuf)
	if err != nil {
		return 1
	}
	if !ok {
		return 2
	}
	*rdData = C.CBytes(rdBuf)
	
	return 0
}

//export MsHalMemAccess
func MsHalMemAccess(handle C.uintptr_t, write C.int, addr C.int, data unsafe.Pointer, length C.int) C.int {
	h := cgo.Handle(handle)
	ctx := h.Value().(*Context)
	region := ctx.hal.MemoryRegionGet(mshal.MemoryRegionRAM)
	buf := C.GoBytes(data, length)
	_, err := region.Access(write == 1, int(addr), buf)
	if err != nil {
		return 1
	}
	if(write == 0) {
		udata := uintptr(data)
		for i := 0; i < int(length); i++ {
			datac := (*C.char)(unsafe.Pointer(udata))
			*datac = C.char(buf[i])
			udata++
		}
	}
	
	return 0
}

func main() {
	// We need the main function to make possible
	// CGO compiler to compile the package as C shared library
}
