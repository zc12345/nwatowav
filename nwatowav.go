package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/hasenbanck/nwa"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
)

var inputfile = flag.String("inputfile", "", "path to the input file.")
var outputfile = flag.String("outputfile", "", "path to the output file.")

type fileType int

const (
	NONE fileType = iota
	NWA
	NWK
	OVK
)

func main() {
	flag.Parse()

	if *inputfile == "" {
		log.Fatal("You need to define an input file!")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	file, err := os.Open(*inputfile)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	var outfilename, outext, outpath string
	var filetype fileType
	var headblksz int64
	
	if *outputfile == "" {
		outfilename = strings.Split(*inputfile, ".")[0]
	}else{
		outfilename = *outputfile
	}

	switch {
	case strings.Contains(*inputfile, ".nwa"):
		{
			filetype = NWA
			outext = "wav"
		}
	case strings.Contains(*inputfile, ".nwk"):
		{
			filetype = NWK
			headblksz = 12
			outext = "wav"
		}
	case strings.Contains(*inputfile, ".ovk"):
		{
			filetype = OVK
			headblksz = 16
			outext = "ogg"
		}
	}
	if filetype == NONE {
		log.Fatal("This program can only handle .nwa/.nwk/.ovk files right now.")
	}

	if filetype == NWA {
		var data io.Reader
		if data, err = nwa.NewNwaFile(file); err != nil {
			log.Fatal(err)
		}

		outpath = fmt.Sprintf("%s.%s", outfilename, outext)

		var out *os.File
		out, err = os.Create(outpath)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close()

		if _, err = io.Copy(out, data); err != nil {
			log.Fatal(err)
		}
	} else { // NWK or OVK files
		var indexcount int32
		binary.Read(file, binary.LittleEndian, &indexcount)
		if indexcount <= 0 {
			if filetype == OVK {
				log.Fatalf("Invalid Ogg-ovk file: %s: index = %d\n", inputfile, indexcount)
			} else {
				log.Fatalf("Invalid Koe-nkw file: %s: index = %d\n", inputfile, indexcount)
			}
		}

		tblsiz := make([]int32, indexcount)
		tbloff := make([]int32, indexcount)
		tblcnt := make([]int32, indexcount)
		tblorigsiz := make([]int32, indexcount)

		var i int32
		for i = 0; i < indexcount; i++ {
			buffer := new(bytes.Buffer)
			if count, err := io.CopyN(buffer, file, headblksz); count != headblksz || err != nil {
				log.Fatal("Couldn't read the index entries!")
			}
			binary.Read(buffer, binary.LittleEndian, &tblsiz[i])
			binary.Read(buffer, binary.LittleEndian, &tbloff[i])
			binary.Read(buffer, binary.LittleEndian, &tblcnt[i])
			binary.Read(buffer, binary.LittleEndian, &tblorigsiz[i])
		}

		c := make(chan int, indexcount)
		for i = 0; i < indexcount; i++ {
			if tbloff[i] <= 0 || tblsiz[i] <= 0 {
				log.Fatalf("Invalid table[%d]: cnt %d, off %d, size %d\n", i, tblcnt[i], tbloff[i], tblsiz[i])
				continue
			}
			outpath = fmt.Sprintf("%s-%d.%s", outfilename, tblcnt[i], outext)
			go doDecode(filetype, outpath, *inputfile, tbloff[i], tblsiz[i], c)
		}
		for i = 0; i < indexcount; i++ {
			<-c
		}
	}
}

func doDecode(filetype fileType, filename string, datafile string, offset int32, size int32, c chan int) {
	var count int64
	var data io.Reader

	file, err := os.Open(datafile)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	buffer := new(bytes.Buffer)
	file.Seek(int64(offset), 0)
	if count, err = io.CopyN(buffer, file, int64(size)); count != int64(size) || err != nil {
		log.Fatalf("Couldn't read the data for filename %s: off %d, size %d. Error: %s\n", filename, offset, size, err)
	}

	if filetype == NWK {
		if data, err = nwa.NewNwaFile(buffer); err != nil {
			log.Fatal(err)
		}
	} else {
		data = buffer
	}

	out, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, data); err != nil {
		log.Fatal(err)
	}
	c <- 1
}
