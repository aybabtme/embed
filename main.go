package main

import (
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	app := newApp()
	log.SetFlags(0)
	log.SetPrefix(app.Name + ": ")
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Author = "Antoine Grondin"
	app.Email = "antoinegrondin@gmail.com"
	app.Name = "embed"
	app.Usage = "embeds the content of files in Go strings or bytes"
	app.Commands = []cli.Command{
		embedUniqueFile(),
	}

	return app
}

func fatalf(c *cli.Context, format string, args ...interface{}) {
	log.Printf(format, args...)
	cli.ShowCommandHelp(c, c.Command.Name)
	os.Exit(1)
}

func mustString(c *cli.Context, flag cli.StringFlag) string {
	value := c.String(flag.Name)
	if value == "" && flag.Value == "" {
		fatalf(c, "flag %q is mandatory", flag.Name)
	} else if value == "" && flag.Value != "" {
		return flag.Value
	}
	return value
}

func mustOpenFile(c *cli.Context, flag cli.StringFlag) *os.File {
	filename := mustString(c, flag)
	file, err := os.Open(filename)
	switch err {
	case nil:
		return file
	case os.ErrNotExist:
		fatalf(c, "file %q does not exist", filename)
	default:
		fatalf(c, "can't open file %q, %v", filename, err)
	}
	panic("unreachable")
}

func embedUniqueFile() cli.Command {

	varNameFlag := cli.StringFlag{
		Name:  "var",
		Usage: "sets this variable to the content of the file",
	}

	dirFlag := cli.StringFlag{
		Name:  "dir",
		Usage: "name of a directory in which the variable is declared",
		Value: ".",
	}

	sourceFlag := cli.StringFlag{
		Name:  "source",
		Usage: "name of a file which content's will be set in the value of the variable",
	}

	keepFlag := cli.BoolFlag{
		Name:  "keep",
		Usage: "keeps the Go source file intact, creating a new file where the variable is set",
	}

	return cli.Command{
		Name:        "file",
		ShortName:   "f",
		Usage:       "embeds a unique file",
		Description: "embeds the content of a file into a variable",
		Flags:       []cli.Flag{varNameFlag, dirFlag, sourceFlag, keepFlag},
		Action: func(c *cli.Context) {

			// this code is the second worst code
			varName := mustString(c, varNameFlag)
			dirName := mustString(c, dirFlag)
			source := c.String(sourceFlag.Name)

			var src io.Reader
			if source == "" {
				src = timeoutReader{os.Stdin, time.Second}
			} else {
				file := mustOpenFile(c, sourceFlag)
				defer file.Close()
				src = file
			}

			content, err := ioutil.ReadAll(src)
			if err != nil {
				log.Fatalf("couldn't read content, %v", err)
			}

			filename, newFileContent, err := setVariable(dirName, varName, content)
			if err != nil {
				log.Fatalf("couldn't set content of variable %q, %v", varName, err)
			}

			var dstFilename string
			if c.Bool(keepFlag.Name) {
				dstFilename = filepath.Join(filepath.Dir(filename), "generated_"+filepath.Base(filename))
			} else {
				dstFilename = filename
			}

			dst, err := os.Create(dstFilename)
			if err != nil {
				log.Fatalf("couldn't create file %q, %v", dstFilename, err)
			}
			defer dst.Close()
			_, err = dst.Write(newFileContent)
			if err != nil {
				log.Fatalf("couldn't write to file %q, %v", dstFilename, err)
			}
			log.Printf("in file %q; value of %q set", dstFilename, varName)
		},
	}
}
