package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/JojiiOfficial/SystemdGoService"
	"github.com/mkideal/cli"
)

type createT struct {
	cli.Helper
	ExecFile    string                       `cli:"F,file" usage:"Specify the ExecStart file" `
	ExecCommand string                       `cli:"C,exec" usage:"Specify the ExecStart command" `
	Name        string                       `cli:"*N,name" usage:"Specify the name of the service"`
	Description string                       `cli:"D,description" usage:"Specify the description of the service"`
	User        string                       `cli:"U,user" usage:"Specify the user for the service"`
	Group       string                       `cli:"G,group" usage:"Specify the group for the service"`
	Type        SystemdGoService.ServiceType `cli:"T,type" usage:"Specify the type of the service"`
	Start       bool                         `cli:"s,start" usage:"Starts the service after creating"`
	Enable      bool                         `cli:"e,enable" usage:"Enables the service after creating"`
	Yes         bool                         `cli:"y,yes" usage:"Skip confirm messages" dft:"false"`
	Overwrite   bool                         `cli:"o,overwrite" usage:"Overwrite an existing service" dft:"false"`
}

func (argv *createT) Validate(ctx *cli.Context) error {
	if (len(argv.ExecFile) == 0 && len(argv.ExecCommand) == 0) || (len(argv.ExecFile) > 0 && len(argv.ExecCommand) > 0) {
		return errors.New("You need to set ONE of the Exec arguments. Type \"" + binFile + " create -h\" for more information")
	}
	types := []string{
		(string)(SystemdGoService.Simple),
		(string)(SystemdGoService.Exec),
		(string)(SystemdGoService.Dbus),
		(string)(SystemdGoService.Notify),
		(string)(SystemdGoService.Forking),
		(string)(SystemdGoService.Oneshot),
	}
	if len(argv.Type) == 0 || !isInStrArr((string)(argv.Type), types) {
		return errors.New("Wrong type! Allowed types are:   simple, exec, dbus, notify, forking, oneshot")
	}
	return nil
}

var createCMD = &cli.Command{
	Name:    "create",
	Desc:    "Create a systemd service",
	Aliases: []string{"creat", "c"},
	Argv:    func() interface{} { return new(createT) },
	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*createT)
		reader := bufio.NewReader(os.Stdin)
		if os.Getgid() != 0 {
			fmt.Println("You need to be root to run this command")
			return nil
		}
		description := "An easy service for " + argv.Name

		var exec string
		if len(argv.ExecFile) > 0 {
			file := argv.ExecFile
			if !strings.HasPrefix(file, "/") {
				ex, err := os.Executable()
				if err != nil {
					log.Fatal(err)
				}
				dir := path.Dir(ex)
				if strings.HasPrefix(file, "./") {
					file = dir + "/" + file[2:]
				} else {
					file = dir + "/" + file
				}
			}
			if _, er := os.Stat(argv.ExecFile); er != nil {
				fmt.Println("File not found")
				return nil
			}
			exec = file
		} else {
			exec = argv.ExecCommand
		}
		if len(argv.Description) > 0 {
			description = argv.Description
		}
		if SystemdGoService.SystemfileExists(argv.Name) {
			if !argv.Overwrite {
				fmt.Println("Service already exists! Use -o to overwrite it")
				return nil
			}
			if !argv.Yes {
				y, i := confirmInput("Do you really want to overwrite the service \""+argv.Name+"\" [y/n]> ", reader)
				if i == -1 || !y {
					return nil
				}
			}
		}
		service := SystemdGoService.NewDefaultService(argv.Name, description, exec)
		service.Service.User = "root"
		service.Service.Type = argv.Type
		if len(argv.User) != 0 {
			service.Service.User = argv.User
		}
		if len(argv.Group) != 0 {
			service.Service.Group = argv.Group
		}

		err := service.Create()
		if err != nil {
			fmt.Println("Error creating service: " + err.Error())
		} else {
			SystemdGoService.DaemonReload()
			fmt.Println("Service created successfully: \"" + serviceFolder + SystemdGoService.NameToServiceFile(argv.Name) + "\"")
			if argv.Enable {
				err = service.Start()
				if err != nil {
					fmt.Println("Error starting service:", err.Error())
					return nil
				}
				fmt.Println("Service started successfully")
				err = service.Enable()
				if err != nil {
					fmt.Println("Error enabling service:", err.Error())
					return nil
				}
				fmt.Println("Service enabled successfully")
			}
		}
		return nil
	},
}
