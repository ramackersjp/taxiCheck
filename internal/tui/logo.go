package tui

const taxiASCII = `    ______
   /  __  \
  /  /  \  \
 |  | TAXI | |
 |  |______| |
  \__________/
`

func GetLogo() string {
	return logoStyle.Render(taxiASCII)
}
