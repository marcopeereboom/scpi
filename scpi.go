package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func readN(c net.Conn, n int) ([]byte, error) {
	d := make([]byte, n)
	at := 0
	for at < n {
		x, err := c.Read(d[at:])
		if err != nil {
			return nil, err
		}
		at += x
	}
	return d, nil
}

func readOne(c net.Conn) ([]byte, error) {
	return readN(c, 1)
}

func readTMCBlockHeaderIntegerSize(c net.Conn) (int, error) {
	// find size of integer
	d, err := readOne(c)
	if err != nil {
		return -1, err
	}
	size, err := strconv.ParseInt(string(d), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid file size read")
	}

	// read file size
	fileSizeString, err := readN(c, int(size))
	if err != nil {
		return -1, err
	}

	fileSize, err := strconv.ParseInt(string(fileSizeString), 10, 64)
	if err != nil {
		return -1, err
	}
	//fmt.Fprintf(os.Stderr, "fileSize: %v\n", fileSize)

	return int(fileSize), nil
}

func readTMCBlockHeader(c net.Conn) (int, error) {
	// find #
	d, err := readOne(c)
	if err != nil {
		return -1, err
	}
	if d[0] != '#' {
		return -1, fmt.Errorf("invalid header")
	}

	return readTMCBlockHeaderIntegerSize(c)
}

func readTMCBlock(c net.Conn) ([]byte, error) {
	size, err := readTMCBlockHeader(c)
	if err != nil {
		return nil, err
	}

	return readN(c, size)
}

func command(c net.Conn, cmd, params string) error {
	x := cmd + " " + params + "\n"
	//fmt.Fprintf(os.Stderr, "command: %v", x)
	_, err := c.Write([]byte(x))
	if err != nil {
		return err
	}
	return nil
}

func boolText(x bool) string {
	if x {
		return "ON"
	}
	return "OFF"
}

func screenShot(c net.Conn, f *os.File, args []string) error {
	var params string
	switch len(args) {
	case 0:
		params = "ON,OFF,BMP24"
	case 3:
		switch strings.ToUpper(args[2]) {
		case "BMP24":
		case "BMP8":
		case "PNG":
		case "JPEG":
		case "TIFF":
		default:
			return fmt.Errorf("invalid format, valid options are:" +
				" {BMP24|BMP8|PNG|JPEG|TIFF}")
		}
		color, err := strconv.ParseBool(args[0])
		if err != nil {
			return err
		}
		inverted, err := strconv.ParseBool(args[1])
		if err != nil {
			return err
		}
		params = fmt.Sprintf("%v,%v,%v", boolText(color),
			boolText(inverted), strings.ToUpper(args[2]))
	default:
		return fmt.Errorf("expected {color bool,inverted bool,type " +
			"{{BMP24|BMP8|PNG|JPEG|TIFF}}}")
	}

	err := command(c, ":DISP:DATA?", params)
	if err != nil {
		return err
	}

	data, err := readTMCBlock(c)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func csv(c net.Conn, f *os.File) error {
	// assume ascii
	err := command(c, ":WAVeform:FORMat", "ASCii")
	if err != nil {
		return err
	}
	err = command(c, ":WAVeform:DATA?", "")
	if err != nil {
		return err
	}

	data, err := readTMCBlock(c)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}

func raw(c net.Conn, f *os.File, args []string) error {
	var cmd string
	for _, v := range args {
		cmd += v + " "
	}
	cmd += "\n"
	err := command(c, cmd, "")
	if err != nil {
		return err
	}

	// if there is a ? read the output
	if !strings.Contains(cmd, "?") {
		return nil
	}

	// if the first byte is # assume it is a TBC block header
	d, err := readOne(c)
	if err != nil {
		return err
	}
	if d[0] == '#' {
		// assume TBC block
		size, err := readTMCBlockHeaderIntegerSize(c)
		if err != nil {
			return err
		}
		data, err := readN(c, size)
		if err != nil {
			return err
		}
		_, err = f.Write(data)
		if err != nil {
			return err
		}
		return nil
	}

	// echo output one byte at a time
	fmt.Fprintf(f, "%c", d[0])
	for {
		d, err := readOne(c)
		if err != nil {
			return err
		}
		fmt.Fprintf(f, "%c", d[0])
		if d[0] == '\n' {
			break
		}
	}

	return nil
}

func _main() error {
	host := flag.String("h", "127.0.0.1:5555", "host:port")
	filename := flag.String("f", "-", "output filename, default is stdout")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "scpi -h {host} -f {out} {csv|screen}"+
			" [args].\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "  csv args: none\n")
		fmt.Fprintf(os.Stderr, "  raw {command string} (e.g. "+
			":WAVeform:SOURce?)")
		fmt.Fprintf(os.Stderr, " Raw commands with a ? suffix will "+
			"echo the instrument output.\n")
		fmt.Fprintf(os.Stderr, "  screen args: {color bool} "+
			"{inverted bool} "+
			"{type string {bmp24|bmp8|png|jpeg|tiff}}\n")
	}

	flag.Parse()

	var (
		f   *os.File
		err error
	)
	if *filename == "-" {
		f = os.Stdout
	} else {
		f, err = os.Create(*filename)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	if len(flag.Args()) == 0 {
		flag.Usage()
		return fmt.Errorf("invalid action")
	}

	c, err := net.Dial("tcp", *host)
	if err != nil {
		return err
	}

	switch flag.Args()[0] {
	case "csv":
		return csv(c, f)
	case "raw":
		return raw(c, f, flag.Args()[1:])
	case "screen":
		return screenShot(c, f, flag.Args()[1:])
	}

	flag.Usage()
	return fmt.Errorf("invalid action")
}

func main() {
	err := _main()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
