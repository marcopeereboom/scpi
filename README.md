scpi
=====

[![Build Status](https://travis-ci.org/marcopeereboom/scpi.png?branch=master)](https://travis-ci.org/marcopeereboom/scpi)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/marcopeereboom/scpi)
[![Go Report Card](https://goreportcard.com/badge/github.com/marcopeereboom/scpi)](https://goreportcard.com/report/github.com/marcopeereboom/scpi)

## scpi Overview

scpi is a "single shot" tool to script scpi enabled instruments. It implements
several simplified commands to create screenshots and csv files of the screen
and additionally it exposes the raw command interface. It is open source and
requires NO drivers and/or manufacturer software and is therefore much easier
to use and maintain. This code is still experimental and patches are welcome.

A description of scpi can be found at
https://en.wikipedia.org/wiki/Standard_Commands_for_Programmable_Instruments

Currently only the LAN interface has been tested but adding a RS232 version
should be trvial.

This tool has been tested with a Rigol DS1054Z and the programmers guide can be
found at:
http://beyondmeasure.rigoltech.com/acton/attachment/1579/f-0386/1/-/-/-/-/DS1000Z_Programming%20Guide_EN.pdf

## Examples

Help:
```
scpi -h {host} -f {out} {csv|screen} [args].

  -f string
        output filename, default is stdout (default "-")
  -h string
        host:port (default "127.0.0.1:5555")

  csv args: none
  raw {command string} (e.g. :WAVeform:SOURce?) Raw commands with a ? suffix will echo the instrument output.
  screen args: {color bool} {inverted bool} {type string {bmp24|bmp8|png|jpeg|tiff}}
```

Take a bmp screenshot of the instrument:
```
scpi -h 10.0.0.196:5555 screen > a.bmp
```

Take a color inverted png of the instrument:
```
scpi -h 10.0.0.196:5555 -f b.png screen true true png
```

Dump csv of current screen on the instrument:
```
scpi -h 10.0.0.196:5555 csv > a.csv
```

Run a raw query command (suffixed with ?):
```
scpi -h 10.0.0.196:5555 raw :WAVeform:SOURce? 
CHAN1
```

Simulate pressing the *auto* button:
```
scpi -h 10.0.0.196:5555 raw :AUToscale
```

Simulate pressing the *stop* button:
```
scpi -h 10.0.0.196:5555 raw :STOP
```

Setting the waveform capture format to ascii. Note the `"` (double quotes) used
for escaping the shell.
```
scpi -h 10.0.0.196:5555 raw ":WAVeform:FORMat ASCii"
scpi -h 10.0.0.196:5555 raw :WAVeform:FORMat?
ASC
```



### Build from source (all platforms)

Building or updating from source requires the following build dependencies:

- **Go**

  Installation instructions can be found here: https://golang.org/doc/install.
  It is recommended to add `$GOPATH/bin` to your `PATH` at this point.

- **Installing scpi**

  `go get github.com/marcopeereboom/scpi`

## License

scpi is licensed under the [copyfree](http://copyfree.org) ISC License.

