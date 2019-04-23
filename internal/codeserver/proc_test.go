package codeserver

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parsePort(t *testing.T) {
	t.Parallel()

	var tests = []struct {
		ipPortHex string
		port      string
		err       bool
	}{
		{
			"0100007F:8870",
			"34928",
			false,
		},
		{
			"3A01A8C0:DD14",
			"56596",
			false,
		},
		{
			"3A01A8C0:ACF8",
			"44280",
			false,
		},
		{
			"abc123:456:801294j",
			"0",
			true,
		},
	}

	for _, test := range tests {
		port, err := parsePort(test.ipPortHex)
		if test.err {
			require.Error(t, err)
			return
		}

		require.NoError(t, err)

		require.Equal(t, test.port, port)
	}
}

func Test_parseNetTCPStats(t *testing.T) {
	t.Parallel()

	const procNetTcp = `0100007F:300C 00000000:0000 54838306
0100007F:BEB3 00000000:0000 58828878
00000000:E115 00000000:0000 42213838
017AA8C0:0035 00000000:0000 34316
0101007F:0035 00000000:0000 24568
0100007F:0277 00000000:0000 63674503
0100007F:8C57 00000000:0000 58075873
0100007F:A7F9 00000000:0000 44917881
00000000:227C 00000000:0000 64398395
00000000:AE7D 00000000:0000 64539978
3A01A8C0:8CB4 4AC23AD8:01BB 56951042
3A01A8C0:9F96 35E0BA23:01BB 64464448
3A01A8C0:C72C 4301A8C0:1F49 63436166
3A01A8C0:EA26 A106D9AC:01BB 64534357
3A01A8C0:C9EA 7CFD1EC0:01BB 64363317
3A01A8C0:A878 8E09D9AC:01BB 64511233
3A01A8C0:C21A A97D1A64:01BB 49489905
3A01A8C0:A648 8AC13AD8:01BB 48923906`

	netStats, err := parseNetTCPStats([]byte(procNetTcp))
	require.NoError(t, err)

	require.Len(t, netStats, 18)
	require.Equal(t, "57621", netStats[2].localPort)

	require.Equal(t, "48923906", netStats[17].inode)

	require.Equal(t, "0", netStats[0].remotePort)
}
