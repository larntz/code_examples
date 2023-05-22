package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ew := flag.NewFlagSet("nw", flag.ExitOnError)
	ewInterstate := ew.Int("interstate", 0, "Which interstate we should use.")

	ns := flag.NewFlagSet("ns", flag.ExitOnError)
	nsInterstate := ns.Int("interstate", 0, "Which interstate we should use.")

	if len(os.Args) == 1 {
		fmt.Println("You must enter a subcommand")
		fmt.Println(" ns for north/south")
		fmt.Println(" ew for east/west")
		return
	}

	switch os.Args[1] {
	case "ew":
		ew.Parse(os.Args[2:])
		fmt.Println(*ewInterstate)
	case "ns":
		ns.Parse(os.Args[2:])
		fmt.Println(*nsInterstate)
	default:
		fmt.Printf("%q is not a valid subcommand\n", os.Args[1])
		os.Exit(1)
	}
}
