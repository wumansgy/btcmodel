package main

import (
	"bufio"
	"os"
	"strings"
	"fmt"
)

func main() {
	cli := CLI{}

	input := bufio.NewScanner(os.Stdin)
	Welcome()
	cli.Help()

	for {
		fmt.Printf("\ncmd>")

		//send 1111 222    10
		if input.Scan() {
			s := input.Text()
			fmt.Printf("%s\n", s)
			cmds := strings.Fields(s)
			//array := strings.Split(s, " ")

			if len(cmds) != 0 {
				cli.Run(cmds)
				fmt.Printf("+++++++++++++++++++++++++++++++\n")
			}
		}
	}
}
