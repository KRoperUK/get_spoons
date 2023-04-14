# Get Spoons

## get_spoons

Produce a CSV file of all the current JD Wetherspoon free houses.

## csv2osm

Produces an Open Street Map file, that, along with the contents of the web directory (place on a webserver) can be used to produce a map of free houses that you have visited.  Edit the latest_list.txt and change no_visit to visit, you could use the description field to record the date(s) of visit.

## get_spoons.py

Made by [KRoperUK](https://github.com/KRoperUK). CLI-based approach to generate a csv of all. Use ```python get_spoons.py -h``` for usage help.

## add_visited.py

Made by [KRoperUK](https://github.com/KRoperUK). Use as ```python add_visited.py -i "filename.csv"``` to add "Visited" field to CSV with "N" as default value.

## Actions

Added by [KRoperUK](https://github.com/KRoperUK). A new spoons list will be automatically generated every 28 days using the ["Update CSV"](./.github/workflows/update.yml) workflow. It will then trigger a ["Update latest list"](./.github/workflows/latest_csv.yml) workflow to update [latest_list.csv](./latest_list.csv).
