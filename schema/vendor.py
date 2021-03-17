import urllib.request

wireshark_url = "https://gitlab.com/wireshark/wireshark/-/raw/master/manuf"

req_headers = {"User-Agent": "Python vendor.py" }
req = urllib.request.Request(wireshark_url, headers=req_headers)
response = urllib.request.urlopen(req)
body = response.read().decode("utf-8", "replace")

rules = [l.split("\t") for l in body.splitlines() if not (l.strip().startswith("#")) and l.strip() != ""]

table = {}

for r in rules:
    rule = r[0]
    name = r[1]
    if len(r) > 2:
        name = r[2]
    if len(rule) == 8:
        table[rule] = name
    elif "/28" in rule:
        table[rule[:10]] = name
        try:
            del table[rule[:8]]
        except KeyError:
            pass
    elif "/36" in rule:
        table[rule[:13]] = name
        try:
            del table[rule[:8]]
        except KeyError:
            pass
    else:
        print("Unable to parse rule:", r)

with open("vendor.sql", "w") as f:
    f.write("start transaction;\n")
    f.write("insert into vendor(prefix, name) values\n")
    for idx, (prefix, name) in enumerate(table.items()):
        n = name.replace("'", "''")
        f.write(f"\t('{prefix.lower()}', '{n}')")
        if idx == len(table) - 1:
            f.write("\n")
        else:
            f.write(",\n")
    f.write(";\nend transaction;\n")
