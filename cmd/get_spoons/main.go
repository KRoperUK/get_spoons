package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/KRoperUK/get_spoons/jdw"
)

func main() {
	outputFile := flag.String("output", "latest_list.csv", "Output file path (CSV by default, JSON if -json flag is strictly used)")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	expand := flag.Bool("expand", false, "Expand venue details (only valid with -json)")
	appVersion := flag.String("app-version", getEnv("JDW_APP_VERSION", "6.7.1"), "JDW App Version")
	token := flag.String("token", getEnv("JDW_TOKEN", ""), "JDW Bearer Token")
	userAgent := flag.String("user-agent", getEnv("JDW_USER_AGENT", "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"), "User Agent")
	debug := flag.Bool("debug", false, "Enable debug logging")
	limit := flag.Int("limit", 0, "Limit number of venues (0 for all)")
	menus := flag.Bool("menus", false, "Fetch menus for each venue (implies -expand)")
	items := flag.Bool("items", false, "Fetch menu items (implies -menus)")
	flag.Parse()

	if *token == "" {
		fmt.Fprintf(os.Stderr, "Error: JDW Bearer Token is required. Set JDW_TOKEN env var or use --token flag.\n")
		os.Exit(1)
	}

	client := jdw.NewClient(*appVersion, *token, *userAgent)
	client.SetDebug(*debug)

	fmt.Println("Fetching venues from JDW API...")
	venues, err := client.GetVenues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching venues: %v\n", err)
		os.Exit(1)
	}

	if *limit > 0 && *limit < len(venues) {
		fmt.Printf("Limiting output to %d venues.\n", *limit)
		venues = venues[:*limit]
	}

	if *jsonOutput {
		if *expand || *menus || *items {
			fmt.Printf("Fetching details for %d venues...\n", len(venues))
			var detailedVenues []map[string]interface{}
			for i, v := range venues {
				fmt.Printf("\rProcessing venue %d/%d", i+1, len(venues))
				// Try using VenueRef instead of ID, as ID resulted in 404s
				details, err := client.GetVenueDetails(v.VenueRef)
				if err != nil {
					fmt.Fprintf(os.Stderr, "\nError fetching details for venue ID %d (Ref %d): %v\n", v.ID, v.VenueRef, err)
					continue // Or exit/handle differently
				}

				if *menus || *items {
					// Extract Sales Area ID
					if salesAreas, ok := details["salesAreas"].([]interface{}); ok && len(salesAreas) > 0 {
						if firstArea, ok := salesAreas[0].(map[string]interface{}); ok {
							if salesAreaIDFloat, ok := firstArea["id"].(float64); ok {
								salesAreaID := int(salesAreaIDFloat)
								menuData, err := client.GetMenus(v.VenueRef, salesAreaID)
								if err != nil {
									fmt.Fprintf(os.Stderr, "\nError fetching menus for venue %d: %v\n", v.VenueRef, err)
								} else {
									if *items {
										// Iterate over menus and fetch items for each
										for mIdx, mVal := range menuData {
											if menuMap, ok := mVal.(map[string]interface{}); ok {
												if menuIDFloat, ok := menuMap["id"].(float64); ok {
													menuID := int(menuIDFloat)
													menuDetails, err := client.GetMenuItems(v.VenueRef, salesAreaID, menuID)
													if err != nil {
														// Be tolerant of errors for individual menus
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
				detailedVenues = append(detailedVenues, details)
			}
			fmt.Println("\nDone fetching details.")
			err = writeJSON(*outputFile, detailedVenues)
		} else {
			err = writeJSON(*outputFile, venues)
		}
	} else {
		fmt.Printf("Successfully fetched %d venues. Writing to %s...\n", len(venues), *outputFile)
		err = writeCSV(*outputFile, venues)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done.")
}

func writeJSON(filename string, data interface{}) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func writeCSV(filename string, venues []jdw.Venue) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	writer := csv.NewWriter(file)
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
