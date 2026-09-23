package redisson

import "testing"

func TestReconnectOnHelloCommandErrorEnablesRESP2OnlyOnce(t *testing.T) {
	for _, errString := range []string{
		"unsupported command `hello`",
		"ERR unknown command HELLO",
	} {
		t.Run(errString, func(t *testing.T) {
			c := &client{v: NewConf()}

			if !reconnectErrors[1](c, errString) {
				t.Fatal("HELLO command error should enable RESP2")
			}
			if !c.v.GetAlwaysRESP2() {
				t.Fatal("RESP2 was not enabled")
			}
			if reconnectErrors[1](c, errString) {
				t.Fatal("HELLO command error should not reconnect after RESP2 is enabled")
			}
		})
	}
}
