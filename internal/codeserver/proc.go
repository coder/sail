package codeserver

import (
	"bytes"
	"strconv"
	"strings"

	"go.coder.com/sail/internal/dockutil"
	"golang.org/x/xerrors"
)

var (
	// PortNotFoundError is returned whenever the port isn't found.
	// This can happen if code-server hasn't started it's listener yet
	// or if the code-server process failed for any reason.
	PortNotFoundError = xerrors.New("failed to find port")
)

// PID returns the pid of code-server running inside of the container.
func PID(containerName string) (int, error) {
	out, err := dockutil.FmtExec(containerName, "pgrep -P 1 code-server").CombinedOutput()
	if err != nil {
		return 0, xerrors.Errorf("%s: %w", out, err)
	}

	return strconv.Atoi(strings.TrimSpace(string(out)))
}

// Port returns the port that code-server is listening on.
// To get the port value, it first finds all of the socket
// inodes the process is using, then reads the entries in
// `/proc/net/tcp`. It maps the socket inodes to the inodes
// listed in `/proc/net/tcp` and returns the port that has a
// remote address of `0` since this is the listener.
func Port(containerName string) (string, error) {
	inodes, err := codeServerSocketInodes(containerName)
	if err != nil {
		return "", err
	}

	stats, err := netTCPStats(containerName)
	if err != nil {
		return "", err
	}

	m := make(map[string]*netStat, len(stats))
	for _, stat := range stats {
		m[stat.inode] = stat
	}

	for _, inode := range inodes {
		stat, ok := m[inode]
		if !ok {
			continue
		}

		if stat.remotePort == "0" {
			return stat.localPort, nil
		}
	}

	return "", PortNotFoundError
}

// codeServerSocketInodes returns all of the socket inodes in use by the code-server process.
//
// This function reads the code-server processes' `/proc/<pid>/fd` directory to see all of
// the open file descriptors the process has open. We grep for any file descriptors that
// are links to sockets and we awk to just parse out the inode field.
//
// See: http://man7.org/linux/man-pages/man5/proc.5.html for more information about `/proc/<pid>/fd`.
func codeServerSocketInodes(containerName string) ([]string, error) {
	pid, err := PID(containerName)
	if err != nil {
		return nil, xerrors.Errorf("failed to find code-server pid: %w", err)
	}

	// This command parses the output of `find` to access the inode field.
	// For example, this line from `find`:
	// `65595472      0 lrwx------   1 root     root           64 Apr 23 11:30 /proc/1/fd/8 -> socket:[50784188]`
	//
	// would turn into:
	// `50784188`
	out, err := dockutil.FmtExec(
		containerName,
		// find all fd that are links | grep for sockets | get the socket:[inode] | split on `:`, remove the `[]` brackets, and output the inode.
		`find /proc/%d/fd -type l -ls | grep socket | awk '{ print $13 }' | awk -F ":" '{ gsub("\\[|\\]", "", $2); print $2 }'`,
		pid,
	).CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("%s: %w", out, err)
	}

	return strings.Split(string(bytes.TrimSpace(out)), "\n"), nil
}

type netStat struct {
	remotePort string
	localPort  string
	inode      string
}

// netTCPStats returns the entries in /proc/net/tcp inside of the container.
func netTCPStats(containerName string) ([]*netStat, error) {
	// This command reads the entries in `/proc/net/tcp`, removes the header line with `NR > 1`, and gets
	// the local_address, rem_address, and inode fields. See: https://www.kernel.org/doc/Documentation/networking/proc_net_tcp.txt
	//
	// An example of the first two lines in `/proc/net/tcp` before doing any awk transformations:
	// `sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode`
	// `0: 0100007F:BEB3 00000000:0000 0A 00000000:00000000 00:00000000 00000000  1000        0 58828878 1 0000000000000000 100 0 0 10 0`
	//
	// After the awk transformation, it would turn into:
	// `0100007F:BEB3 00000000:0000 58828878`
	out, err := dockutil.FmtExec(containerName, `cat /proc/net/tcp | awk 'NR > 1 {print $2, $3, $10 }'`).CombinedOutput()
	if err != nil {
		return nil, xerrors.Errorf("%s: %w", out, err)
	}

	return parseNetTCPStats(out)
}

// parseNetTCPStats parses the fields from the netTCPStats output.
func parseNetTCPStats(out []byte) ([]*netStat, error) {
	lines := bytes.Split(bytes.TrimSpace(out), []byte("\n"))

	var (
		err   error
		stats = make([]*netStat, 0, len(lines))
	)
	for _, line := range lines {
		fields := strings.Fields(string(bytes.TrimSpace(line)))
		if len(fields) != 3 {
			return nil, xerrors.Errorf("line formatted incorrectly: %s", line)
		}

		var stat netStat
		stat.localPort, err = parseHexPort(fields[0])
		if err != nil {
			return nil, err
		}

		stat.remotePort, err = parseHexPort(fields[1])
		if err != nil {
			return nil, err
		}

		stat.inode = fields[2]

		stats = append(stats, &stat)
	}

	return stats, nil
}

// parseHexPort parses the port field from the ip:port hex combination.
// This takes in a local_address or rem_address field from `/proc/net/tcp`
// and parses the hex port into a base 10 port string.
func parseHexPort(ipPortHex string) (string, error) {
	fields := strings.Split(ipPortHex, ":")
	if len(fields) != 2 {
		return "", xerrors.Errorf("failed to parse port: %s", ipPortHex)
	}

	portHex := fields[1]

	port, err := strconv.ParseUint(portHex, 16, 16)
	if err != nil {
		return "", err
	}

	return strconv.FormatUint(port, 10), nil
}
