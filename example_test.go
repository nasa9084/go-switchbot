package switchbot_test

import (
	"context"
	"fmt"

	"github.com/nasa9084/go-switchbot"
)

func ExamplePrintPhysicalDevices() {
	const openToken = "blahblahblah"

	c := switchbot.New(openToken)

	// get physical devices and show
	pdev, _, _ := c.Device().List(context.Background())

	for _, d := range pdev {
		fmt.Printf("%s\t%s\n", d.Type, d.Name)
	}
}
