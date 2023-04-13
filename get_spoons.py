import requests, csv
from time import sleep
from bs4 import BeautifulSoup
from validators import url as validURL

from datetime import date

defaultCSVPath = date.today().strftime("spoons-%Y-%m-%d.csv")

import argparse
p = argparse.ArgumentParser(prog='SpoonScraper',
                    description='This script scrapes the Wetherspoon website for pub data.',
                    epilog='Use --full to scrape all pubs, or -link "<link>" to scrape a specific pub.')
p.add_argument('-f','--full', help="Access every pub.",dest='allPubs', action='store_true', default=False)
p.add_argument('-l','--link', help="Pass in a specific pub URL.", dest='specificURL', metavar="URL", default="")
p.add_argument('-o', '--output', help="Output to a specific file.", dest='outputDest', metavar="OUTPUT_FILE", default=defaultCSVPath)
p.add_argument('-v', '--no-visited', help="Choose to not store visited column (Default for column is \"N\")", action="store_true",dest='ignoreVisitedCol', default=False)
p.add_argument('-d', '--delay', help="Length of delay per request. Default is 5 seconds.", metavar="DELAY_LENGTH", type=int,dest='delay', default=5)

prefixes: list = ["/pubs/","/hotels/"]

bannedLinks: list = ["https://www.jdwetherspoon.com/pubs/all-pubs",
                     "https://www.jdwetherspoon.com/pubs/all-pubs?searchterm={{ pubSearchTerm }}",]

delay: int = 1; # Delay in seconds between requests

baseURL: str = "https://www.jdwetherspoon.com"
allPubs: str = "https://www.jdwetherspoon.com/pubs/all-pubs"

def getPubs(link: str):

    response = requests.get(link)
    soupedResponse = BeautifulSoup(response.text, "html.parser")

    pubs = []

    pubsResponse = soupedResponse.find_all("a")

    for pub in pubsResponse:
        for prefix in prefixes:
            if prefix in pub["href"][0:12]:
                pubs.append(baseURL + pub["href"])

    print("[DEBUG - getPubs] Got " + str(len(pubs)) + " pubs")
    return pubs

def getPubInfo(link: str):

    if link not in bannedLinks:
        try:
            response = requests.get(link)
            soupedResponse = BeautifulSoup(response.text, "html.parser")

            pubData = {}

            pubData["pubName"] = soupedResponse.find("h1", {"class": "banner-inner__title"}).text.strip()
            pubData["Latitude"] = soupedResponse.find("div", {"id": "map"})["data-location-lat"]
            pubData["Longitude"] = soupedResponse.find("div", {"id": "map"})["data-location-long"]
            pubData["Street"] = soupedResponse.find("span", {"itemprop": "streetAddress"}).text.strip().split("\n")[0][:-1].strip()
            pubData["Locality"] = soupedResponse.find("span", {"itemprop": "addressLocality"}).text.strip()
            pubData["Region"] = soupedResponse.find("span", {"itemprop": "addressRegion"}).text.strip()
            pubData["Postcode"] = soupedResponse.find("span", {"itemprop": "postalCode"}).text.strip()
            pubData["Telephone"] = soupedResponse.find("a", {"class": "location-block__telephone"}).text.strip()
            pubData["SourceURL"] = link
            pubData["error"] = "None"

            print("[DEBUG - pubInfo - SUCCESS] Got pub info for: " + pubData["pubName"] + "")

            return pubData
        except:
            print("[DEBUG - pubInfo - ERROR] Error getting pub info for: " + link + "")
            return {"error": "Error getting pub info"}
    else:
        print("[DEBUG - pubInfo - ERROR] Banned link: " + link + "")
        return {"error": "Banned link"}

def main(**kwargs):

    if (not kwargs["allPubs"] and kwargs["specificURL"] == ""):
        print("[SpoonScrape] Error: No arguments passed. Use -h for help.")
    elif kwargs["allPubs"]:
        print("[SpoonScrape] Scraping all pubs...")
        with open(kwargs["outputDest"], "w", newline="") as csvFile:
            errors = []
            pubs = getPubs(allPubs)
            counter = 0

            if kwargs["ignoreVisitedCol"]:
                fieldnames = ["pubName", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL"]  
            else:
                fieldnames = ["pubName", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL", "Visited"]
            writer = csv.DictWriter(csvFile, fieldnames=fieldnames)
            writer.writeheader()

            for pub in pubs:
                
                counter += 1
                sleep(kwargs["delay"]) 
                pubInfo = getPubInfo(pub)
                if pubInfo["error"] == "None":
                    del pubInfo["error"]
                    pubInfo["Visited"] = "N"
                    writer.writerow(pubInfo)
                    print("[DEBUG - writing - SUCCESS] Wrote pub info for: " + pubInfo["pubName"] + " [" + str(counter+1) + "/" + str(len(pubs)) + "]")
                else:
                    errors.append(f"- {pubInfo['pubName']}: {pubInfo['error']}")
                    print("[DEBUG - writing - ERROR] Passed error: " + pubInfo["error"])
    else:
        if kwargs["specificURL"]:
            if validURL(kwargs["specificURL"]):
                print("[SpoonScrape] Scraping specific pub...")
                pubInfo = getPubInfo(kwargs["specificURL"])
                if pubInfo["error"] == "None":
                    with open(pubInfo["pubName"] + date.today().strftime("-%Y-%m-%d.csv"), "w", newline="") as csvFile:
                        if kwargs["ignoreVisitedCol"]:
                            fieldnames = ["pubName", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL"]  
                        else:
                            pubInfo["Visited"] = "N"
                            fieldnames = ["pubName", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL", "Visited"]
                        writer = csv.DictWriter(csvFile, fieldnames=fieldnames)
                        writer.writeheader()
                        del pubInfo["error"]
                        writer.writerow(pubInfo)
                        print("[DEBUG - writing - SUCCESS] Wrote pub info for: " + pubInfo["pubName"] + "")
                else:
                    print("[DEBUG - writing - ERROR] Passed error: " + pubInfo["error"])
            else:
                print("[SpoonScrape] Error: Invalid URL")
    print("[SpoonScrape] Finished with the following errors:")
    for error in errors:
        print(error)
    csvFile.close()
    return None

if __name__ == "__main__":
    args = p.parse_args()
    main(**vars(args))