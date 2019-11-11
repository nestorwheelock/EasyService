package main

import (
	"fmt"
	"os"

	"github.com/JojiiOfficial/SystemdGoService"
	"github.com/mkideal/cli"
)

type createT struct {
	cli.Helper
	Name        string `cli:"*N,name" usage:"Specify the name of the service"`
	ExecFile    string `cli:"*F,file" usage:"Specify the ExecStart file" `
	Description string `cli:"D,description" usage:"Specify the description of the service"`
	User        string `cli:"U,user" usage:"Specify the user for the service"`
	Group       string `cli:"G,group" usage:"Specify the group for the service"`
}

var createCMD = &cli.Command{
	Name:    "create",
	Aliases: []string{"create"},
	Argv:    func() interface{} { return new(createT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*createT)
		description := "An easy service for " + argv.Name
		if len(argv.Name) == 0 || len(argv.ExecFile) == 0 {
			fmt.Println("Missing parameter value")
			return nil
		}
		if _, er := os.Stat(argv.ExecFile); er != nil {
			fmt.Println("File not found")
			return nil
		}
		if len(argv.Description) > 0 {
			description = argv.Description
		}
		if SystemdGoService.SystemfileExists(argv.Name) {
			fmt.Println("File already exists")
			return nil
		}
		service := SystemdGoService.NewDefaultService(argv.Name, description, argv.ExecFile)
		err := service.Create()
		if err != nil {
			fmt.Println("Error creating servic: " + err.Error())
		}
		return nil
	},
}