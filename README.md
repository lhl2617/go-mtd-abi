# Go MTD ABI

[![Godoc](https://godoc.org/github.com/nathany/looper?status.svg)](https://godoc.org/github.com/lhl2617/go-mtd-abi)
[![License: MIT](https://img.shields.io/badge/License-MIT-informational.svg)](https://opensource.org/licenses/MIT)
[![Linux](https://img.shields.io/static/v1?label=Linux+Kernel+Version&message=v5.12&color=informational)](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tag/?h=v5.12)


Golang implementation of helper functions for the `ioctl` calls in the [Linux MTD ABI](https://git.kernel.org/pub/scm/linux/kernel/git/torvalds/linux.git/tree/include/uapi/mtd/mtd-abi.h?h=v5.12), kernel version `v5.12`.


## Status

EXPERIMENTAL.

No warranties--use at your own risk!

## Usage

Opens a file pointing to an MTD device (`/dev/mtd0`) and obtains MTD info using `MEMGETINFO`.
```golang
package main

import (
	"fmt"
	"os"

	mtdabi "github.com/lhl2617/go-mtd-abi"
	"golang.org/x/sys/unix"
)

const mtdPath = "/dev/mtd0"

func main() {
	mtdFile, err := os.OpenFile(mtdPath, os.O_RDWR, 0644)
	check(err)

	// ▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼▼
	var mtdInfo unix.MtdInfo
	err = mtdabi.MemGetInfo(mtdFile.Fd(), &mtdInfo)
	check(err)
	// ▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲▲

	fmt.Printf("%#v\n", mtdInfo)

	check(mtdFile.Close())
}

// check panics if the error is non-nil.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
```

See more usage examples in the test file ([`mtdabi_test.go`](./mtdabi_test.go)).

## Development Guide

Please run tests in the Vagrant box given. You may spin up the Vagrant box and run tests as such:
```bash
# Set up and boot up the Vagrant box
vagrant up
# SSH into the box
vagrant ssh
# cd into the repo
cd /vagrant
# Run tests
sudo go test ./...
```

## Contributing

* File bugs and/or feature requests in the [GitHub repository](https://github.com/lhl2617/go-mtd-abi)
* Pull requests are welcome in the [GitHub repository](https://github.com/lhl2617/go-mtd-abi)
* Buy me a Coffee ☕️ via [PayPal](https://paypal.me/lhl2617)
