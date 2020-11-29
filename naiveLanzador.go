package main

import (
	"bufio"
	"fmt"
	"os"
	"net"
)

func main(){
	conex, _ := net.Dial("tcp","localhost:8000")
	defer conex.Close()

	rIn := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese mensaje: ")
	msg, _ := rIn.ReadString('\n')

	
	fmt.Fprintln(conex, msg)

}


