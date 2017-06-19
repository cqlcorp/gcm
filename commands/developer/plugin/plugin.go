package plugin

import (
	"fmt"
	"github.com/gocms-io/gcm/config"
	"github.com/gocms-io/gcm/utility"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const plugin_name = "name"
const plugin_name_short = "n"
const flag_hard = "delete"
const flag_hard_short = "d"
const flag_watch = "watch"
const flag_watch_short = "w"
const flag_binary = "binary"
const flag_binary_short = "b"
const flag_entry = "entry"
const flag_entry_short = "e"
const flag_dir_file_to_copy = "copy"
const flag_dir_file_to_copy_short = "c"

var CMD_PLUGIN = cli.Command{
	Name:      "plugin",
	Usage:     "copy plugin files from development directory into the gocms plugin directory",
	ArgsUsage: "<source> <gocms installation>",
	Action:    cmd_copy_plugin,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  plugin_name + ", " + plugin_name_short,
			Usage: "Name of the plugin. *Required",
		},
		cli.BoolFlag{
			Name:  flag_hard + ", " + flag_hard_short,
			Usage: "Delete the existing destination and replace with the contents of the source.",
		},
		cli.BoolFlag{
			Name:  flag_watch + ", " + flag_watch_short,
			Usage: "Watch for file changes in source and copy to destination on change.",
		},
		cli.StringFlag{
			Name:  flag_entry + ", " + flag_entry_short,
			Usage: "Build the plugin using the following entry point. Defaults to 'main.go'.",
		},
		cli.StringFlag{
			Name:  flag_binary + ", " + flag_binary_short,
			Usage: "Build the plugin using the following name for the output. Defaults to -n <plugin name>.",
		},
		cli.StringSliceFlag{
			Name:  flag_dir_file_to_copy + ", " + flag_dir_file_to_copy_short,
			Usage: "Directory or file to copy with plugin. Accepts multiple instances of the flag.",
		},
	},
}

func cmd_copy_plugin(c *cli.Context) error {

	entryPoint := "main.go"
	// verify there is a source and destination
	if !c.Args().Present() {
		fmt.Println("A source and destination directory must be specified.")
		return nil
	}

	// verify that a plugin name is given
	if c.String(plugin_name) == "" {
		fmt.Println("A plugin name must be specified with the --name or -n flag.")
		return nil
	}
	pluginName := c.String(plugin_name)
	binaryName := pluginName

	srcDir := c.Args().Get(0)
	destDir := c.Args().Get(1)

	if srcDir == "" || destDir == "" {
		fmt.Println("A source and destination directory must be specified.")
		return nil
	}

	if c.String(flag_binary) != "" {
		binaryName = c.String(flag_binary)
	}

	if c.String(flag_entry) != "" {
		entryPoint = c.String(flag_entry)
	}

	var filesToCopy []string
	filesToCopy = append(filesToCopy, filepath.Join(srcDir, config.PLUGIN_MANIFEST))
	filesToCopy = append(filesToCopy, filepath.Join(srcDir, config.PLUGIN_DOCS))

	// run go generate
	goGenerate := exec.Command("go", "generate", filepath.Join(srcDir, entryPoint))
	if c.GlobalBool(config.FLAG_VERBOSE) {
		goGenerate.Stdout = os.Stdout
	}
	goGenerate.Stderr = os.Stderr
	err := goGenerate.Run()
	if err != nil {
		fmt.Printf("Error running 'go generate %v': %v\n", filepath.Join(srcDir, entryPoint), err.Error())
		return nil
	}

	// build go binary
	contentPath := filepath.Join(destDir, config.CONTENT_DIR, config.PLUGINS_DIR)
	pluginPath := filepath.Join(contentPath, pluginName)
	pluginBinaryPath := filepath.Join(pluginPath, binaryName)
	goBuild := exec.Command("go", "build", "-o", pluginBinaryPath, filepath.Join(srcDir, entryPoint))
	if c.GlobalBool(config.FLAG_VERBOSE) {
		goBuild.Stdout = os.Stdout
	}
	goBuild.Stderr = os.Stderr

	err = goBuild.Run()
	if err != nil {
		fmt.Printf("Error running 'go build -o %v %v': %v\n", pluginBinaryPath, filepath.Join(srcDir, entryPoint), err.Error())
		return nil
	}
	// set permissions to run
	err = os.Chmod(pluginBinaryPath, os.FileMode(0755))
	if err != nil {
		fmt.Printf("Error setting plugin to executable: %v\n", err.Error())
	}

	if c.StringSlice(flag_dir_file_to_copy) != nil {
		filesToCopy = append(filesToCopy, c.StringSlice(flag_dir_file_to_copy)...)
	}

	// copy files to plugin
	for _, file := range filesToCopy {
		destFile := strings.Replace(file, srcDir, "", 1)
		err = utility.Copy(file, filepath.Join(pluginPath, destFile), true, c.GlobalBool(config.FLAG_VERBOSE))
		if err != nil {
			fmt.Printf("Error copying %v: %v\n", file, err.Error())
			return nil
		}
	}

	return nil
}
