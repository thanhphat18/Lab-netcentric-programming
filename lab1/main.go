package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Property struct {
	Name      string
	PriceVND  float64
	AreaM2    float64
	Bedrooms  int
	District  string
	Available bool
}

func main() {
	properties := []Property{
		{
			Name:      "Saigon Apartment",
			PriceVND:  2500000000,
			AreaM2:    75.5,
			Bedrooms:  2,
			District:  "District 1",
			Available: true,
		},
		{
			Name:      "Thu Duc Condo",
			PriceVND:  1900000000,
			AreaM2:    62.0,
			Bedrooms:  2,
			District:  "Thu Duc",
			Available: true,
		},
		{
			Name:      "Binh Thanh Studio",
			PriceVND:  1400000000,
			AreaM2:    45.0,
			Bedrooms:  1,
			District:  "Binh Thanh",
			Available: false,
		},
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== Property Analyzer Menu ===")
		fmt.Println("1. View all properties and show the cheapeast property by price")
		fmt.Println("2. Categorize Property")
		fmt.Println("3. Property Search")
		fmt.Println("4. District Analysis with Maps")
		fmt.Println("5. Investment Calculator (ROI)")
		fmt.Println("6. Loan Calculator")
		fmt.Println("7. Smart Recommendation System")
		fmt.Println("8. Property Portfolio Optimizer")
		fmt.Println("0. Exit")

		choice := readInt(reader, "Choose option: ")

		switch choice {
		case 1:
			viewAllProperties(properties)
		case 2:
			displayPropertyCategories(properties)
		case 3:
			searchByBudget(properties, reader)
		case 4:
			districtAnalysis(properties)
		case 5:
			investmentCalculator(properties, reader)
		case 6:
			loanCalculator(reader)
		case 7:
			smartRecommendations(properties, reader)
		case 8:
			portfolioOptimizer(properties, reader)
		case 0:
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid option!")
		}
	}
}

func readLine(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func readInt(reader *bufio.Reader, prompt string) int {
	for {
		input := readLine(reader, prompt)
		value, err := strconv.Atoi(input)
		if err == nil {
			return value
		}
		fmt.Println("Please enter a valid integer.")
	}
}

func readFloat(reader *bufio.Reader, prompt string) float64 {
	for {
		input := readLine(reader, prompt)
		value, err := strconv.ParseFloat(input, 64)
		if err == nil {
			return value
		}
		fmt.Println("Please enter a valid number.")
	}
}

// TASK 1.1
func viewAllProperties(properties []Property) {
	fmt.Println("\n=== Property Comparison ===")

	if len(properties) == 0 {
		fmt.Println("No properties available.")
		return
	}

	cheapestIndex := 0
	cheapestPPM := pricePerM2(properties[0])

	for i, p := range properties {
		ppM2 := pricePerM2(p)

		fmt.Printf(
			"Property %d: %s - %.0f VND/m² | Price: %.0f VND | Area: %.1f m² | Beds: %d | Available: %s\n",
			i+1, p.Name, ppM2, p.PriceVND, p.AreaM2, p.Bedrooms, boolToYesNo(p.Available),
		)

		if p.AreaM2 > 0 && ppM2 < cheapestPPM {
			cheapestPPM = ppM2
			cheapestIndex = i
		}
	}

	cheapest := properties[cheapestIndex]
	fmt.Printf(
		"\nCheapest per m²: %s at %.0f VND/m²\n",
		cheapest.Name, cheapestPPM,
	)
}

func pricePerM2(p Property) float64 {
	if p.AreaM2 <= 0 {
		return 0
	}
	return p.PriceVND / p.AreaM2
}

func boolToYesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// TASK 1.2

func categorizeProperty(pricePerM2 float64) string {
	if pricePerM2 > 50000000 {
		return "LUXURY"
	} else if pricePerM2 > 30000000 {
		return "PREMIUM"
	} else if pricePerM2 > 20000000 {
		return "STANDARD"
	}
	return "BUDGET"
}

func displayPropertyCategories(properties []Property) {
	fmt.Println("\n=== Property Categories ===")

	categoryCount := map[string]int{
		"LUXURY":   0,
		"PREMIUM":  0,
		"STANDARD": 0,
		"BUDGET":   0,
	}

	for _, p := range properties {
		ppM2 := pricePerM2(p)
		category := categorizeProperty(ppM2)
		categoryCount[category]++

		fmt.Printf("%s: %s \n", p.Name, category)
	}

	fmt.Println("\nCategory Summary:")
	for category, count := range categoryCount {
		fmt.Printf("%s: %d properties\n", category, count)
	}
}

// Part 2:
// Task 2.1:
func displayProperty(p Property) {
	ppM2 := pricePerM2(p)

	fmt.Printf(
		"- %s | Price: %.0f VND | %.0f VND/m² | Area: %.1f m² | Beds: %d | Available: %s\n",
		p.Name,
		p.PriceVND,
		ppM2,
		p.AreaM2,
		p.Bedrooms,
		boolToYesNo(p.Available),
	)
}
func searchByBudget(properties []Property, reader *bufio.Reader) {
	fmt.Println("\n--- Search by Budget ---")

	if len(properties) == 0 {
		fmt.Println("No properties available.")
		return
	}

	maxBudget := readFloat(reader, "Enter your max budget (VND): ")
	minBeds := readInt(reader, "Minimum bedrooms (enter 0 to skip): ")
	onlyAvailableInput := readLine(reader, "Only show available properties? (y/n): ")

	onlyAvailable := false
	if strings.ToLower(onlyAvailableInput) == "y" {
		onlyAvailable = true
	}

	fmt.Println("\nResults:")
	found := 0

	for _, p := range properties {
		if p.PriceVND > maxBudget {
			continue
		}

		if minBeds > 0 && p.Bedrooms < minBeds {
			continue
		}

		if onlyAvailable && !p.Available {
			continue
		}

		displayProperty(p)
		found++
	}

	if found == 0 {
		fmt.Println("No matching properties found.")
	} else {
		fmt.Printf("Total matches: %d\n", found)
	}
}

// Task 2.2
type DistrictStats struct {
	District         string
	Count            int
	TotalPrice       float64
	AvgPrice         float64
	MostExpensive    Property
	HasMostExpensive bool
}

func districtAnalysis(properties []Property) {
	fmt.Println("\n=== District Analysis ===")

	if len(properties) == 0 {
		fmt.Println("No properties available.")
		return
	}

	statsMap := make(map[string]*DistrictStats)

	for _, p := range properties {
		d := p.District
		if strings.TrimSpace(d) == "" {
			d = "Unknown"
		}

		if _, ok := statsMap[d]; !ok {
			statsMap[d] = &DistrictStats{District: d}
		}

		s := statsMap[d]
		s.Count++
		s.TotalPrice += p.PriceVND

		if !s.HasMostExpensive || p.PriceVND > s.MostExpensive.PriceVND {
			s.MostExpensive = p
			s.HasMostExpensive = true
		}
	}

	var list []DistrictStats
	for _, s := range statsMap {
		s.AvgPrice = s.TotalPrice / float64(s.Count)
		list = append(list, *s)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].AvgPrice > list[j].AvgPrice
	})

	for _, s := range list {
		fmt.Printf("\n[%s]\n", s.District)
		fmt.Printf("Count: %d\n", s.Count)
		fmt.Printf("Average price: %.0f VND\n", s.AvgPrice)

		if s.HasMostExpensive {
			fmt.Printf("Most expensive: %s (%.0f VND)\n", s.MostExpensive.Name, s.MostExpensive.PriceVND)
		}
	}
}

// Part 3
// Task 3.1
// -------------------------
// Part 3 - Task 3.1
// -------------------------
func investmentCalculator(properties []Property, reader *bufio.Reader) {
	fmt.Println("\n=== Investment Calculator (ROI) ===")

	if len(properties) == 0 {
		fmt.Println("No properties available.")
		return
	}

	// Show list so user can pick
	for i, p := range properties {
		fmt.Printf("%d) %s | Price: %.0f VND | District: %s\n", i+1, p.Name, p.PriceVND, p.District)
	}

	choice := readInt(reader, "Choose a property (number): ")
	if choice < 1 || choice > len(properties) {
		fmt.Println("Invalid property number.")
		return
	}

	p := properties[choice-1]

	monthlyRent := readFloat(reader, "Monthly rent income (VND): ")
	appreciationRate := readFloat(reader, "Annual appreciation rate (%): ")
	annualMaintenance := readFloat(reader, "Annual maintenance cost (VND): ")

	annualRent := monthlyRent * 12
	appreciationGain := p.PriceVND * (appreciationRate / 100.0)
	netAnnualGain := annualRent + appreciationGain - annualMaintenance

	roi := 0.0
	if p.PriceVND > 0 {
		roi = (netAnnualGain / p.PriceVND) * 100.0
	}

	fmt.Println("\n--- Result ---")
	fmt.Printf("Property: %s\n", p.Name)
	fmt.Printf("Annual rent: %.0f VND\n", annualRent)
	fmt.Printf("Appreciation gain: %.0f VND\n", appreciationGain)
	fmt.Printf("Annual maintenance: %.0f VND\n", annualMaintenance)
	fmt.Printf("Net annual gain: %.0f VND\n", netAnnualGain)
	fmt.Printf("ROI: %.2f%%\n", roi)

	// Simple rating (optional but helpful)
	fmt.Print("Investment grade: ")
	if roi >= 10 {
		fmt.Println("EXCELLENT")
	} else if roi >= 6 {
		fmt.Println("GOOD")
	} else if roi >= 3 {
		fmt.Println("OK")
	} else {
		fmt.Println("RISKY")
	}
}

// Task 3.2
type LoanInfo struct {
	Principal     float64
	AnnualRatePct float64
	Years         int
}

func (l LoanInfo) MonthlyPayment() float64 {
	n := l.Years * 12
	if n <= 0 {
		return 0
	}

	r := (l.AnnualRatePct / 100.0) / 12.0 // monthly interest rate

	// If rate is 0%, simple division
	if r == 0 {
		return l.Principal / float64(n)
	}

	// Amortization formula:
	// M = P*r*(1+r)^n / ((1+r)^n - 1)
	pow := 1.0
	for i := 0; i < n; i++ {
		pow *= (1 + r)
	}

	return l.Principal * r * pow / (pow - 1)
}

func loanCalculator(reader *bufio.Reader) {
	fmt.Println("\n=== Loan Calculator ===")

	loan := LoanInfo{
		Principal:     readFloat(reader, "Loan amount (VND): "),
		AnnualRatePct: readFloat(reader, "Annual interest rate (%): "),
		Years:         readInt(reader, "Loan term (years): "),
	}

	monthly := loan.MonthlyPayment()
	months := loan.Years * 12
	totalPayment := monthly * float64(months)
	totalInterest := totalPayment - loan.Principal

	fmt.Println("\n--- Result ---")
	fmt.Printf("Monthly payment: %.0f VND\n", monthly)
	fmt.Printf("Total payment: %.0f VND\n", totalPayment)
	fmt.Printf("Total interest: %.0f VND\n", totalInterest)
}

// Part 4
// Task 4.1
func smartRecommendations(properties []Property, reader *bufio.Reader) {
	fmt.Println("\n=== Smart Recommendation System ===")

	if len(properties) == 0 {
		fmt.Println("No properties available.")
		return
	}

	budget := readFloat(reader, "Your budget (VND): ")
	minBeds := readInt(reader, "Minimum bedrooms (0 to skip): ")

	type Recommendation struct {
		Property Property
		Score    int
		Warning  string
	}

	var recs []Recommendation

	for _, p := range properties {
		if budget > 0 && p.PriceVND > budget {
			continue
		}
		if minBeds > 0 && p.Bedrooms < minBeds {
			continue
		}

		score := 0

		if isHotDistrict(p.District) {
			score += 3
		}

		if p.AreaM2 >= 50 && p.AreaM2 <= 100 {
			score += 2
		}

		if p.Available {
			score += 1
		}

		warning := ""
		if pricePerM2(p) > 60000000 {
			warning = "⚠ High price/m²"
		}

		recs = append(recs, Recommendation{Property: p, Score: score, Warning: warning})
	}

	if len(recs) == 0 {
		fmt.Println("No properties match your filters.")
		return
	}

	// Sort by score desc, then cheaper price/m²
	sort.Slice(recs, func(i, j int) bool {
		if recs[i].Score != recs[j].Score {
			return recs[i].Score > recs[j].Score
		}
		return pricePerM2(recs[i].Property) < pricePerM2(recs[j].Property)
	})

	fmt.Println("\nTop Recommendations:")
	limit := 3
	if len(recs) < limit {
		limit = len(recs)
	}

	for i := 0; i < limit; i++ {
		p := recs[i].Property
		fmt.Printf(
			"%d) %s | District: %s | Score: %d | %.0f VND/m² | Price: %.0f VND | Area: %.1f m² | Beds: %d | Available: %s %s\n",
			i+1,
			p.Name,
			p.District,
			recs[i].Score,
			pricePerM2(p),
			p.PriceVND,
			p.AreaM2,
			p.Bedrooms,
			boolToYesNo(p.Available),
			recs[i].Warning,
		)
	}
}

func isHotDistrict(d string) bool {
	d = strings.TrimSpace(strings.ToLower(d))
	return d == "district 1" || d == "district 2" || d == "district 7"
}

// Task 4.2
func portfolioOptimizer(properties []Property, reader *bufio.Reader) {
	fmt.Println("\n=== Property Portfolio Optimizer ===")

	if len(properties) == 0 {
		fmt.Println("No properties available.")
		return
	}

	totalBudget := readFloat(reader, "Enter total budget (VND): ")
	if totalBudget <= 0 {
		fmt.Println("Budget must be greater than 0.")
		return
	}

	monthlyRent := readFloat(reader, "Expected monthly rent per property (VND): ")
	appreciationRate := readFloat(reader, "Expected annual appreciation rate (%): ")

	type Candidate struct {
		Property Property
		ROI      float64
	}

	var candidates []Candidate

	for _, p := range properties {
		if p.PriceVND <= 0 {
			continue
		}

		annualRent := monthlyRent * 12
		appGain := p.PriceVND * (appreciationRate / 100.0)
		netGain := annualRent + appGain

		roi := (netGain / p.PriceVND) * 100.0
		candidates = append(candidates, Candidate{Property: p, ROI: roi})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].ROI != candidates[j].ROI {
			return candidates[i].ROI > candidates[j].ROI
		}
		return candidates[i].Property.PriceVND < candidates[j].Property.PriceVND
	})

	remaining := totalBudget
	var selected []Candidate
	totalSpent := 0.0

	for _, c := range candidates {
		if c.Property.PriceVND <= remaining {
			selected = append(selected, c)
			remaining -= c.Property.PriceVND
			totalSpent += c.Property.PriceVND
		}
	}

	if len(selected) == 0 {
		fmt.Println("No properties can fit within your budget.")
		return
	}

	fmt.Println("\nSelected Portfolio:")
	for i, s := range selected {
		fmt.Printf("%d) %s | Price: %.0f VND | ROI: %.2f%% | District: %s\n",
			i+1, s.Property.Name, s.Property.PriceVND, s.ROI, s.Property.District)
	}

	fmt.Printf("\nTotal spent: %.0f VND\n", totalSpent)
	fmt.Printf("Remaining budget: %.0f VND\n", remaining)
}
