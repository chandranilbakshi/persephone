package purrCommands

import "os"

func InitPurrDirectories(basePath string) error {
	dirs := []string{
		basePath + "/.purr/objects",
		basePath + "/.purr/refs/heads",
		basePath + "/.purr/logs",
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}
