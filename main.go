package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/playwright-community/playwright-go"
)

func main() {
	for {
		poll()
		time.Sleep(2 * time.Minute)
	}
}

// var freeDaysRegex = regexp.MustCompile(`\d+ frei`)

func poll() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	page, err := browser.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto("https://egov.potsdam.de/tnv/?START_OFFICE=buergerservice"); err != nil {
		log.Fatalf("could not goto: %v", err)
	}
	if err := page.Locator("#action_officeselect_termnew_prefix1333626470").Click(); err != nil {
		log.Fatalf("could not click entry button: %v", err)
	}

	t := true
	selectedValues, err := page.GetByLabel("Beantragung eines Reisepasses", playwright.PageGetByLabelOptions{Exact: &t}).SelectOption(playwright.SelectOptionValues{Values: &[]string{"1"}})
	if err != nil {
		log.Fatalf("failed to select the Reisepass option: %v", err)
	} else if len(selectedValues) != 1 || selectedValues[0] != "1" {
		log.Fatal("selected something other than the Reisepass option")
	}

	if err := page.Locator("#action_concernselect_next").Click(); err != nil {
		log.Fatalf("could not click continue button after selecting reisepass: %v", err)
	}

	if err := page.Locator("#action_concerncomments_next").Click(); err != nil {
		log.Fatalf("could not click continue button after selecting reisepass: %v", err)
	}

	time.Sleep(3 * time.Second)
	cells, err := page.Locator("td.ekolCalendar_CellInRange").All()
	if err != nil {
		log.Fatalf("error during locating cells: %v", err)
	} else if len(cells) < 10 {
		log.Fatalf("failed to parse cells (suspiciously low number of cells: %d)", len(cells))
	}

	foundSomething := false
	for _, cell := range cells {
		weekday, err := cell.Locator("b.ui-table-cell-label").InnerText()
		if err != nil {
			log.Fatalf("failed to parse weekday: %v", err)
		}
		dayInMonth, err := cell.Locator("div.ekolCalendar_DayNumberInRange").InnerText()
		if err != nil {
			log.Fatalf("Failed to parse day in month: %v", err)
		}

		monthNumberWithSurroundingDots, err := cell.Locator("span.conMonthNr").InnerText()
		if err != nil {
			log.Fatalf("Failed to parse day in month: %v", err)
		}

		freeDaysText, err := cell.Locator("div.ekolCalendar_FreeTimeContainer").InnerText()
		if err != nil {
			log.Fatalf("failed to parse freeDaysText: %v", err)
		}
		if freeDaysText == "geschlossen" {
			continue
		}
		countString := strings.Split(freeDaysText, " ")[0]
		count, err := strconv.Atoi(countString)
		if err != nil {
			log.Fatalf("failed to convert string to number: %v", err)
		}
		if count == 0 {
			continue
		}

		foundSomething = true
		message := fmt.Sprintf("Am %s, den %s%s ist ein Termin frei! Schnapp ihn dir: https://egov.potsdam.de/tnv/?START_OFFICE=buergerservice", weekday, dayInMonth, monthNumberWithSurroundingDots)
		if count > 1 {
			message = fmt.Sprintf("Am %s, den %s%s sind %d Termine frei! Schnapp sie dir: https://egov.potsdam.de/tnv/?START_OFFICE=buergerservice", weekday, dayInMonth, monthNumberWithSurroundingDots, count)
		}
		sendTelegramMessage(message)
	}

	if !foundSomething {
		log.Print("no appointments found")
	} else {
		log.Print("success! Found 1 or more appointments")
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
