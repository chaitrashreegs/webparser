package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chaitrashreegs/webparser/pkg"
	"github.com/chaitrashreegs/webparser/util"
)

func main() {
	//read the inputs from cmd-line
	input := util.UserInput{}
	flag.StringVar(&input.Address, "address", "0.0.0.0", "--address on which web server runs.\n Format:(x:x:x:x)\n")
	flag.StringVar(&input.Port, "port", "8090", "--Port on which web server runs.\n")
	flag.StringVar(&input.OutputFilePath, "file-path", "./counter", "--file-path to specify the path to store the counts")
	flag.DurationVar(&input.WindowSize, "window-size", 60*time.Second, "--window-size for which counter to be displayed.\n")
	flag.DurationVar(&input.Precison, "precision", 1*time.Second, "--precision for which counter to be displayed.\n --Permissable values {second,milli,micro,nano}\n")
	flag.Parse()
	log.Println("Parsed input values:", input)

	//validate IP address from the user brfore starting server
	if (net.ParseIP(input.Address)) == nil && input.Address != "localhost" {
		log.Fatal("Failed to start server ,invalid address format")
	}

	if input.WindowSize < input.Precison {
		log.Fatal("Invalid window size and precision , always windowsize >precision")
	}
	counter := pkg.NewRequestCounter(input.Precison, input.WindowSize, input.OutputFilePath)
	pkg.Initializaion(counter)

	mux := http.NewServeMux()
	mux.Handle("/counter", pkg.GetCounter(counter))

	go func() {
		err := pkg.NewServer(input.Address+":"+input.Port, mux)
		if err != nil {
			log.Fatal("Failed to start server ,err :", err)
		}
	}()

	handleSignal(counter)
}

func handleSignal(c pkg.Parser) {
	// Create a channel to receive signals
	sigCh := make(chan os.Signal, 1)

	// Register the channel to receive specified signals
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	// Block until a signal is received
	<-sigCh
	fmt.Println("recived signal to store file")
	c.HandleShutdown()
}
