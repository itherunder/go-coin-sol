package discriminator

import (
	"fmt"
	"testing"

	go_test_ "github.com/pefish/go-test"
)

func TestGetDiscriminator(t *testing.T) {
	r := GetDiscriminator("global", "swap")
	fmt.Println(r)
	go_test_.Equal(t, "f8c69e91e17587c8", r)

	r = GetDiscriminator("global", "swap_v2")
	fmt.Println(r)
	go_test_.Equal(t, "2b04ed0b1ac91e62", r)
}
