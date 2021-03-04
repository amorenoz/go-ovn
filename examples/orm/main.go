package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	goovn "github.com/ebay/go-ovn"
)

const (
	ovnnbSocket = "ovnnb_db.sock"
)

var (
	orm             goovn.ORMClient
	exampleModel, _ = goovn.NewDBModel([]goovn.Model{
		&LogicalRouter{},
	})
)

type ormSignal struct{}

func (s ormSignal) OnCreated(m goovn.Model) {
	switch m.Table() {
	case "Logical_Router":
		lr := m.(*LogicalRouter)
		fmt.Printf("Hey! I got a new Logical Router! Check it out:\n")
		fmt.Printf("%+v\n", *lr)
	}
}

func (s ormSignal) OnDeleted(m goovn.Model) {
	switch m.Table() {
	case "Logical_Router":
		lr := m.(*LogicalRouter)
		fmt.Printf(":( A poor Logical Router got deleted...Bye bye:\n")
		fmt.Printf("%+v\n", *lr)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		fmt.Printf("Got signal %s", sig)
		done <- true
	}()

	var ovs_rundir = os.Getenv("OVS_RUNDIR")
	if ovs_rundir == "" {
		log.Fatalf("specify OVS_RUNDIR")
	}
	config := goovn.Config{
		Db:          goovn.DBNB,
		Addr:        "unix:" + ovs_rundir + "/" + ovnnbSocket,
		ORMSignalCB: ormSignal{},
		DBModel:     exampleModel,
	}
	orm, err := goovn.NewORMClient(&config)
	if err != nil {
		panic(err)
	}
	defer orm.Close()
	fmt.Println("Waiting for signal or new Logical Routers")
	<-done
	fmt.Println("Exiting")
}
