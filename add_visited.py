import csv, argparse

p = argparse.ArgumentParser(prog="Add Visited",
                            description="Add visited to given .csv file")
p.add_argument("-i", "--input", help="Input .csv file", dest="inputFile", type=argparse.FileType('r'), metavar="INPUT_FILE")
p.add_argument("-v", "--visited", help="Add Visited field", action="store_true", default=False, dest="visited")
p.add_argument("-c", "--closed", help="Add Closed field", action="store_true", default=False, dest="closed")

def addField(csvFile: str, fieldname: str):
    print(f"[Add {fieldname}] Adding visited column to " + csvFile)
    try:
        with open(csvFile, "r") as f:
            reader = csv.reader(f)
            data = list(reader)
            print(f"[Add {fieldname}] Found the following fieldnames: " + str(data[0]))
            if fieldname in data[0]:
                print(f"[Add {fieldname}] {fieldname} already exists. Aborting.")
                return False
            data[0].append(fieldname)
            for i in range(1, len(data)):
                data[i].append("N")
    except Exception as e:
        print(f"[Add {fieldname}] Error at read file: " + str(e))
        return False
    try:
        with open(csvFile, "w") as g:  
            writer = csv.writer(g,quoting=csv.QUOTE_NONNUMERIC)
            writer.writerows(data)
        print(f"[Add {fieldname}] Added visited column to " + csvFile + " successfully.")
        return True
    except Exception as e:
        print(f"[Add {fieldname}] Error at write file: " + str(e))
        return False

def main(**kwargs):
    if kwargs["inputFile"]:
        if kwargs["visited"]:
            addField(kwargs["inputFile"].name, "Visited")
        if kwargs["closed"]:
            addField(kwargs["inputFile"].name, "Closed")
        if not kwargs["visited"] and not kwargs["closed"]:
            print("[Add Visited] No field specified. Use -h for help.")
    else:
        print("[Add Visited] No input file specified")

if __name__ == "__main__":
    args = p.parse_args()
    main(**vars(args))