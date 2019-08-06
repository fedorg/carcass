from pprint import pprint
import sharepy
import json

HOST = "sharepoint.com"
SITE = "CTD"
LIST = "RIT_CTD"
s = sharepy.connect(f"https://{HOST}", "username", "password")

typefilter = """<Eq><FieldRef Name="FSObjType" /><Value Type="Integer">1</Value></Eq>"""

vf = """<ViewFields>
  <FieldRef Name="ID"></FieldRef>
  <FieldRef Name='FileLeafRef' />
</ViewFields>"""
# vf = ""

rl = """<RowLimit>100</RowLimit>"""
rl = ""

responses = []

idshift = 500
lastid = 0

while(True):
    camlq = f'''<View Scope="RecursiveAll"><Query>
        <OrderBy><FieldRef Name="ID"/></OrderBy>
        <Where>
            <And>
                <Geq><FieldRef Name="ID"/><Value Type="Number">{lastid}</Value></Geq>
                <Leq><FieldRef Name="ID"/><Value Type="Number">{lastid + idshift}</Value></Leq>
            </And>
        </Where>
    </Query>{rl}{vf}</View>'''

    postdata = {
        'query': {
            '__metadata': {'type': 'SP.CamlQuery'},
            'ViewXml': camlq
        }
    }

    resp = s.post(
        f"https://{HOST}/sites/{SITE}/_api/web/lists/getbytitle('{LIST}')/GetItems", json=postdata)

    resp_js = resp.json()
    try:
        objs = resp_js['d']['results']
        if (len(objs) == 0):
            print("Finished normally: " + json.dumps(resp_js))
            break
        print([lastid, lastid+idshift, objs[-1]['ID']])
        responses.extend(objs)
        lastid = lastid + idshift # + 1
    except Exception as e:
        import sys
        print(e, file=sys.stderr)
        break

with open('spout2.json', 'w') as wf:
    wf.write(json.dumps(responses))


