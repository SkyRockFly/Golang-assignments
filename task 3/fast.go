package main

import (
	"bufio"
	"fmt"
	"hw3/models"
	"io"
	"os"
	"strings"

	"github.com/mailru/easyjson"
)

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	seenBrowsers := make([]string, 0, 120)
	uniqueBrowsers := 0

	reader := bufio.NewReader(file)

	users := make([]models.User, 0, 1000)

	var sb strings.Builder

	for {
		user := models.User{}
		line, err := reader.ReadSlice('\n')

		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		err = easyjson.Unmarshal(line, &user)

		if err != nil {
			panic(err)
		}
		users = append(users, user)
	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers := user.Browsers

		for _, browser := range browsers {

			if ok := strings.Contains(browser, "Android"); ok {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
						break
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
					break
				}
			}
		}

		for _, browser := range browsers {

			if ok := strings.Contains(browser, "MSIE"); ok {
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
						break
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
					break
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.Replace(user.Email, "@", " [at] ", -1)
		fmt.Fprintf(&sb, "[%d] %s <%s>\n", i, user.Name, email)
	}

	fmt.Fprintln(out, "found users:\n"+sb.String())
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
