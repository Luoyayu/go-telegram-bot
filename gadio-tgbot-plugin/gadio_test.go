package gadioRss

import (
	"bytes"
	"encoding/json"
	"github.com/smartystreets/assertions"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"os"
	"testing"
)

func TestGetGadioList(t *testing.T) {
	radiosNum := 10
	if radios, err := GetGRadios(radiosNum); err != nil {
		log.Println(err)
		assert.Fail(t, err.Error())
	} else {
		log.Println(len(*radios.Data), *radios.Data)
		dataFile, _ := os.Create("tmp_data.json")
		includedFile, _ := os.Create("tmp_included.json")

		gadioBytes, _ := json.Marshal(*radios.Data)
		includedByty, _ := json.Marshal(*radios.Included)

		_, _ = io.Copy(dataFile, bytes.NewReader(gadioBytes))
		_, _ = io.Copy(includedFile, bytes.NewReader(includedByty))
		defer dataFile.Close()
		defer includedFile.Close()

		assertions.ShouldEqual(len(*radios.Data), radiosNum)
	}
}
