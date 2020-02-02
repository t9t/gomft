package fragment_test

import (
	"io/ioutil"
	"testing"
	"bytes"
	"math/rand"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t9t/gomft/fragment"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestFragmentReader_Sequential(t *testing.T) {
	
	testData := generateTestData()

	fragments := []fragment.Fragment{
		fragment.Fragment{Offset: 0, Length: 147},
		fragment.Fragment{Offset: 147, Length: 1198},
		fragment.Fragment{Offset: 1345, Length: 1711},
		fragment.Fragment{Offset: 3056, Length: 463},
		fragment.Fragment{Offset: 3519, Length: 1534},
		fragment.Fragment{Offset: 5053, Length: 701},
		fragment.Fragment{Offset: 5754, Length: 1351},
		fragment.Fragment{Offset: 7105, Length: 703},
		fragment.Fragment{Offset: 7808, Length: 1948},
		fragment.Fragment{Offset: 9756, Length: 484},
	}

	r := fragment.NewReader(bytes.NewReader(testData), fragments)

	data, err := ioutil.ReadAll(r)
	require.Nilf(t, err, "unable to read: %v", err)
	
	assert.Equal(t, testData, data)
}

func TestFragmentReader_NonSequential(t *testing.T) {
	testData := generateTestData()

	fragments := []fragment.Fragment{
		fragment.Fragment{Offset: 3756, Length: 1810},
		fragment.Fragment{Offset: 6645, Length: 3423},
		fragment.Fragment{Offset: 803, Length: 6154},
	}

	r := fragment.NewReader(bytes.NewReader(testData), fragments)

	data, err := ioutil.ReadAll(r)
	require.Nilf(t, err, "unable to read: %v", err)

	expected := make([]byte, 0)
	expected = append(expected, testData[3756:3756+1810]...)
	expected = append(expected, testData[6645:6645+3423]...)
	expected = append(expected, testData[803:803+6154]...)
	
	assert.Equal(t, expected, data)
}

func generateTestData() []byte {
	ret := make([]byte, 10240)
	_, _ = rand.Read(ret)
	return ret
}
