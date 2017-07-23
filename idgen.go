package idgen

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/OneOfOne/xxhash"
)

// Return a byte array containing the first non-loopback MAC address
// on an interface of this machine.
func GetMACAddress() (net.HardwareAddr, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Name != "lo" {
			return iface.HardwareAddr, nil
		}
	}
	return nil, fmt.Errorf("Could not find a MAC address to use in unique IDs No suitable interface on this machine.")
}

// Push new IDs onto a provided channel. A mac address should be
// provided to ensure that resulting IDs are unique to a machine/node,
// and a routineID should be provided to ensure that results form
// concurrent goroutines are unique.
func GenerateIDs(mac net.HardwareAddr, routineID int, resultChan chan []byte) {
	var hash *xxhash.XXHash64
	var uniquePart *bytes.Buffer
	var output *bytes.Buffer

	uniquePart = bytes.NewBuffer(make([]byte, 28, 28))
	hash = xxhash.New64()

	go func() {
		for {
			// Reset instead of reallocating, this is
			// faster, and we have a block of memory
			// assigned per generator, which is
			// synchronized to an unbuffered channel,
			// therefore we shouldn't have any reentrance
			// problems.
			uniquePart.Reset()

			// This is timestamp obviously vulnerable to
			// individual CPU cores getting circa 1200
			// times faster than my i7-5600u is - in that
			// event we can add a per-goroutine counter
			// (uint64) to uniquify sub-nano second
			// results from a single goroutine.
			timestamp := time.Now().UnixNano()

			binary.Write(uniquePart, binary.LittleEndian, timestamp)
			binary.Write(uniquePart, binary.LittleEndian, mac)
			binary.Write(uniquePart, binary.LittleEndian, routineID)
			// As per the uniquePart buffer, Reset() is faster.
			hash.Reset()
			hash.Write(uniquePart.Bytes())

			// This buffer is passed into the channel, and
			// needs to be allocated freshly (not doing
			// this results in multiple results sharing
			// the same memory allocation (try it, tests
			// will fail, thankfully).
			output = new(bytes.Buffer)
			binary.Write(output, binary.LittleEndian, timestamp)
			binary.Write(output, binary.LittleEndian, hash.Sum64())

			resultChan <- output.Bytes()
		}
	}()
}

// For convenience, generate a batch of IDs in parallel
func GenerateNIDs(mac net.HardwareAddr, n uint64) (ids [][]byte) {
	CPUCount := runtime.GOMAXPROCS(0)
	ids = make([][]byte, n, n)

	// Because we're only using a single routine to drain the
	// channel below, the size of the buffer here has an impact on
	// performance.  This would probably be good to make a
	// tunable value in some applications configuration.
	results := make(chan []byte, 128*CPUCount)

	for routineID := 0; routineID < CPUCount; routineID++ {
		GenerateIDs(mac, routineID, results)
	}

	for i := uint64(0); i < n; i++ {
		ids[i] = <-results
	}
	return
}

// Extract the number nanoseconds since the UNIX epoch at which the ID
// was generated from the ID.  This can be used for sorting, or can be
// converted to a golang Time struct using time.Unix.
func GetUnixNanoFromID(id []byte) (result int64) {
	buf := bytes.NewReader(id)
	binary.Read(buf, binary.LittleEndian, &result)
	return result
}

// Sorting Type for IDs
type ByIDCreationTime [][]byte

func (a ByIDCreationTime) Len() int {
	return len(a)
}
func (a ByIDCreationTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByIDCreationTime) Less(i, j int) bool {

	return GetUnixNanoFromID(a[i]) < GetUnixNanoFromID(a[j])
}
