package gadioRss

import (
	"github.com/smartystreets/assertions"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestGetGadioList(t *testing.T) {
	radiosNum := 10
	if radios, err := GetGRadios(radiosNum); err != nil {
		log.Println(err)
		assert.Fail(t, err.Error())
	} else {
		log.Println(len(*radios.Data), *radios.Data)
		assertions.ShouldEqual(len(*radios.Data), radiosNum)
	}
}
