from pprint import pprint
import sharepy
import json
import os
import urllib

HOST = os.environ.get('GSPR_HOST', '')
s = sharepy.connect(
    f"https://{HOST}", os.environ.get('GSPR_USER', ''), os.environ.get('GSPR_PASS', ''))


def caml(site, listname):
    API = f"https://{HOST}/sites/{site}/_api/web"
    LISTAPI = f"{API}/lists/getbytitle('{listname}')"
    typefilter = """<Eq><FieldRef Name="FSObjType" /><Value Type="Integer">1</Value></Eq>"""

    vf = """<ViewFields>
    <FieldRef Name="ID" />
    <FieldRef Name="FileLeafRef" />
    <FieldRef Name="FileRef" />
    </ViewFields>"""
    # vf = ""

    rl = """<RowLimit>100</RowLimit>"""
    rl = ""

    responses = []

    idshift = 500
    lastid = 0

    while(True):
        camlq = f'''<View Scope="RecursiveAll">
        {vf}
        <Query>
            <OrderBy><FieldRef Name="ID"/></OrderBy>
            <Where>
                <And>
                    <Geq><FieldRef Name="ID"/><Value Type="Number">{lastid}</Value></Geq>
                    <Leq><FieldRef Name="ID"/><Value Type="Number">{lastid + idshift}</Value></Leq>
                </And>
            </Where>
        </Query>{rl}</View>'''

        postdata = {
            'query': {
                '__metadata': {'type': 'SP.CamlQuery'},
                'ViewXml': camlq
            }
        }

        resp = s.post(
            f"{LISTAPI}/GetItems", json=postdata)

        resp_js = resp.json()
        try:
            objs = resp_js['d']['results']
            if (len(objs) == 0):
                print("Finished normally: " + json.dumps(resp_js))
                break
            print([lastid, lastid+idshift, objs[-1]['ID']])
            responses.extend(objs)
            lastid = lastid + idshift  # + 1
        except Exception as e:
            import sys
            print(e, file=sys.stderr)
            break
    return responses


def odata(site, listname):
    API = f"https://{HOST}/sites/{site}/_api/web"
    LISTAPI = f"{API}/lists/getbytitle('{listname}')"
    odata_select = "*,FileLeafRef,FileDirRef,ParentList,FileRef,FSObjType"
    page = 0
    pagesize = 5000

    ret = []
    prevdir = "/"
    start_url = f"{LISTAPI}/items?$top={pagesize}&$select={odata_select}"
    # url = f"{API}/GetFolderByServerRelativeUrl('{prevdir}')/Items?$select={odata_select}"
    nexturl = start_url
    while(True):
        resp = s.get(nexturl)
        resp_js = resp.json()
        try:
            err = resp_js.get('error')
            if (err):
                raise Exception(str(err['message']['value']))
            nexturl = resp_js['d'].get('__next', '')
            objs = resp_js['d']['results']
            ret.extend(objs)
            print(f"{listname}: page {page}")
            if not nexturl:
                print(f"{listname}: finished; items: {page*pagesize+len(objs)}")
                break
        except Exception as e:
            print(e)
            break
        # pprint(resp_js)
        page = page + 1
    return ret


for l in ["RIT_CTD", "BEV_CTD", "TRA_CTD"]:
    responses = odata("CTD", l)
    with open(f'{l}-spout-odata.json', 'w', encoding='utf-8') as wf:
        wf.write(json.dumps(responses))

    with open(f'{l}-spout-odata-list.txt', 'w', encoding='utf-8') as wf:
        for o in responses:
            suffix = ("/" if (o['FSObjType'] == 1) else "")
            wf.write(o['FileRef']+suffix+"\n")

    # responses = caml("CTD","")
    # with open(f'{LIST}-spout-caml.json', 'w') as wf:
    #     wf.write(json.dumps(responses))
