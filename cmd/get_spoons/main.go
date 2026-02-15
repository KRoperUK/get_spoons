package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strconv"

	"get_spoons/jdw"
)

func main() {
	outputFile := flag.String("output", "latest_list.csv", "Output CSV file path")
	appVersion := flag.String("app-version", getEnv("JDW_APP_VERSION", "6.7.1"), "JDW App Version")
	token := flag.String("token", getEnv("JDW_TOKEN", ""), "JDW Bearer Token")
	userAgent := flag.String("user-agent", getEnv("JDW_USER_AGENT", "Mozilla/5.0 (iPhone; CPU iPhone OS 18_7 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148"), "User Agent")
	flag.Parse()

	if *token == "" {
		fmt.Fprintf(os.Stderr, "Error: JDW Bearer Token is required. Set JDW_TOKEN env var or use --token flag.\n")
		os.Exit(1)
	}

	client := jdw.NewClient(*appVersion, *token, *userAgent)

	fmt.Println("Fetching venues from JDW API...")
	venues, err := client.GetVenues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching venues: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully fetched %d venues. Writing to %s...\n", len(venues), *outputFile)
	err = writeCSV(*outputFile, venues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done.")
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
