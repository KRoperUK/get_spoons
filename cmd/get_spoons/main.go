package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/KRoperUK/get_spoons/jdw"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

var Version = "v0.0.0"

func main() {
	if err := Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// Run executes the CLI logic and returns any errors.
func Run(args []string) error {
	fs := flag.NewFlagSet("get_spoons", flag.ContinueOnError)
	version := fs.Bool("version", false, "Print version and exit")
	outputFile := fs.String("output", "", "Output file path (default: stdout)")
	csvOutput := fs.Bool("csv", false, "Output as CSV")
	expand := fs.Bool("expand", false, "Expand venue details (only valid with -json)")
	appVersion := fs.String("app-version", getEnv("JDW_APP_VERSION", "6.7.1"), "JDW App Version")
	token := fs.String("token", getEnv("JDW_TOKEN", "1|SFS9MMnn5deflq0BMcUTSijwSMBB4mc7NSG2rOhqb2765466"), "JDW Bearer Token")
	userAgent := fs.String("user-agent", getEnv("JDW_USER_AGENT", "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"), "User Agent")
	debug := fs.Bool("debug", false, "Enable debug logging")
	limit := fs.Int("limit", 0, "Limit number of venues (0 for all)")
	menus := fs.Bool("menus", false, "Fetch menus for each venue (implies -expand)")
	items := fs.Bool("items", false, "Fetch menu items (implies -menus)")
	concurrency := fs.Int("concurrency", 1, "Number of concurrent requests")
	venueID := fs.Int("venue", 0, "Specific venue ID to fetch")
	searchQuery := fs.String("search", "", "Search for a venue by name")
	noFuzzy := fs.Bool("no-fuzzy", false, "Disable fuzzy searching (use case-insensitive substring match)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *version {
		fmt.Fprintf(os.Stderr, "get_spoons %s\n", Version)
		return nil
	}

	client := jdw.NewClient(*appVersion, *token, *userAgent)
	if apiURL := os.Getenv("JDW_API_URL"); apiURL != "" {
		client.SetBaseURL(apiURL)
	}
	client.SetDebug(*debug)

	var venues []jdw.Venue
	var err error

	if *venueID != 0 {
		fmt.Fprintf(os.Stderr, "Fetching venue %d...\n", *venueID)
		v, err := client.GetVenue(*venueID)
		if err != nil {
			return fmt.Errorf("fetching venue %d: %w", *venueID, err)
		}
		venues = []jdw.Venue{*v}
	} else {
		fmt.Fprintln(os.Stderr, "Fetching venues from JDW API...")
		venues, err = client.GetVenues()
		if err != nil {
			return fmt.Errorf("fetching venues: %w", err)
		}
	}

	if *searchQuery != "" {
		venues = searchVenues(venues, *searchQuery, *noFuzzy)
		fmt.Fprintf(os.Stderr, "Found %d matches.\n", len(venues))
	}

	if *limit > 0 && *limit < len(venues) {
		fmt.Fprintf(os.Stderr, "Limiting output to %d venues.\n", *limit)
		venues = venues[:*limit]
	}

	if *items && len(venues) > 10 {
		fmt.Fprintf(os.Stderr, "WARNING: Fetching menu items for %d venues. Fully expanded venue data is large (approx 20MB per venue); total output could exceed %d MB.\n", len(venues), len(venues)*20)
	}
	var finalData interface{}
	finalData = venues // Default to standard venues

	if *expand || *menus || *items {
		finalData = expandVenues(client, venues, *concurrency, *menus, *items)
	}

	// Output Section
	var out io.Writer = os.Stdout
	if *outputFile != "" {
		fmt.Fprintf(os.Stderr, "Successfully fetched %d venues. Writing to %s...\n", len(venues), *outputFile)
		f, err := os.Create(*outputFile)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}
		defer f.Close()
		out = f
	}

	err = writeFormattedOutput(out, venues, finalData, *csvOutput)
	if err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	if *outputFile != "" {
		fmt.Fprintln(os.Stderr, "Done.")
	}
	return nil
}

func expandVenues(client *jdw.Client, venues []jdw.Venue, concurrency int, includeMenus, includeItems bool) []map[string]interface{} {
	fmt.Fprintf(os.Stderr, "Fetching details for %d venues...\n", len(venues))

	if concurrency < 1 {
		concurrency = 1
	}

	var (
		detailedVenues []map[string]interface{}
		wg             sync.WaitGroup
		mu             sync.Mutex
		processedCount int
		sem            = make(chan struct{}, concurrency)
	)

	reportProgress := func() {
		processedCount++
		fmt.Fprintf(os.Stderr, "\rProcessing venue %d/%d", processedCount, len(venues))
	}

	for _, v := range venues {
		wg.Add(1)
		sem <- struct{}{}

		go func(v jdw.Venue) {
			defer wg.Done()
			defer func() { <-sem }()

			details, err := client.GetVenueDetails(v.VenueRef)
			if err != nil {
				mu.Lock()
				fmt.Fprintf(os.Stderr, "\nError fetching details for venue ID %d (Ref %d): %v\n", v.ID, v.VenueRef, err)
				reportProgress()
				mu.Unlock()
				return
			}

			if includeMenus || includeItems {
				if salesAreas, ok := details["salesAreas"].([]interface{}); ok && len(salesAreas) > 0 {
					if firstArea, ok := salesAreas[0].(map[string]interface{}); ok {
						if salesAreaIDFloat, ok := firstArea["id"].(float64); ok {
							salesAreaID := int(salesAreaIDFloat)
							menuData, err := client.GetMenus(v.VenueRef, salesAreaID)
							if err != nil {
								fmt.Fprintf(os.Stderr, "\nError fetching menus for venue %d: %v\n", v.VenueRef, err)
							} else {
								if includeItems {
									for mIdx, mVal := range menuData {
										if menuMap, ok := mVal.(map[string]interface{}); ok {
											if menuIDFloat, ok := menuMap["id"].(float64); ok {
												menuID := int(menuIDFloat)
												menuDetails, err := client.GetMenuItems(v.VenueRef, salesAreaID, menuID)
												if err != nil {
													fmt.Fprintf(os.Stderr, "\nError fetching items for menu %d (Venue %d): %v\n", menuID, v.VenueRef, err)
												} else {
													menuMap["details"] = menuDetails
													menuData[mIdx] = menuMap
												}
											}
										}
									}
								}
								details["menus"] = menuData
							}
						}
					}
				}
			}

			mu.Lock()
			detailedVenues = append(detailedVenues, details)
			reportProgress()
			mu.Unlock()
		}(v)
	}
	wg.Wait()
	fmt.Fprintln(os.Stderr, "\nDone fetching details.")
	return detailedVenues
}

func writeFormattedOutput(w io.Writer, venues []jdw.Venue, finalData interface{}, asCSV bool) error {
	if asCSV {
		return writeCSV(w, venues)
	}
	return writeJSON(w, finalData)
}

func writeJSON(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func writeJSONToFile(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeJSON(file, data)
}

func writeCSVToFile(filename string, venues []jdw.Venue) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeCSV(file, venues)
}

func writeCSV(w io.Writer, venues []jdw.Venue) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"Pub Name", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL"}
	if err := writer.Write(header); err != nil {
		return err
	}

	for _, v := range venues {
		street := v.Address.Line1
		if v.Address.Line2 != nil && *v.Address.Line2 != "" {
			street += ", " + *v.Address.Line2
		}
		if v.Address.Line3 != nil && *v.Address.Line3 != "" {
			street += ", " + *v.Address.Line3
		}

		record := []string{
			v.Name,
			strconv.FormatFloat(v.Address.Location.Latitude, 'f', 8, 64),
			strconv.FormatFloat(v.Address.Location.Longitude, 'f', 8, 64),
			street,
			v.Address.Town,
			v.Address.County,
			v.Address.Postcode,
			"",
			"",
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func searchVenues(venues []jdw.Venue, searchQuery string, noFuzzy bool) []jdw.Venue {
	if noFuzzy {
		fmt.Fprintf(os.Stderr, "Searching for venues matching \"%s\" (substring)...\n", searchQuery)
		query := strings.ToLower(searchQuery)
		var filtered []jdw.Venue
		for _, v := range venues {
			// Search name, address, town, county, postcode
			found := strings.Contains(strings.ToLower(v.Name), query) ||
				strings.Contains(strings.ToLower(v.Address.Line1), query) ||
				strings.Contains(strings.ToLower(v.Address.Town), query) ||
				strings.Contains(strings.ToLower(v.Address.Postcode), query) ||
				strings.Contains(strings.ToLower(v.Address.County), query)

			if found {
				filtered = append(filtered, v)
			}
		}
		return filtered
	}

	fmt.Fprintf(os.Stderr, "Searching for venues matching \"%s\" (fuzzy)...\n", searchQuery)

	type searchResult struct {
		venue jdw.Venue
		rank  int
	}
	var results []searchResult

	for _, v := range venues {
		// Check name
		bestRank := fuzzy.RankMatchFold(searchQuery, v.Name)

		// Check address fields and take the best (lowest) rank
		fields := []string{v.Address.Line1, v.Address.Town, v.Address.County, v.Address.Postcode}
		for _, f := range fields {
			r := fuzzy.RankMatchFold(searchQuery, f)
			if r >= 0 && (bestRank < 0 || r < bestRank) {
				bestRank = r
			}
		}

		if bestRank >= 0 {
			results = append(results, searchResult{v, bestRank})
		}
	}

	// Sort by rank: shorter distance first, then alphabetically
	sort.Slice(results, func(i, j int) bool {
		if results[i].rank != results[j].rank {
			return results[i].rank < results[j].rank
		}
		return results[i].venue.Name < results[j].venue.Name
	})

	var filtered []jdw.Venue
	for _, res := range results {
		filtered = append(filtered, res.venue)
	}
	return filtered
}
