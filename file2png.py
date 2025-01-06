import argparse
import math
import os
import struct
from PIL import Image
import sys

def encode_file(input_stream, output_stream):
    data = input_stream.read()
    file_size = len(data)
    pixel_count = (file_size + 8) // 3
    dimension = math.ceil(math.sqrt(pixel_count))

    img = Image.new('RGBA', (dimension, dimension), (0, 0, 0, 255))
    pixels = img.load()

    size_buf = struct.pack('<Q', file_size)
    x, y = 0, 0

    for i in range(0, 8, 3):
        r = size_buf[i]
        g = size_buf[i+1] if i+1 < 8 else 0
        b = size_buf[i+2] if i+2 < 8 else 0
        pixels[x, y] = (r, g, b, 255)
        x += 1
        if x >= dimension:
            x = 0
            y += 1

    for i in range(0, len(data), 3):
        r = data[i]
        g = data[i+1] if i+1 < len(data) else 0
        b = data[i+2] if i+2 < len(data) else 0
        pixels[x, y] = (r, g, b, 255)
        x += 1
        if x >= dimension:
            x = 0
            y += 1

    img.save(output_stream, format='PNG')

def decode_file(input_stream, output_stream):
    img = Image.open(input_stream)
    pixels = img.load()
    width, height = img.size

    size_buf = bytearray(8)
    pixel_index = 0
    for i in range(0, 8, 3):
        x = pixel_index % width
        y = pixel_index // width
        r, g, b, _ = pixels[x, y]
        size_buf[i] = r
        if i+1 < 8:
            size_buf[i+1] = g
        if i+2 < 8:
            size_buf[i+2] = b
        pixel_index += 1

    file_size = struct.unpack('<Q', size_buf)[0]
    data = bytearray(file_size)

    data_index = 0
    start_pixel = (8 + 2) // 3
    for y in range(start_pixel // width, height):
        start_x = 0 if y != start_pixel // width else start_pixel % width
        for x in range(start_x, width):
            if data_index >= file_size:
                break
            r, g, b, _ = pixels[x, y]
            data[data_index] = r
            if data_index + 1 < file_size:
                data[data_index+1] = g
            if data_index + 2 < file_size:
                data[data_index+2] = b
            data_index += 3

    output_stream.write(data[:file_size])

def main():
    parser = argparse.ArgumentParser(description="file2png - Convert any file to PNG and back")
    parser.add_argument("-d", action="store_true", help="decode PNG to file")
    parser.add_argument("input", nargs="?", default="-", help="Input file (default: stdin)")
    parser.add_argument("output", nargs="?", default="-", help="Output file (default: stdout)")

    args = parser.parse_args()

    input_stream = sys.stdin.buffer if args.input == "-" else open(args.input, "rb")
    output_stream = sys.stdout.buffer if args.output == "-" else open(args.output, "wb")

    try:
        if args.d:
            decode_file(input_stream, output_stream)
        else:
            encode_file(input_stream, output_stream)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    finally:
        if input_stream is not sys.stdin.buffer:
            input_stream.close()
        if output_stream is not sys.stdout.buffer:
            output_stream.close()

if __name__ == "__main__":
    main()
