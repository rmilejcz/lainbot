package cmd

import (
	"bytes"
	"fmt"

	"github.com/lainbot/framework"
)

func HelpCommand(ctx framework.Context) {
	print("something")
	cmds := ctx.CmdHandler.GetCmds()
	buffer := bytes.NewBufferString("Commands: \n")
	for cmdName, cmdStruct := range cmds {
		if len(cmdName) == 1 {
			continue
		}
		// log.Println(cmdName, cmdStruct)
		msg := fmt.Sprintf("\t %s%s - %s\n", ctx.Conf.Prefix, cmdName, cmdStruct.GetHelp())
		buffer.WriteString(msg)
	}
	str := buffer.String()
	ctx.Reply(str[:len(str)-2])
}
