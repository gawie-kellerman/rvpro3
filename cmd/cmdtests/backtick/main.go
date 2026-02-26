package main

import "rvpro3/radarvision.com/utils"

func main() {
	bt := `
SELECT *
  FROM abc;
`
	utils.Print.Ln(bt)
}
