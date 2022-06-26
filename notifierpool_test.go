package wsnotifier

import "testing"

func TestNotifierCreate(t *testing.T) {
	p := NewNotifierPool()
	p.AddBroadcast("hi")

	p.RemoveBroadcast("hi")
}
