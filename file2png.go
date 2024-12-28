package main

import (
    "encoding/binary"
    "flag"
    "fmt"
    "image"
    "image/color"
    "image/png"
    "io"
    "math"
    "os"
)

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "file2png - Convert any file to PNG and back\n\n")
        fmt.Fprintf(os.Stderr, "Usage:\n")
        fmt.Fprintf(os.Stderr, "  Encode: file2png input_file output.png\n")
        fmt.Fprintf(os.Stderr, "  Decode: file2png -d input.png output_file\n")
        fmt.Fprintf(os.Stderr, "  Pipe:   command1 | file2png > output.png\n")
        fmt.Fprintf(os.Stderr, "          file2png -d < input.png > output_file\n\n")
        fmt.Fprintf(os.Stderr, "Options:\n")
        fmt.Fprintf(os.Stderr, "  -d    decode PNG to file\n")
        fmt.Fprintf(os.Stderr, "  -h    show this help message\n\n")
        fmt.Fprintf(os.Stderr, "Examples:\n")
        fmt.Fprintf(os.Stderr, "  file2png document.pdf document.png\n")
        fmt.Fprintf(os.Stderr, "  file2png -d document.png document.pdf\n")
        fmt.Fprintf(os.Stderr, "  cat file | file2png > encoded.png\n")
    }

    var decode bool
    flag.BoolVar(&decode, "d", false, "decode PNG to file")
    flag.Parse()

    args := flag.Args()

    var input io.Reader = os.Stdin
    var output io.Writer = os.Stdout

    // Show usage only if arguments are provided but incorrect
    if len(args) != 0 && len(args) != 2 {
        flag.Usage()
        os.Exit(1)
    }

    if len(args) == 2 {
        inFile, err := os.Open(args[0])
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error opening input file: %v\n", err)
            os.Exit(1)
        }
        defer inFile.Close()
        input = inFile

        outFile, err := os.Create(args[1])
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error creating output file: %v\n", err)
            os.Exit(1)
        }
        defer outFile.Close()
        output = outFile
    }

    if decode {
        if err := decodeFile(input, output); err != nil {
            fmt.Fprintf(os.Stderr, "Error decoding: %v\n", err)
            os.Exit(1)
        }
        return
    }

    if err := encodeFile(input, output); err != nil {
        fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
        os.Exit(1)
    }
}

func encodeFile(r io.Reader, w io.Writer) error {
    data, err := io.ReadAll(r)
    if err != nil {
        return err
    }

    fileSize := len(data)
    pixelCount := (fileSize + 8) / 3
    dimension := int(math.Ceil(math.Sqrt(float64(pixelCount))))
    
    img := image.NewRGBA(image.Rect(0, 0, dimension, dimension))

    sizeBuf := make([]byte, 8)
    binary.LittleEndian.PutUint64(sizeBuf, uint64(fileSize))

    x, y := 0, 0

    for i := 0; i < 8; i += 3 {
        r := sizeBuf[i]
        g := byte(0)
        b := byte(0)
        if i+1 < 8 {
            g = sizeBuf[i+1]
        }
        if i+2 < 8 {
            b = sizeBuf[i+2]
        }
        img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
        x++
        if x >= dimension {
            x = 0
            y++
        }
    }

    for i := 0; i < len(data); i += 3 {
        r := data[i]
        g := byte(0)
        b := byte(0)
        if i+1 < len(data) {
            g = data[i+1]
        }
        if i+2 < len(data) {
            b = data[i+2]
        }
        img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
        x++
        if x >= dimension {
            x = 0
            y++
        }
    }

    return png.Encode(w, img)
}

func decodeFile(r io.Reader, w io.Writer) error {
    img, err := png.Decode(r)
    if err != nil {
        return err
    }

    bounds := img.Bounds()
    width, height := bounds.Max.X, bounds.Max.Y

    sizeBuf := make([]byte, 8)
    pixelIndex := 0
    for i := 0; i < 8; i += 3 {
        x := pixelIndex % width
        y := pixelIndex / width
        c := img.At(x, y)
        r, g, b, _ := c.RGBA()
        sizeBuf[i] = byte(r)
        if i+1 < 8 {
            sizeBuf[i+1] = byte(g)
        }
        if i+2 < 8 {
            sizeBuf[i+2] = byte(b)
        }
        pixelIndex++
    }

    fileSize := binary.LittleEndian.Uint64(sizeBuf)
    data := make([]byte, fileSize)
    dataIndex := 0

    startPixel := (8 + 2) / 3

    for y := startPixel / width; y < height; y++ {
        startX := 0
        if y == startPixel/width {
            startX = startPixel % width
        }
        for x := startX; x < width; x++ {
            if dataIndex >= int(fileSize) {
                break
            }
            c := img.At(x, y)
            r, g, b, _ := c.RGBA()
            data[dataIndex] = byte(r)
            if dataIndex+1 < int(fileSize) {
                data[dataIndex+1] = byte(g)
            }
            if dataIndex+2 < int(fileSize) {
                data[dataIndex+2] = byte(b)
            }
            dataIndex += 3
        }
    }

    _, err = w.Write(data[:fileSize])
    return err
}
