package browser

import (
	"errors"
	"os/exec"
	"runtime"
)

// OpenURL launches the default web browser on the operating system to open the given URL.
func OpenURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// rundll32 url.dll,FileProtocolHandler obre la URL sense obrir una finestra cmd addicional
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return errors.New("sistema operatiu no suportat per obrir el navegador automàticament")
	}

	return cmd.Start()
}
