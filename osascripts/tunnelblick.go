package osascripts

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lordvidex/vpn-gate/utils"
)

var tunnelblickSources = []string{
	"/Library/Application Support/Tunnelblick/Shared",
	utils.ExpandHome("~/Library/Application Support/Tunnelblick/Configurations"),
	"/Applications/Tunnelblick.app/Contents/Resources/Deploy",
}

// TunnelblickFiles returns all the configuration files found in Tunnelblick paths
// These files include ovpn, conf and tblk files that are in the Library directory
// of tunnelblick
func TunnelblickFiles() []string {
	// PS: using goroutines were 25% faster than calling syscalls sequentially
	result := make([]string, 0)
	ch := make(chan string)
	go func() {
		// sender goroutine
		var wg sync.WaitGroup
		for _, source := range tunnelblickSources {
			wg.Add(1)
			// worker goroutine
			go func(source string) {
				defer wg.Done()
				files, _ := os.ReadDir(source)
				for _, file := range files {
					ch <- filepath.Join(source, file.Name())
				}
			}(source)
		}
		wg.Wait()
		close(ch)
	}()
	for file := range ch {
		result = append(result, file)
	}
	return result
}
