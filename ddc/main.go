// Copyright 2021 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// ddc communicates to an monitor over DDC.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
)

func calc(d []byte) []byte {
	d = append([]byte{0x50, 0x80 | byte(len(d))}, d...)
	c := byte(0x50)
	for _, b := range d {
		c = c ^ b
	}
	return append(d, c)
}

func mainImpl() error {
	busName := flag.String("b", "", "I²C bus to use")
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	bus, err := i2creg.Open(*busName)
	if err != nil {
		return err
	}
	defer bus.Close()

	// DDC-CI command address.
	d := i2c.Dev{Bus: bus, Addr: 0x37}

	// https://glenwing.github.io/docs/VESA-DDCCI-1.1.pdf
	w := calc([]byte{0xF3, 0, 0})
	if err = d.Tx(w[:], nil); err != nil {
		return err
	}
	time.Sleep(40 * time.Millisecond)

	r := [256]byte{}
	if err = d.Tx(nil, r[:2]); err != nil {
		return err
	}
	l := r[1] &^ 0x80
	for i := byte(0); i < l; i += 32 {
		v := l - i
		if v > 32 {
			v = 32
		}
		if err = d.Tx(nil, r[i:v]); err != nil {
			return err
		}
	}
	fmt.Printf("%#x\n", r[:l])
	/*
		for i, b := range buf {
			if i != 0 {
				if _, err = fmt.Print(", "); err != nil {
					break
				}
			}
			if _, err = fmt.Printf("0x%02X", b); err != nil {
				break
			}
		}
		_, err = fmt.Print("\n")
		return err
	*/
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ddc: %s.\n", err)
		os.Exit(1)
	}
}
