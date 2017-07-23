package idgen

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IdGenTestSuite struct {
	suite.Suite
}

// GenerateIDs returns a 16 byte (128bit)
func (suite *IdGenTestSuite) TestGenerateIDsSingleID() {
	mac, _ := GetMACAddress()
	results := make(chan []byte, 1)
	GenerateIDs(mac, 0, results)
	result1 := <-results
	assert.Equal(suite.T(), 16, len(result1), "ID should be a 16 byte (128bit) array")
}

// GenerateIDs returns a different value each time (note, this test is
// not definitive, but it would catch dumb changes).
func (suite *IdGenTestSuite) TestGenerateIDsDoesNotRepeatItself() {
	mac, _ := GetMACAddress()
	results := make(chan []byte, 1)
	GenerateIDs(mac, 0, results)
	result1 := <-results
	result2 := <-results
	assert.NotEqual(suite.T(), result1, result2, fmt.Sprintf("%v should not equal %v\n",
		result1, result2))
}

// Benchmark plain GenerateIDs performance when generating 1 ID.
func BenchmarkGenerateIDs(b *testing.B) {
	mac, _ := GetMACAddress()
	results := make(chan []byte, 1)
	GenerateIDs(mac, 0, results)

	// Repeat the GenerateIDs call
	for n := 0; n < b.N; n++ {
		<-results
	}
}

// GenerateNIDs creates a given number of IDs
func (suite *IdGenTestSuite) TestGenerateNIDsCreatesNIDs() {
	mac, _ := GetMACAddress()
	result := GenerateNIDs(mac, 10)
	assert.Equal(suite.T(), 10, len(result))
}

// GenerateNIDs doesn't repeat itself
func (suite *IdGenTestSuite) TestGenerateNIDsDoesNotRepeatItself() {
	mac, _ := GetMACAddress()
	result := GenerateNIDs(mac, 10)
	for a := 0; a < 9; a++ {
		for b := a + 1; b < 10; b++ {
			// We iterate across all the combinations and
			// check that no two IDS match.
			assert.NotEqual(suite.T(), result[a], result[b])
		}
	}
}

// Benchmark GenerateNIDs performance when generating 1 ID.
func BenchmarkGenerateNIDs1(b *testing.B) {
	mac, _ := GetMACAddress()

	// Repeat the GenerateIDs call
	for n := 0; n < b.N; n++ {
		GenerateNIDs(mac, 1)
	}
}

// Benchmark GenerateNIDs performance when generating 10000 ids.
func BenchmarkGenerateNIDs10000(b *testing.B) {
	mac, _ := GetMACAddress()

	// Repeat the GenerateIDs call
	for n := 0; n < b.N; n++ {
		GenerateNIDs(mac, 10000)
	}
}

// Benchmark GenerateNIDs performance when generating 100000 ids.
func BenchmarkGenerateNIDs100000(b *testing.B) {
	mac, _ := GetMACAddress()

	// Repeat the GenerateNIDs call
	for n := 0; n < b.N; n++ {
		GenerateNIDs(mac, 100000)
	}
}

// Benchmark GenerateNIDs performance when generating 100000 ids.
func BenchmarkGenerateNIDs1000000(b *testing.B) {
	mac, _ := GetMACAddress()

	// Repeat the GenerateNIDs call
	for n := 0; n < b.N; n++ {
		GenerateNIDs(mac, 1000000)
	}
}

func (suite *IdGenTestSuite) TestGetUnixNanoFromID() {
	id := []byte("\x7e\x58\xdc\xce\x28\xda\xbd\x14\x9b\x89\xd1\x2b\xc5\x8c\xa5\xe0")
	result := GetUnixNanoFromID(id)
	assert.Equal(suite.T(), int64(1494590520160966782), result)
}

// ByIdCreationTime sorts IDs in ascending order of creation time
func (suite *IdGenTestSuite) TestByIDCreationTime() {
	first := []byte("\x7e\x58\xdc\xce\x28\xda\xbd\x14\x9b\x89\xd1\x2b\xc5\x8c\xa5\xe0")
	second := []byte("\xef\x61\xdc\xce\x28\xda\xbd\x14\x8e\xe1\xf2\x81\x57\xe9\x9a\x42")
	third := []byte("\xa3\x7a\xdc\xce\x28\xda\xbd\x14\xe1\x2b\xa2\x8a\xb7\x6f\x3b\x11")

	first_time := GetUnixNanoFromID(first)
	second_time := GetUnixNanoFromID(second)
	third_time := GetUnixNanoFromID(third)

	// Start with an unordered slice
	input := make([][]byte, 3)
	input[0] = second
	input[1] = third
	input[2] = first
	sort.Sort(ByIDCreationTime(input))
	assert.Equal(suite.T(), first_time, GetUnixNanoFromID(input[0]))
	assert.Equal(suite.T(), second_time, GetUnixNanoFromID(input[1]))
	assert.Equal(suite.T(), third_time, GetUnixNanoFromID(input[2]))
}

func TestIdGenTestSuite(t *testing.T) {
	suite.Run(t, new(IdGenTestSuite))
}
