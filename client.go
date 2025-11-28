package main

import (
	"fmt"
	"image/color"
	"net/rpc"
	"os"
	"strconv"

	"github.com/ahmedsat/alat/alat"
)

type Client struct {
}

func (c *Client) Run(args ...string) error {

	client, err := rpc.Dial("tcp", "localhost:8080") // Connect to server at localhost:8080
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to server:", err)
		return err
	}
	defer client.Close()

	command := args[0]
	var serviceMethod string
	var result string

	switch command {
	case "Close":
		usage := "Usage: Close <code>"
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, usage)
			return nil
		}

		serviceMethod = "Server.Close"
		code, err := strconv.ParseInt(args[1], 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, usage)
			fmt.Fprintln(os.Stderr, "Error parsing code")
			return err
		}
		err = client.Call(serviceMethod, int(code), &result)
	case "CloseWindow":
		usage := "Usage: CloseWindow <id>"
		if len(args) != 2 {
			fmt.Fprintln(os.Stderr, usage)
			return nil
		}

		serviceMethod = "WindowCreator.Close"
		id := args[1]
		err = client.Call(serviceMethod, id, &result)
	case "SolidColor":

		usage := "Usage: SolidColor <id> <width> <height> <title> <r> <g> <b>"
		if len(args) != 8 {
			fmt.Fprintln(os.Stderr, usage)
			return nil
		}

		serviceMethod = "WindowCreator.Solid"
		id := args[1]
		width, err := strconv.ParseInt(args[2], 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, usage)
			fmt.Fprintln(os.Stderr, "Error parsing width")
			return err
		}
		height, err := strconv.ParseInt(args[3], 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, usage)
			fmt.Fprintln(os.Stderr, "Error parsing height")
			return err
		}
		title := args[4]
		r, err := strconv.ParseInt(args[5], 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, usage)
			fmt.Fprintln(os.Stderr, "Error parsing r")
			return err
		}
		g, err := strconv.ParseInt(args[6], 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, usage)
			fmt.Fprintln(os.Stderr, "Error parsing g")
			return err
		}
		b, err := strconv.ParseInt(args[7], 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, usage)
			fmt.Fprintln(os.Stderr, "Error parsing b")
			return err
		}
		err = client.Call(serviceMethod, alat.SolidColorArgs{
			Id:     id,
			Width:  int(width),
			Height: int(height),
			Title:  title,
			Color: color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: 0xff,
			},
		}, &result)

	case "QrWindow":
		usage := "Usage: QrWindow <id> <text> -s [size] -t [title] -l [recoveryLevel]"
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, usage)
			return nil
		}

		serviceMethod = "WindowCreator.Qr"
		id := args[1]
		text := args[2]
		size := 512
		title := "QR"
		recoveryLevel := 0
		for i := 3; i < len(args); i += 2 {
			switch args[i] {
			case "-s":
				s, err := strconv.ParseInt(args[i+1], 0, 64)
				if err != nil {
					fmt.Fprintln(os.Stderr, usage)
					fmt.Fprintln(os.Stderr, "Error parsing size")
					return err
				}
				size = int(s)
			case "-t":
				title = args[i+1]
			case "-l":
				l, err := strconv.ParseInt(args[i+1], 0, 64)
				if err != nil {
					fmt.Fprintln(os.Stderr, usage)
					fmt.Fprintln(os.Stderr, "Error parsing recoveryLevel")
					return err
				}
				recoveryLevel = int(l)
			}
		}
		err = client.Call(serviceMethod, alat.QrArgs{
			Id:            id,
			Text:          text,
			Size:          size,
			Title:         title,
			RecoveryLevel: recoveryLevel,
		}, &result)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error calling %s: %v", serviceMethod, err)
		return err
	}
	fmt.Printf("Result of %s:%v", serviceMethod, result)

	return nil
}
