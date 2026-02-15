package ui

import "fmt"

const bannerArt = `
 ██╗   ██╗ ██████╗ ██╗██████╗ ██╗   ██╗██████╗ ███╗   ██╗
 ██║   ██║██╔═══██╗██║██╔══██╗██║   ██║██╔══██╗████╗  ██║
 ██║   ██║██║   ██║██║██║  ██║██║   ██║██████╔╝██╔██╗ ██║
 ╚██╗ ██╔╝██║   ██║██║██║  ██║╚██╗ ██╔╝██╔═══╝ ██║╚██╗██║
  ╚████╔╝ ╚██████╔╝██║██████╔╝ ╚████╔╝ ██║     ██║ ╚████║
   ╚═══╝   ╚═════╝ ╚═╝╚═════╝   ╚═══╝  ╚═╝     ╚═╝  ╚═══╝`

func Banner() string {
	colored := PurpleStyle.Render(bannerArt)
	tagline := AccentStyle.Render("  Secure. Private. Void.")
	return fmt.Sprintf("%s\n%s\n", colored, tagline)
}
