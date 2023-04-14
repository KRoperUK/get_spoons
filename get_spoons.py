import requests, csv, sys, argparse
from time import sleep
from bs4 import BeautifulSoup
from validators import url as validURL

from datetime import date

# TODO - Make debug statements only show in verbose mode.
# TODO - Ensure inputted output file is a CSV file.
# TODO - Ensure inputted output file works with specific URL.
# TODO - Add repair mode.
# TODO - Add option to use prexisting CSV file as input + add to banned list.

defaultCSVPath = date.today().strftime("spoons_list_%Y%m%d.csv")
defaultDelay = 4

p = argparse.ArgumentParser(prog='SpoonScraper',
                    description='This script scrapes the Wetherspoon website for pub data.',
                    epilog='Use --full to scrape all pubs, or -link "<link>" to scrape a specific pub.')
p.add_argument('-f','--full', help="Access every pub.",dest='allPubs', action='store_true', default=False)
p.add_argument('-l','--link', help="Pass in a specific pub URL.", dest='specificURL', metavar="URL", default="")
p.add_argument('-o', '--output', help="Output to a specific file.", dest='outputDest', metavar="OUTPUT_FILE", default=defaultCSVPath)
p.add_argument('-d', '--delay', help=f"Length of delay per request. Default is {defaultDelay} seconds.", metavar="DELAY_LENGTH", type=int,dest='delay', default=defaultDelay)
p.add_argument('--no-visited', help="Choose to not store visited column (Default for column is \"N\")", action="store_true",dest='ignoreVisitedCol', default=False)
# TODO: Add verbose mode

prefixes: list = ["/pubs/"]

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

            pubData["Pub Name"] = soupedResponse.find("h1", {"class": "banner-inner__title"}).text.strip()

            pubData["Latitude"] = soupedResponse.find("div", {"id": "map"})["data-location-lat"]
            pubData["Longitude"] = soupedResponse.find("div", {"id": "map"})["data-location-long"]

            pubData["Street"] = soupedResponse.find("span", {"itemprop": "streetAddress"}).text.strip().split("\n")[0][:-1].strip()

            potentialLocality = soupedResponse.find("span", {"itemprop": "addressLocality"})
            if potentialLocality:
                pubData["Locality"] = potentialLocality.text.strip()
            else:
                pubData["Locality"] = ""

            potentionalRegion = soupedResponse.find("span", {"itemprop": "addressRegion"})
            if potentionalRegion:
                pubData["Region"] = potentionalRegion.text.strip()
            else:
                pubData["Region"] = pubData["Locality"]
            
            potentialPostcode = soupedResponse.find("span", {"itemprop": "postalCode"})
            if potentialPostcode:
                pubData["Postcode"] = potentialPostcode.text.strip()
            else:
                pubData["Postcode"] = ""

            potentialTelephone = soupedResponse.find("a", {"class": "location-block__telephone"})
            if potentialTelephone:
                pubData["Telephone"] = potentialTelephone.text.strip()
            else:
                pubData["Telephone"] = ""

            pubData["SourceURL"] = link
            pubData["error"] = "None"

            print("[DEBUG - pubInfo - SUCCESS] Got pub info for: " + pubData["Pub Name"] + "")

            return pubData
        except Exception  as e:
            print("[DEBUG - pubInfo - ERROR] Error getting pub info for: " + link + "")
            return {"error": f"Error getting pub info: {e}", "Pub Name": link}
    else:
        print("[DEBUG - pubInfo - ERROR] Banned link: " + link + "")
        return {"error": "Banned link", "Pub Name": link}

def main(**kwargs):
    errors = []

    if (not kwargs["allPubs"] and kwargs["specificURL"] == ""):
        print("[SpoonScrape] Error: No arguments passed. Use -h for help.")
        return None
    elif kwargs["allPubs"]:
        print("[SpoonScrape] Scraping all pubs...")
        with open(kwargs["outputDest"], "w", newline="") as csvFile:
            pubs = getPubs(allPubs)
            counter = 0

            if kwargs["ignoreVisitedCol"]:
                fieldnames = ["Pub Name", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL"]  
            else:
                fieldnames = ["Pub Name", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL", "Visited"]
            writer = csv.DictWriter(csvFile, fieldnames=fieldnames,quoting=csv.QUOTE_NONNUMERIC)
            writer.writeheader()

            for pub in pubs:
                
                counter += 1
                sleep(kwargs["delay"] / 2) 
                pubInfo = getPubInfo(pub)
                if pubInfo["error"] == "None":
                    del pubInfo["error"]
                    pubInfo["Visited"] = "N"
                    writer.writerow(pubInfo)
                    print("[DEBUG - writing - SUCCESS] Wrote pub info for: " + pubInfo["Pub Name"] + " [" + str(counter+1) + "/" + str(len(pubs)) + "]")
                else:
                    errors.append(f"- {pubInfo['Pub Name']} : {pubInfo['error']}")
                    print("[DEBUG - writing - ERROR] Passed error: " + pubInfo["error"])
        csvFile.close()
    else:
        if kwargs["specificURL"]:
            if validURL(kwargs["specificURL"]):
                print("[SpoonScrape] Scraping specific pub...")
                pubInfo = getPubInfo(kwargs["specificURL"])
                if pubInfo["error"] == "None":
                    with open(pubInfo["Pub Name"] + date.today().strftime("-%Y-%m-%d.csv"), "w", newline="") as csvFile:
                        if kwargs["ignoreVisitedCol"]:
                            fieldnames = ["Pub Name", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL"]  
                        else:
                            pubInfo["Visited"] = "N"
                            fieldnames = ["Pub Name", "Latitude", "Longitude", "Street", "Locality", "Region", "Postcode", "Telephone", "SourceURL", "Visited"]
                        writer = csv.DictWriter(csvFile, fieldnames=fieldnames,quoting=csv.QUOTE_NONNUMERIC)
                        writer.writeheader()
                        del pubInfo["error"]
                        writer.writerow(pubInfo)
                        print("[DEBUG - writing - SUCCESS] Wrote pub info for: " + pubInfo["Pub Name"] + "")
                    csvFile.close()
                else:
                    print("[DEBUG - writing - ERROR] Passed error: " + pubInfo["error"])
            else:
                print("[SpoonScrape] Error: Invalid URL")

    if len(errors) > 0:
        with open("errors.log", "w") as errorFile:
            for error in errors:
                errorFile.write(error + "\n")
        errorFile.close()
        print("[SpoonScrape] Finished with the following errors:")
        for error in errors:
            print(error)
    return None

if __name__ == "__main__":
    args = p.parse_args()
    main(**vars(args))
