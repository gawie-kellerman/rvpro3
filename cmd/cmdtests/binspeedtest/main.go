package main

/*
#include <stdlib.h> // Required to declare C.free
#include <stdio.h>  // Optional: for other C standard library functions
#include <stdint.h>


// Helper function to write a string to a file using fwrite
void writeToFile(FILE* file, const char* text) {
	fwrite(text, 1, 15, file);
}

void writeBuffer(FILE* file, unsigned char* buffer, int bufferLen) {
	fwrite(buffer, 1, bufferLen, file);
}
*/
import "C"
import (
	"bufio"
	"encoding/binary"
	"log"
	"os"
	"syscall"
	"time"
	"unsafe"

	"rvpro3/radarvision.com/utils"
)

func main() {
	writeStructure()
	writeIndividual()
	writeBuffered()
	writeOptimized()
	writeSyscall()
	writeCCall()
}

func writeIndividual() {
	utils.Print.Ln("Individual Binary Writes")
	file, err := os.Create("nonreflective.bin")
	if err != nil {
		log.Fatalf("Couldn't create file: %v", err)
	}
	defer file.Close() // 4. Ensure the file is closed

	start := time.Now()
	for i := 0; i < 100; i++ {
		_ = binary.Write(file, binary.LittleEndian, float64(i))
		_ = binary.Write(file, binary.LittleEndian, uint8(i))
		_ = binary.Write(file, binary.LittleEndian, uint32(i))
		//_, _ = file.WriteString("Hello World")

		err = binary.Write(file, binary.LittleEndian, uint16(i))
		//utils.Debug.Panic(err)
		//utils.Print.Ln("Record", i, "written")
	}
	end := time.Now()
	utils.Print.Ln("Individual Field Writes took", end.Sub(start).Milliseconds())
}

func writeStructure() {
	utils.Print.Ln("Structure Writes")

	var data = struct {
		Pi   float64
		Uate uint8
		Mine [3]byte
		Too  uint16
	}{
		Pi:   3.141592653589793,
		Uate: 255,
		Mine: [3]byte{1, 2, 3},
		Too:  61374,
	}

	// 2. Create the output file.
	file, err := os.Create("reflective.bin")
	if err != nil {
		log.Fatalf("Couldn't create file: %v", err)
	}
	defer file.Close() // 4. Ensure the file is closed

	// 3. Write the data using binary.Write, specifying LittleEndian byte order.

	start := time.Now()
	for n := 0; n < 100; n++ {
		err = binary.Write(file, binary.LittleEndian, data)
		if err != nil {
			log.Fatalf("Write failed: %v", err)
		}
	}
	end := time.Now()
	utils.Print.Ln("Struct Write", end.Sub(start).Milliseconds())
}

func writeBuffered() {
	file, err := os.Create("buffered.bin")
	utils.Debug.Panic(err)
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	start := time.Now()
	for i := 0; i < 100; i++ {
		_ = binary.Write(writer, binary.LittleEndian, float64(i))
		_ = binary.Write(writer, binary.LittleEndian, uint8(i))
		_ = binary.Write(writer, binary.LittleEndian, uint32(i))
		//_, _ = file.WriteString("Hello World")

		err = binary.Write(file, binary.LittleEndian, uint16(i))
		//utils.Debug.Panic(err)
		//utils.Print.Ln("Record", i, "written")
	}
	end := time.Now()
	utils.Print.Ln("Buffered Individual Writer", end.Sub(start).Milliseconds())
}

func writeOptimized() {
	file, err := os.Create("golang-optimized.bin")

	utils.Debug.Panic(err)
	defer file.Close()

	//writer := bufio.NewWriter(file)
	//defer writer.Flush()

	buf := make([]byte, 15)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, _ = file.Write(buf)
		//_ = binary.Write(writer, binary.LittleEndian, float64(i))
		//_ = binary.Write(writer, binary.LittleEndian, uint8(i))
		//_ = binary.Write(writer, binary.LittleEndian, uint32(i))
		//_, _ = file.WriteString("Hello World")
		//
		//err = binary.Write(file, binary.LittleEndian, uint16(i))
		//utils.Debug.Panic(err)
		//utils.Print.Ln("Record", i, "written")
	}
	end := time.Now()
	utils.Print.Ln("Write blank buffer of 15 bytes", end.Sub(start).Milliseconds())
}

func writeSyscall() {
	fd, err := syscall.Open("go-syscall.bin", syscall.O_CREAT|syscall.O_WRONLY|syscall.O_TRUNC, 666)
	utils.Debug.Panic(err)

	defer syscall.Close(fd)

	buf := make([]byte, 15)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, _ = syscall.Write(fd, buf)
	}
	end := time.Now()

	utils.Print.Ln("Write Syscall of 15 bytes", end.Sub(start).Milliseconds())
}

func writeCCall() {
	filename := "ccall.bin"

	// Convert Go strings to C strings
	cFilename := C.CString(filename)
	cAttr := C.CString("w+")

	bytes := make([]byte, 15)

	// Defer the freeing of memory allocated by C.CString to prevent memory leaks
	defer C.free(unsafe.Pointer(cFilename))
	defer C.free(unsafe.Pointer(cAttr))

	// Call the C function and retrieve potential error number

	file := C.fopen(cFilename, cAttr)
	if file == nil {
		log.Fatalf("Couldn't open filename: %v", filename)
	}
	defer C.fclose(file)

	start := time.Now()
	for n := 0; n < 1000; n++ {
		C.writeBuffer(file, (*C.uchar)(unsafe.Pointer(&bytes[0])), C.int(len(bytes)))
		//cContent := C.CString(content)
		//errno_val, err := C.fwrite(cContent, C.uint32_t(len(content)), 1, file)
		//C.writeToFile(file, cContent)
		//defer C.free(unsafe.Pointer(cContent))

		// Check if the call succeeded
	}

	end := time.Now()
	utils.Print.Ln("Write CCall of 15 bytes", end.Sub(start).Milliseconds())
}
