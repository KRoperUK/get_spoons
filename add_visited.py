import csv, argparse

p = argparse.ArgumentParser(prog="Add Visited",
                            description="Add visited to given .csv file")
p.add_argument("-i", "--input", help="Input .csv file", dest="inputFile", metavar="INPUT_FILE")
def add_visited(csvFile):
    print("[Add Visited] Adding visited column to " + csvFile)
    try:
        with open(csvFile, "r") as f:
            reader = csv.reader(f)
            data = list(reader)
            data[0].append("Visited")
            for i in range(1, len(data)):
                data[i].append("N")
    except Exception as e:
        print("[Add Visited] Error at read file: " + str(e))
        return False
    try:
        with open(csvFile, "w") as g:  
            writer = csv.writer(g,quoting=csv.QUOTE_NONNUMERIC)
            writer.writerows(data)
        print("[Add Visited] Added visited column to " + csvFile + " successfully.")
        return True
    except Exception as e:
        print("[Add Visited] Error at write file: " + str(e))
        return False

def main(**kwargs):
    if kwargs["inputFile"]:
        add_visited(kwargs["inputFile"])
    else:
        print("[Add Visited] No input file specified")

if __name__ == "__main__":
    args = p.parse_args()
    main(**vars(args))